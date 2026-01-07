package main

import (
	"fmt"
	"io"
	"os"
	"strconv"

	"gopkg.in/gomail.v2"
)

type MailConfig struct {
	FromEmail string
	SMTPHost  string
	SMTPPort  int
	SMTPUser  string
	SMTPPass  string
}

func LoadMailConfig() MailConfig {
	port, _ := strconv.Atoi(os.Getenv("SMTP_PORT"))

	return MailConfig{
		FromEmail: os.Getenv("USER_EMAIL"),
		SMTPHost:  os.Getenv("SMTP_HOST"),
		SMTPPort:  port,
		SMTPUser:  os.Getenv("SMTP_USER"),
		SMTPPass:  os.Getenv("SMTP_PASS"),
	}
}
func SendEmail(query emailPayload, list []string) (bool, error) {
	fmt.Print("Processing emailing...\n")
	var cfg = LoadMailConfig()
	for _, email := range list {
		if err := sendToSingleRecipient(cfg, query, email); err != nil {
			return false, err
		}
	}
	return true, nil
}

func sendToSingleRecipient(cfg MailConfig, payload emailPayload, email string) error {

	fmt.Print("Composing email to "+ email+ "\n")
	// compose MIME message
	m := gomail.NewMessage()
	// header
	m.SetHeader("From", cfg.FromEmail)
	m.SetHeader("To", email)
	m.SetHeader("Subject", payload.title)
	m.SetBody("text/plain", payload.msg)
	// attach payload.adjunct (files)
	for _, file := range payload.adjunct {
		if file == nil {
			continue
		}
		m.Attach(
			file.name, 

			gomail.SetHeader(map[string][]string{
				"Content-Type":              {"image/png"},
				"Content-Disposition":       {`attachment; filename="` + file.name + `"`},
				"Content-Transfer-Encoding": {"base64"},
			}),

			gomail.SetCopyFunc(func(w io.Writer) error {
				_, err := w.Write(file.data)
				return err
			}),
		)
	}
	d := gomail.NewDialer(
		cfg.SMTPHost,
		cfg.SMTPPort,
		cfg.SMTPUser,
		cfg.SMTPPass,
	)

	// send via SMTP
	fmt.Print("Email Dispatched...\n")
	return d.DialAndSend(m)
}
