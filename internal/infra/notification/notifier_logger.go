package notification

import (
	"context"
	"users/internal/domain"
	log "users/pkg/logger"
)

type loggerNotifier struct {
	l log.Interface
}

// NewLoggerNotifierService implements NotificationService using a logger
func NewLoggerNotifierService(logger log.Interface) domain.NotificationService {
	return &loggerNotifier{logger}
}

// Publish uses the configured logger Interface to publish the notification.
func (n *loggerNotifier) Publish(_ context.Context, event *domain.Event) error {
	n.l.Info("Published: Type: %s | Payload: %s", event.Type, string(event.Payload))
	return nil
}
