package basicemail

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/smtp"
	"strings"
)

func SendEmail(smtpServer string, to []string, from string, subject string, body string) {

	header := make(map[string]string)
	header["From"] = from
	header["To"] = strings.Join(to, ",")
	header["Subject"] = subject
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = "text/plain; charset=\"utf-8\""
	header["Content-Transfer-Encoding"] = "base64"

	message := ""
	for k, v := range header {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + base64.StdEncoding.EncodeToString([]byte(body))

	err := smtp.SendMail(smtpServer, nil, from, to, []byte(message))
	if err != nil {
		log.Fatal(err)
	}

}
