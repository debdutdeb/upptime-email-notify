package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/smtp"
	"os"
	"strings"
)

func getEnv(vari string) (val string) {
	var e bool
	if val, e = os.LookupEnv(vari); !e {
		log.Fatal("Environment variable " + vari + " is not set.")
	}
	return
}

func sendMails(subject, body string) error {

	var (
		user,
		password,
		host,
		port,
		from string
		to []string
	)

	if _, e := os.LookupEnv("NOTIFICATION_EMAIL_SENDGRID"); e {
		user = "apikey"
		password = getEnv("NOTIFICATION_EMAIL_SENDGRID_API_KEY")
		host = "smtp.sendgrid.net"
		port = "587"
	} else {
		user = getEnv("NOTIFICATION_EMAIL_SMTP_USERNAME")
		password = getEnv("NOTIFICATION_EMAIL_SMTP_PASSWORD")
		host = getEnv("NOTIFICATION_EMAIL_SMTP_HOST")
		port = getEnv("NOTIFICATION_EMAIL_SMTP_PORT")
	}

	from = getEnv("NOTIFICATION_EMAIL_FROM")
	to = strings.Fields(getEnv("NOTIFICATION_EMAIL_TO"))

	auth := smtp.PlainAuth("", user, password, host)
	msg := []byte("From: " + from + "\r\n" +
		"To: " + to[0] + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n" +
		"\r\n" +
		body + "\r\n")

	log.Debug("NOTIFICATION_EMAIL_SMTP_USERNAME: " + user + "\n" +
		"NOTIFICATION_EMAIL_SMTP_PASSWORD: " + password + "\n" +
		"NOTIFICATION_EMAIL_SMTP_HOST: " + host + "\n" +
		"NOTIFICATION_EMAIL_SMTP_PORT: " + port + "\n" +
		"NOTIFICATION_EMAIL_FROM: " + from + "\n" +
		"NOTIFICATION_EMAIL_TO: " + strings.Join(to, " "))

	if err := smtp.SendMail(
		host+":"+port, auth, from, to, msg); err != nil {
		return fmt.Errorf("sendMails failed. %v", err)
	}

	return nil
}
