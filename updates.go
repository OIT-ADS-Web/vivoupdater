package vivoupdater

import (
	"context"
	"log"
	"time"
)

type UpdateSubscriber interface {
	MaxConnectAttempts() int
	RetryInterval() int
}

type Indexer interface {
	Name() string
}

type Triple struct {
	Subject   string
	Predicate string
	Object    string
}

type UpdateMessage struct {
	Type   string
	Phase  string
	Name   string
	Triple Triple
	// TODO: add a retry, attempts (int) type of flag?
}

type UriBatcher struct {
	BatchSize    int
	BatchTimeout time.Duration
}

func (ub UriBatcher) Batch(ctx context.Context, logger *log.Logger,
	updates chan UpdateMessage, quit chan bool) chan map[string]bool {
	batches := make(chan map[string]bool)

	go func() {
		batch := make(map[string]bool, ub.BatchSize)

		for {
			timer := time.NewTimer(ub.BatchTimeout)
			select {
			case u := <-updates:
				timer.Stop()
				batch[u.Triple.Subject] = true
				if len(batch) == ub.BatchSize {
					batches <- batch
					batch = make(map[string]bool, ub.BatchSize)
				}
			case <-timer.C:
				if len(batch) > 0 {
					batches <- batch
					batch = make(map[string]bool, ub.BatchSize)
				}
			case <-quit:
				logger.Println("vivoupdater should quit...")
				break
			}

		}
	}()
	return batches
}
