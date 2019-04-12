package vivoupdater

import (
	"context"
	"log"
)

type BatchIndexer interface {
	Name() string
	Index(ctx context.Context, b map[string]bool, logger *log.Logger) (map[string]bool, error)
}

func IndexBatch(ctx context.Context, idx BatchIndexer, b map[string]bool, logger *log.Logger) {
	ib, err := idx.Index(ctx, b, logger)

	if err != nil {
		// notification needs to happen here - or something
		logger.Printf("Indexing Error: %v\n", err)
	}
	logger.Printf("%v uris indexed by %s", len(ib), idx.Name())
	logger.Printf("%v", ib)
}
