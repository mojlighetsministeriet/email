package main // import "github.com/mojlighetsministeriet/email"

import (
	"crypto/tls"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/log"
	"github.com/mojlighetsministeriet/utils"
	gomail "gopkg.in/gomail.v2"
)

// SMTPSender is used to send emails
type SMTPSender struct {
	Host      string
	Port      int
	Email     string
	Password  string
	TLSConfig *tls.Config
}

// Send will send an email
func (sender *SMTPSender) Send(to string, subject string, body string) (err error) {
	if sender.Port == 0 {
		sender.Port = 587
	}

	message := gomail.NewMessage()
	message.SetHeader("From", sender.Email)
	message.SetHeader("To", to)
	message.SetHeader("Subject", subject)
	message.SetBody("text/html", body)
	dialer := gomail.NewDialer(sender.Host, sender.Port, sender.Email, sender.Password)
	dialer.TLSConfig = sender.TLSConfig
	err = dialer.DialAndSend(message)

	return
}

type sendEmailRequest struct {
	to      string
	subject string
	body    string
}

func sendEmail(conext echo.Context) (err error) {
	tlsConfig, err := utils.GetCACertificatesTLSConfig()
	if err != nil {
		// TODO: log here and return 500
		return
	}

	sender := SMTPSender{
		Host:      utils.GetEnv("SMTP_HOST", ""),
		Port:      utils.GetEnvInt("SMTP_PORT", 0),
		Email:     utils.GetEnv("SMTP_EMAIL", ""),
		Password:  utils.GetFileAsString("/run/secrets/smtp-password", ""),
		TLSConfig: tlsConfig,
	}

	sender.Send(context.Body, get("to"), subject, body)

	return
}

func main() {
	service := echo.New()
	service.Use(middleware.Gzip())
	service.Logger.SetLevel(log.INFO)

	service.POST("/", sendEmail)

	err := service.Start(":" + utils.GetEnv("PORT", "80"))
	if err != nil {
		panic(err)
	}
}
