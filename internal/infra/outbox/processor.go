package outbox

import (
	"context"
	"fmt"
	"sync"
	"time"
	"users/internal/domain"
	"users/pkg/logger"
)

const _maxConcurrentProcess = 5

type OutboxProcessor interface {
	//ProcessOnce fetches unsent events from the repository up to the defined limit
	ProcessOnce(ctx context.Context, limit int32) error
	// StartScheduleProcess starts a ticker that will trigger the ProcessOnce function at each interval with the defined limit.
	// The _maxConcurrentProcess constant limits the ammount of go routines that run the processOnce function. This is meant to avoid overloading the notification service.
	// In future versions this value will be defined in the config
	StartScheduleProcess(ctx context.Context, interval time.Duration, limit int32)
	// GracefulStop stops gracefully the scheduler
	GracefulStop()
}

type outboxProcessor struct {
	l         logger.Interface
	repo      domain.OutboxRepoCommands
	txhandler domain.Transaction
	notifier  domain.NotificationService
	stopChan  chan struct{}
	tasksChan chan struct{}
	wg        sync.WaitGroup
	once      sync.Once
}

func NewProcessor(logger logger.Interface, txHandler domain.Transaction, repo domain.OutboxRepoCommands, notifier domain.NotificationService) OutboxProcessor {
	return &outboxProcessor{
		l:         logger,
		repo:      repo,
		txhandler: txHandler,
		notifier:  notifier,
		stopChan:  make(chan struct{}),
		tasksChan: make(chan struct{}, _maxConcurrentProcess),
	}
}

func (p *outboxProcessor) ProcessOnce(ctx context.Context, limit int32) error {
	err := p.txhandler.BeginTx(ctx, func(txCtx context.Context) error {
		outboxEvents, err := p.repo.GetUnprocessed(txCtx, limit)
		if err != nil {
			return fmt.Errorf("failed to get events: %w", err)
		}

		for _, evt := range outboxEvents {
			if err := p.notifier.Publish(txCtx, evt); err != nil {
				return fmt.Errorf("failed to publish event: %w", err)
			}

			if err := p.repo.MarkAsProcessed(txCtx, evt.ID.String()); err != nil {
				return fmt.Errorf("failed to mark as processed: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		p.l.Error("Outbox-Process failed: %v", err)
		return err
	}

	p.l.Debug("Outbox-Process completed successfully")
	return nil
}

func (p *outboxProcessor) StartScheduleProcess(ctx context.Context, interval time.Duration, limit int32) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			// this will act as a slot reserve
			p.tasksChan <- struct{}{}
			p.wg.Add(1)
			go func() {
				// this will act as a slot release
				defer func() {
					<-p.tasksChan
					p.wg.Done()
				}()
				if err := p.ProcessOnce(ctx, limit); err != nil {
					p.l.Error("Outbox-Scheduler: %v", err)
				}
			}()
		case <-p.stopChan:
			p.l.Info("Outbox-Scheduler: Stop processing")
			ticker.Stop()
			p.wg.Wait()
			return
		case <-ctx.Done():
			p.l.Info("Outbox-Scheduler: Stop processing (context canceled)")
			ticker.Stop()
			p.wg.Wait()
			return
		}
	}
}

func (p *outboxProcessor) GracefulStop() {
	p.once.Do(func() { close(p.stopChan) })
	p.wg.Wait()
}
