package notification

import (
	"fmt"

	"github.com/google/logger"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

// EmailService handles sending emails via SendGrid
type EmailService struct {
	apiKey           string
	defaultFromEmail string
	defaultFromName  string
}

// NewEmailService creates a new email service instance
func NewEmailService(apiKey, fromEmail, fromName string) *EmailService {
	return &EmailService{
		apiKey:           apiKey,
		defaultFromEmail: fromEmail,
		defaultFromName:  fromName,
	}
}

// SendEmail sends an email using SendGrid
func (s *EmailService) SendEmail(to, toName, subject, htmlBody, textBody string) error {
	from := mail.NewEmail(s.defaultFromName, s.defaultFromEmail)
	recipient := mail.NewEmail(toName, to)
	message := mail.NewSingleEmail(from, subject, recipient, textBody, htmlBody)

	client := sendgrid.NewSendClient(s.apiKey)
	response, err := client.Send(message)
	if err != nil {
		logger.Errorf("Failed to send email to %s: %v", to, err)
		return fmt.Errorf("failed to send email: %w", err)
	}

	if response.StatusCode >= 400 {
		logger.Errorf("SendGrid error %d for %s: %s", response.StatusCode, to, response.Body)
		return fmt.Errorf("sendgrid error: %d - %s", response.StatusCode, response.Body)
	}

	logger.Infof("Email sent successfully to %s (status: %d)", to, response.StatusCode)
	return nil
}

// SendWithCustomFrom sends an email with a custom from address
func (s *EmailService) SendWithCustomFrom(fromEmail, fromName, to, toName, subject, htmlBody, textBody string) error {
	from := mail.NewEmail(fromName, fromEmail)
	recipient := mail.NewEmail(toName, to)
	message := mail.NewSingleEmail(from, subject, recipient, textBody, htmlBody)

	client := sendgrid.NewSendClient(s.apiKey)
	response, err := client.Send(message)
	if err != nil {
		logger.Errorf("Failed to send email to %s: %v", to, err)
		return fmt.Errorf("failed to send email: %w", err)
	}

	if response.StatusCode >= 400 {
		logger.Errorf("SendGrid error %d for %s: %s", response.StatusCode, to, response.Body)
		return fmt.Errorf("sendgrid error: %d - %s", response.StatusCode, response.Body)
	}

	logger.Infof("Email sent successfully to %s from %s (status: %d)", to, fromEmail, response.StatusCode)
	return nil
}
