package vivoupdater

import (
	"context"
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
	// add a retry type of flag?
	Attempts int `json:",omitempty"`
}

type UriBatcher struct {
	BatchSize    int
	BatchTimeout time.Duration
}

func (ub UriBatcher) Batch(ctx context.Context, updates chan UpdateMessage) chan map[string]bool {
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
				notifier := GetNotifier()
				notifier.DoSend("vivoupdater batcher context cancelled", err)
				// TODO: panic here seems to far from main()
				// but otherwise never breaks out of loop
				//panic(err)
			}
		}
	}()
	return batches
}
