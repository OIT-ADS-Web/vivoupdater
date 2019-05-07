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

func (ub UriBatcher) Batch(ctx context.Context, updates chan UpdateMessage, logger *log.Logger) chan map[string]bool {
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
			case <-ctx.Done():
				err := ctx.Err()
				logger.Printf("vivoupdater batcher context cancelled:%v\n", err)
				// TODO: should it exit here?
				// os.Exit(1)
			}
		}
	}()
	return batches
}
