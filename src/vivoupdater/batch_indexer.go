package vivoupdater

import ("log")

type BatchIndexer interface {
	Name() string
	Index(logger *log.Logger, b map[string]bool) (map[string]bool, error)
}

func IndexBatch(ctx Context, i BatchIndexer, b map[string]bool) {
	ib, err := i.Index(ctx.Logger, b)
	if err != nil {
		ctx.handleError("Indexing Error", err, true)
	}
	ctx.Logger.Printf("%v uris indexed by %s", len(ib), i.Name())
	ctx.Logger.Printf("%v", ib)
}
