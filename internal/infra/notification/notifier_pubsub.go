package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"users/internal/domain"
	log "users/pkg/logger"
	pubsub "users/pkg/pubsub"
)

var attributes = map[string]string{
	"origin": "users-service",
	"source": "pubsub-notifier",
}

type pubsubEvent struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type topics struct {
	usersTopic pubsub.Topic
}

type gcpPubSubNotifier struct {
	l      log.Interface
	client pubsub.Interface
	topics *topics
}

// NewPubSubNotifierService implements NotificationService using gcp pubsub
func NewPubSubNotifierService(pubsubClient pubsub.Interface, logger log.Interface, usersTopic string) domain.NotificationService {
	return &gcpPubSubNotifier{logger, pubsubClient, &topics{
		usersTopic: pubsubClient.Topic(usersTopic),
	}}
}

func (n *gcpPubSubNotifier) getTopic(event_type string) (pubsub.Topic, error) {
	switch event_type {
	case "CreateUser", "UpdateUser", "DeleteUser":
		return n.topics.usersTopic, nil
	default:
		return nil, fmt.Errorf("unknown type: %s", event_type)
	}
}

// Publish uses the configured pubsub client to publish the notification.
func (n *gcpPubSubNotifier) Publish(ctx context.Context, event *domain.Event) error {
	topic, err := n.getTopic(event.Type)
	if err != nil {
		n.l.Error("PubSubNotifier Failed to getTopic: %s", err.Error())
		return domain.ErrNotificationNotSent
	}

	jsonEvent, err := json.Marshal(pubsubEvent{
		Type:    event.Type,
		Payload: json.RawMessage(event.Payload),
	})
	if err != nil {
		n.l.Error("PubSubNotifier Failed to marshall event: %s", err.Error())
		return domain.ErrFailedToProcessData
	}
	msg := &pubsub.Message{
		Data:       jsonEvent,
		Attributes: attributes,
	}
	result := topic.Publish(ctx, msg)

	// block until the result is returned and a server-generated ID is returned for the published message.
	if _, err = result.Get(ctx); err != nil {
		n.l.Debug("PubSubNotifier Failed to publish message: %s", err.Error())
		return domain.ErrNotificationNotSent
	}
	n.l.Debug("PubSubNotifier Published: %v", string(jsonEvent))
	return nil
}
