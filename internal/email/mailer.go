package email

import (
	"bytes"
	"fmt"
	"net/smtp"
)

type Mailer struct {
	user string
	pass string
}

func NewMailer(user, pass string) *Mailer {
	return &Mailer{user: user, pass: pass}
}

func (m *Mailer) Send(to, subject string, data ForwardData) error {
	var body bytes.Buffer
	if err := emailTemplate.Execute(&body, data); err != nil {
		return fmt.Errorf("render template: %w", err)
	}

	msg := buildMIME(m.user, to, subject, body.String())
	auth := smtp.PlainAuth("", m.user, m.pass, "smtp.gmail.com")
	return smtp.SendMail("smtp.gmail.com:587", auth, m.user, []string{to}, msg)
}

func buildMIME(from, to, subject, htmlBody string) []byte {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "From: %s\r\n", from)
	fmt.Fprintf(&buf, "To: %s\r\n", to)
	fmt.Fprintf(&buf, "Subject: %s\r\n", subject)
	fmt.Fprintf(&buf, "MIME-Version: 1.0\r\n")
	fmt.Fprintf(&buf, "Content-Type: text/html; charset=\"UTF-8\"\r\n")
	fmt.Fprintf(&buf, "\r\n")
	buf.WriteString(htmlBody)
	return buf.Bytes()
}
