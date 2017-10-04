package monitoring

import (
	"context"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	gmonitoring "google.golang.org/api/monitoring/v3"
)

// StackdriverAggregation collects data points and sends them
// to Stackdriver on the given interval.
//
// metricType example "custom.googleapis.com/stores/daily_sales"
// metricLabels example {"key":"value"}
// resourceType example "global"
type StackdriverAggregation struct {
	projectID      string
	client         *gmonitoring.Service
	ctx            context.Context
	metricType     string
	metricLabels   map[string]string
	resourceType   string
	resourceLabels map[string]string
	points         []*gmonitoring.Point
	mutex          *sync.Mutex
	in             chan *gmonitoring.Point
	isDebug        bool
	do             func(ptscc *gmonitoring.ProjectsTimeSeriesCreateCall) error
}

// Write will send the datapoint (eventually) to Stackdriver.
// The points are buffered by a given time period.
// Note that Write will be blocked if we receive more than a 1000
// data points before the copy of the slice used for storage finished.
func (sa *StackdriverAggregation) Write(point *gmonitoring.Point) {
	sa.in <- point
}

func (sa *StackdriverAggregation) startWriteLoop() {
	sa.do = sa.doer

	// TODO combine into one select statement?

	// Buffer a 1000 data points.
	sa.in = make(chan *gmonitoring.Point, 1000)
	go sa.writeFromChannelToSlice()

	// Send to Stackdriver every minute
	// TODO make configurable
	go sa.ticker(time.Minute*1, sa.send)
}

func (sa *StackdriverAggregation) ticker(
	duration time.Duration,
	sender func(points []*gmonitoring.Point),
) {
	ticker := time.NewTicker(duration)
	for range ticker.C {
		if len(sa.points) > 0 {
			toSend := make([]*gmonitoring.Point, len(sa.points))

			// Write will be blocked if we receive more than a 1000
			// data points before this copy has finished.
			sa.mutex.Lock()
			copy(toSend, sa.points)
			sa.points = []*gmonitoring.Point{}
			sa.mutex.Unlock()

			sender(toSend)
		}
	}
}

func (sa *StackdriverAggregation) writeFromChannelToSlice() {
	for point := range sa.in {
		sa.mutex.Lock()
		sa.points = append(sa.points, point)
		sa.mutex.Unlock()
	}
}

// Send will send the data points to Stackdriver, writing to a timeseries
func (sa *StackdriverAggregation) send(points []*gmonitoring.Point) {
	tsr := gmonitoring.CreateTimeSeriesRequest{
		TimeSeries: []*gmonitoring.TimeSeries{
			{
				Metric: &gmonitoring.Metric{
					Type:   sa.metricType,
					Labels: sa.metricLabels,
				},
				Resource: &gmonitoring.MonitoredResource{
					Type:   sa.resourceType,
					Labels: sa.resourceLabels,
				},
				Points: points,
			},
		},
	}

	if sa.isDebug {
		log.WithField("time series request", tsr).Info("Stackdriver metric not sent in debug mode")
	} else {
		ptscc := sa.client.Projects.TimeSeries.Create(sa.projectResource(), &tsr)
		err := sa.do(ptscc)
		if err != nil {
			log.WithError(err).Error("Failed to write time series data to Stackdriver")
		}
	}
}

func (sa *StackdriverAggregation) doer(ptscc *gmonitoring.ProjectsTimeSeriesCreateCall) error {
	_, err := ptscc.Do()
	return err
}

func (sa *StackdriverAggregation) projectResource() string {
	return "projects/" + sa.projectID
}
