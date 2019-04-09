package vivoupdater

import (
	"log"
)

type BatchIndexer interface {
	Name() string
	Index(b map[string]bool, logger *log.Logger) (map[string]bool, error)
}

func IndexBatch(i BatchIndexer, b map[string]bool, logger *log.Logger) {
	ib, err := i.Index(b, logger)

	if err != nil {
		// notification needs to happen here - or something
		logger.Printf("Indexing Error: %v\n", err)
	}
	logger.Printf("%v uris indexed by %s", len(ib), i.Name())
	logger.Printf("%v", ib)
}
