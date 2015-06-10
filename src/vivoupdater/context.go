package vivoupdater

import (
	"fmt"
	"github.com/OIT-ADS-Web/basicemail"
	"log"
)

type Context struct {
	Notice Notification
	Logger *log.Logger
	Quit   chan bool
}

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
