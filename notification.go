package vivoupdater

import (
	"fmt"

	"github.com/OIT-ADS-Web/basicemail"
)

var Notifier *Notification

type Notification struct {
	From string
	To   []string
	Smtp string
	Send bool
}

func GetNotifier() Notification {
	return Notification{From: NotificationFrom,
		To:   NotificationEmail,
		Smtp: NotificationSmtp}
}

func (n Notification) DoSend(desc string, err error) {
	if n.Smtp != "" {
		subject := fmt.Sprintf("[VIVO Updater Error] %s", desc)
		body := fmt.Sprintf("VIVO Updater has stopped:\n\n %s", err)
		basicemail.SendEmail(
			n.Smtp,
			n.To,
			n.From,
			subject,
			body)
	}
}
