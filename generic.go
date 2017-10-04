package monitoring

import (
	"time"

	gmonitoring "google.golang.org/api/monitoring/v3"
)

// MetricWriter supports writing of a metric data point
type MetricWriter interface {
	Write(point *gmonitoring.Point)
}

// MetricCreator supports creation of a metrics client
type MetricCreator interface {
	CreateMetric(
		metricType string,
		metricLabels map[string]string,
		resourceType string,
		resourceLabels map[string]string,
	) MetricWriter
}

// CreateDataPoint creates a float64 Google monitoring Point
func CreateDataPoint(value float64) *gmonitoring.Point {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	return &gmonitoring.Point{
		Interval: &gmonitoring.TimeInterval{
			StartTime: now,
			EndTime:   now,
		},
		Value: &gmonitoring.TypedValue{
			DoubleValue: &value,
		},
	}
}
