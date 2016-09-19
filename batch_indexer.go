package vivoupdater

import (
	"log"
)

type BatchIndexer interface {
	Name() string
	Index(b map[string]bool, logger *log.Logger) (map[string]bool, error)
}

func IndexBatch(ctx Context, i BatchIndexer, b map[string]bool) {
	ib, err := i.Index(b, ctx.Logger)
	if err != nil {
		ctx.handleError("Indexing Error", err, true)
	} else {
		ctx.Logger.Printf("%v uris indexed by %s", len(ib), i.Name())
		ctx.Logger.Printf("%v", ib)
	}
}
