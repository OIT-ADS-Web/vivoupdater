package vivoupdater

import (
	"fmt"
	"log"

	"github.com/OIT-ADS-Web/basicemail"
)

type Context struct {
	Notice Notification
	Logger *log.Logger
	Quit   chan bool
}

// TODO: not sure about this
/*
func (ctx Context) Deadline() (time.Time, bool) {
	d := time.Now().Add(50 * time.Millisecond)
	return d, false
}

func (ctx Context) Done() <-chan struct{} {
	return ctx.Quit
}
*/

func (ctx Context) handleError(desc string, err error, fatal bool) {
	ctx.Logger.Println(desc)
	ctx.Logger.Println(err)
	if fatal {
		if ctx.Notice.Smtp != "" {
			subject := fmt.Sprintf("[VIVO Updater Error] %s", desc)
			body := fmt.Sprintf("VIVO Updater has stopped:\n\n %s", err)
			basicemail.SendEmail(
				ctx.Notice.Smtp,
				ctx.Notice.To,
				ctx.Notice.From,
				subject,
				body)
		}
		ctx.Quit <- true
	}
}
