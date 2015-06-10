package vivoupdater

type UpdateMessage struct {
	Type   string
	Phase  string
	Name   string
	Triple Triple
}
