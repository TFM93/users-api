package pubsub

import (
	"context"
	"fmt"
	"time"
	log "users/pkg/logger"

	gcpPubSub "cloud.google.com/go/pubsub"
)

const (
	_defaultConnAttempts = 10
	_defaultConnTimeout  = time.Second
)

type Message = gcpPubSub.Message
type PublishResult = gcpPubSub.PublishResult

type Topic interface {
	// Publish publishes msg to the topic asynchronously.
	Publish(ctx context.Context, msg *Message) *PublishResult
}

type Interface interface {
	// Close closes the client
	// if the client is meant to be always available, theres no need to close the client
	Close()

	// Ping returns true if connection returns at least one topic without errors
	Ping(ctx context.Context) bool

	// Topic returns the client topic based on the topic id
	Topic(id string) Topic

	// IsEnabled returns true if pubsub client instance is enabled
	IsEnabled() bool
}

type pubsubClient struct {
	l            log.Interface
	client       *gcpPubSub.Client
	connAttempts int
	connTimeout  time.Duration
	enabled      bool
}

func New(enabled bool, projectID string, opts ...Option) (_ Interface, err error) {
	pubsubClient := &pubsubClient{
		connAttempts: _defaultConnAttempts,
		connTimeout:  _defaultConnTimeout,
		enabled:      enabled,
	}
	for _, opt := range opts {
		opt(pubsubClient)
	}

	if pubsubClient.l == nil {
		pubsubClient.l = log.New("")
	}

	if !pubsubClient.enabled {
		return pubsubClient, nil
	}

	pubsubClient.client, err = gcpPubSub.NewClient(context.Background(), projectID)
	if err != nil {
		return nil, fmt.Errorf("pubsub.NewClient: %w", err)
	}

	for pubsubClient.connAttempts > 0 {
		if pubsubClient.Ping(context.Background()) {
			break
		}

		pubsubClient.l.Warn("PubSub is trying to connect, attempts left: %d", pubsubClient.connAttempts)
		time.Sleep(pubsubClient.connTimeout)

		pubsubClient.connAttempts--
	}
	if pubsubClient.connAttempts == 0 {
		return nil, fmt.Errorf("pubsub.NewClient failed to connect")
	}
	return pubsubClient, nil
}

// Topic returns the client topic based on the topic id
func (p *pubsubClient) Topic(id string) Topic {
	if p.client != nil {
		return p.client.Topic(id)
	}
	return nil
}

// Close closes the client
// if the client is meant to be always available, theres no need to close the client
func (p *pubsubClient) Close() {
	if p.client != nil {
		p.client.Close()
	}
}

// IsEnabled returns true if the pubsub client instance is enabled
func (p *pubsubClient) IsEnabled() bool {
	return p.enabled
}

// Ping returns true if connection returns at least one topic without errors
func (p *pubsubClient) Ping(ctx context.Context) bool {
	var err error
	if p.client != nil {
		_, err = p.client.Topics(ctx).Next()
	}
	return err == nil
}

// Option allows to configure pubsub connection
type Option func(*pubsubClient)

// WithLogger injects the logger dependency
func WithLogger(logger log.Interface) Option {
	return func(c *pubsubClient) {
		c.l = logger
	}
}

// ConnAttempts configures the postgresql connection attempts
func ConnAttempts(attempts int) Option {
	return func(c *pubsubClient) {
		c.connAttempts = attempts
	}
}

// ConnTimeout configures the postgresql connection timeout duration
func ConnTimeout(timeout time.Duration) Option {
	return func(c *pubsubClient) {
		c.connTimeout = timeout
	}
}
