package vivoupdater

import (
	"time"
)

type UriBatcher struct {
	BatchSize    int
	BatchTimeout time.Duration
}

func (ub UriBatcher) Batch(ctx Context, updates chan UpdateMessage) chan map[string]bool {
	batches := make(chan map[string]bool)
	go func() {
		batch := make(map[string]bool, ub.BatchSize)
		for {
			timer := time.NewTimer(ub.BatchTimeout)
			select {
			case u := <-updates:
				ctx.Logger.Println("got some updates")
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

			}
		}
	}()
	return batches
}
