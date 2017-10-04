package monitoring

import (
	"context"
	"sync"

	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2/google"

	gmonitoring "google.golang.org/api/monitoring/v3"
)

// Stackdriver is used for creating metrics to write to on Stackdriver
type Stackdriver struct {
	ProjectID string
	Client    *gmonitoring.Service
	ctx       context.Context
	isDebug   bool
}

// NewStackdriver is the constructor/factory for the Stackdriver struct
func NewStackdriver(ctx context.Context, isDebug bool, projectID string) (*Stackdriver, error) {
	hc, err := google.DefaultClient(ctx, gmonitoring.MonitoringScope)
	if err != nil {
		log.WithError(err).Error("Failed to create monitoring client")
		return nil, err
	}

	client, err := gmonitoring.New(hc)
	if err != nil {
		log.WithError(err).Error("Failed to create monitoring service")
		return nil, err
	}

	return &Stackdriver{ProjectID: projectID, Client: client, isDebug: isDebug}, nil
}

// GetProjectID returns the project id
func (s Stackdriver) GetProjectID() string {
	return s.ProjectID
}

// CreateMetric creates a metric that can then be written to
func (s Stackdriver) CreateMetric(
	metricType string,
	metricLabels map[string]string,
	resourceType string,
	resourceLabels map[string]string,
) *StackdriverAggregation {

	sa := &StackdriverAggregation{
		projectID:      s.ProjectID,
		client:         s.Client,
		ctx:            s.ctx,
		metricType:     metricType,
		metricLabels:   metricLabels,
		resourceType:   resourceType,
		resourceLabels: resourceLabels,
		points:         []*gmonitoring.Point{},
		mutex:          &sync.Mutex{},
		isDebug:        s.isDebug,
	}

	sa.startWriteLoop()
	return sa
}
