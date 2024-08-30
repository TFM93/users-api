package app

import (
	"context"
	"users/internal/domain"
)

type healthCheckQueries struct {
	repo   domain.MonitoringRepoQueries
	pubsub domain.MonitoringRepoQueries
}

// NewHealthCheckQueries creates a service that satisfies the interface HealthCheckQueries
func NewHealthCheckQueries(repo domain.MonitoringRepoQueries, pubsub domain.MonitoringRepoQueries) HealthCheckQueries {
	return &healthCheckQueries{repo: repo, pubsub: pubsub}
}

// Check pings the following dependencies:
// - repository
// - pubsub (if enabled)
func (h *healthCheckQueries) Check(ctx context.Context) bool {
	return h.repo.Ping(ctx) && (!h.pubsub.IsEnabled() || h.pubsub.Ping(ctx))
}
