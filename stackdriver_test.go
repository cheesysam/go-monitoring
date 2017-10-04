package monitoring

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	gmonitoring "google.golang.org/api/monitoring/v3"
)

func TestNewStackdriver(t *testing.T) {
	ctx := context.Background()
	s, err := NewStackdriver(ctx, true, "PROJECTID")
	assert.Nil(t, err)
	assert.Equal(t, "PROJECTID", s.ProjectID)
	assert.NotNil(t, s.Client)
}

func TestGetProjectID(t *testing.T) {
	s := Stackdriver{ProjectID: "PROJECTID"}
	assert.Equal(t, "PROJECTID", s.GetProjectID())
}

func TestCreateMetric(t *testing.T) {
	s := Stackdriver{ProjectID: "PROJECTID"}
	metricLabels := make(map[string]string)
	resourceLabels := make(map[string]string)
	sa := s.CreateMetric("METRICTYPE", metricLabels, "RESOURCETYPE", resourceLabels)
	assert.Equal(t, "PROJECTID", sa.projectID)
}

func TestCreateDataPoint(t *testing.T) {
	d := CreateDataPoint(42.42)
	v := float64(42.42)
	expected := &gmonitoring.TypedValue{DoubleValue: &v}
	assert.Equal(t, expected, d.Value)
}
