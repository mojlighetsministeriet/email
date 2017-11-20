package main // import "github.com/mojlighetsministeriet/email"

import (
	"crypto/tls"
	"net/http"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/log"
	"github.com/mojlighetsministeriet/utils"
	"github.com/mojlighetsministeriet/utils/jsonvalidator"
	gomail "gopkg.in/gomail.v2"
)

// TODO: Add JWT authorization to reatrict sending email to specific internal micro services only

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
	To      string `json:"to" validate:"required,email"`
	Subject string `json:"subject" validate:"required"`
	Body    string `json:"body" validate:"required"`
}

var sender SMTPSender

func sendEmail(context echo.Context) (err error) {
	request := sendEmailRequest{}

	if err = context.Bind(&request); err != nil {
		return jsonvalidator.NewMalformedJSONResponse(context)
	}

	if err = context.Validate(request); err != nil {
		return
	}

	if err = sender.Send(request.To, request.Subject, request.Body); err != nil {
		context.Logger().Error(err)
		return context.JSONBlob(http.StatusInternalServerError, []byte(`{ "message": "Failed to send email" }`))
	}

	return context.JSONBlob(http.StatusCreated, []byte(`{ "message": "Email was sent" }`))
}

func main() {
	tlsConfig, tlsError := utils.GetCACertificatesTLSConfig()
	if tlsError != nil {
		panic(tlsError)
	}

	sender = SMTPSender{
		Host:      utils.GetEnv("SMTP_HOST", ""),
		Port:      utils.GetEnvInt("SMTP_PORT", 0),
		Email:     utils.GetEnv("SMTP_EMAIL", ""),
		Password:  utils.GetFileAsString("/run/secrets/smtp-password", ""),
		TLSConfig: tlsConfig,
	}

	tlsConfig.ServerName = sender.Host

	service := echo.New()
	service.Validator = jsonvalidator.NewValidator()
	service.Use(middleware.Gzip())
	service.Logger.SetLevel(log.INFO)

	service.POST("/", sendEmail)

	service.Logger.Fatal(service.Start(":" + utils.GetEnv("PORT", "80")))
}
