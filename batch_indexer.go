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
		// notification needs to happen here - or something
		logger.Printf("Indexing Error: %v\n", err)
	}
	logger.Printf("%v uris indexed by %s", len(ib), idx.Name())
	logger.Printf("%v", ib)
}
