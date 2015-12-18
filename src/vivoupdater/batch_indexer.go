package vivoupdater

// could filter here with extra --- filter
type BatchIndexer interface {
	Name() string
	Index(b map[string]bool) (map[string]bool, error)
//        Filter(b map[string]bool) (map[string]bool, error)
}

func IndexBatch(ctx Context, i BatchIndexer, b map[string]bool) {
	ib, err := i.Index(b)
	if err != nil {
		ctx.handleError("Indexing Error", err, true)
	}
	ctx.Logger.Printf("%v uris indexed by %s", len(ib), i.Name())
	ctx.Logger.Printf("%v", ib)
}
