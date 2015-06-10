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
		batch := make(map[string]bool)
		for {
			select {
			case u := <-updates:
				batch[u.Triple.Subject] = true
				if len(batch) == ub.BatchSize {
					batches <- batch
					batch = make(map[string]bool)
				}
			case <-time.After(ub.BatchTimeout):
				if len(batch) > 0 {
					batches <- batch
					batch = make(map[string]bool)
				}
			}
		}
	}()
	return batches
}
