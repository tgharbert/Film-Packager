package email

import (
	"net/smtp"
	"os"
)

func SendEmail(to string, subject string, body string) error {
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"
	from := os.Getenv("EMAIL_ADDRESS")
	password := os.Getenv("EMAIL_PASSWORD")

	auth := smtp.PlainAuth("", from, password, smtpHost)

	msg := []byte("Subject: " + subject + "\r\n" + "\r\n" + body + "\r\n")

	return smtp.SendMail(smtpHost+":"+ smtpPort, auth, from, []string{to}, msg)
}