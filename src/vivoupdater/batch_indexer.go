package vivoupdater

type BatchIndexer interface {
	Index(b map[string]bool) (bool, error)
}

func IndexBatch(ctx Context, i BatchIndexer, b map[string]bool) {
	_, err := i.Index(b)
	if err != nil {
		ctx.handleError("Indexing Error", err, true)
	}
}
