package vivoupdater

type Notification struct {
	From string
	To   []string
	Smtp string
	Send bool
}
