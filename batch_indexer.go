package vivoupdater

import (
	"log"
)

type BatchIndexer interface {
	Name() string
	Index(b map[string]bool, logger *log.Logger) (map[string]bool, error)
}

func IndexBatch(idx BatchIndexer, b map[string]bool, logger *log.Logger) {
	ib, err := idx.Index(b, logger)

	if err != nil {
		// notification needs to happen here? - or something
		logger.Printf("Indexing Error: %v\n", err)
	}
	logger.Printf("batched uris=%v", ib)
	// NOTE: just means sent to indexer - they choose to index (or not)
	logger.Printf("%v uris sent to %s", len(ib), idx.Name())
}
