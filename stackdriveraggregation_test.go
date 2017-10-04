package monitoring

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	gmonitoring "google.golang.org/api/monitoring/v3"
)

func TestMain(m *testing.M) {
	// log.SetLevel(log.PanicLevel)
	os.Exit(m.Run())
}

func TestWrite(t *testing.T) {
	sa := StackdriverAggregation{}
	// TODO need a standalone constructor, make it private / non exported?
	sa.points = []*gmonitoring.Point{}
	sa.mutex = &sync.Mutex{}
	sa.in = make(chan *gmonitoring.Point)

	point := CreateDataPoint(42.42)

	go func() {
		received := <-sa.in
		assert.Equal(t, point, received)
	}()
	sa.Write(point)
}

func TestWriteFromChannelToSlice(t *testing.T) {
	sa := StackdriverAggregation{
		in:     make(chan *gmonitoring.Point),
		points: []*gmonitoring.Point{},
		mutex:  &sync.Mutex{},
	}
	go sa.writeFromChannelToSlice()

	point := CreateDataPoint(42.42)
	sa.in <- point

	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, 1, len(sa.points))
}

func TestTicker(t *testing.T) {
	count := 0
	sender := func(points []*gmonitoring.Point) {
		count++
	}

	sa := StackdriverAggregation{
		in:     make(chan *gmonitoring.Point),
		points: []*gmonitoring.Point{},
		mutex:  &sync.Mutex{},
	}

	sa.points = append(sa.points, CreateDataPoint(42.42))
	go sa.ticker(time.Millisecond*1, sender)
	time.Sleep(time.Millisecond * 10)

	assert.Equal(t, 0, len(sa.points))
	assert.Equal(t, 1, count)
}

func TestProjectResource(t *testing.T) {
	sa := StackdriverAggregation{projectID: "PROJECTID"}
	assert.Equal(t, "projects/PROJECTID", sa.projectResource())
}

func TestSend(t *testing.T) {
	hook := test.NewGlobal()
	sa := StackdriverAggregation{isDebug: true, projectID: "PROJECTID"}
	points := []*gmonitoring.Point{CreateDataPoint(42.42)}
	sa.send(points)
	assert.Equal(t, 1, len(hook.Entries))
	assert.Equal(t, logrus.InfoLevel, hook.LastEntry().Level)
	assert.Equal(t, "Stackdriver metric not sent in debug mode", hook.LastEntry().Message)
}

func TestDoError(t *testing.T) {
	hook := test.NewGlobal()
	s, err := NewStackdriver(context.TODO(), false, "PROJECTID")
	assert.Nil(t, err)
	sa := s.CreateMetric("", nil, "", nil)
	sa.do = func(ptscc *gmonitoring.ProjectsTimeSeriesCreateCall) error {
		return errors.New("ERROR")
	}

	points := []*gmonitoring.Point{CreateDataPoint(42.42)}
	sa.send(points)
	assert.Equal(t, 1, len(hook.Entries))
	assert.Equal(t, logrus.ErrorLevel, hook.LastEntry().Level)
	assert.Equal(
		t,
		"Failed to write time series data to Stackdriver",
		hook.LastEntry().Message)
}

func TestDoSuccess(t *testing.T) {
	s, err := NewStackdriver(context.TODO(), false, "PROJECTID")
	assert.Nil(t, err)
	sa := s.CreateMetric("", nil, "", nil)
	count := 0
	sa.do = func(ptscc *gmonitoring.ProjectsTimeSeriesCreateCall) error {
		count++
		return nil
	}

	points := []*gmonitoring.Point{CreateDataPoint(42.42)}
	sa.send(points)
	assert.Equal(t, 1, count)
}
