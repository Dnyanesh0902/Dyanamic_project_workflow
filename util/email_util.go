package util

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"gopkg.in/gomail.v2"
)

// EmailConfig holds the email configuration
type EmailConfig struct {
	SMTPServer   string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	FromAddress  string
	FromName     string
}

type EmailRequest struct {
	From       string
	To         []string
	Cc         []string
	Subject    string
	Template   string
	Attachment []string
	Body       interface{}
}

// LoadEmailConfig loads email configuration from environment variables or a file
func LoadEmailConfig() (EmailConfig, error) {
	err := godotenv.Load() // Load the .env file
	if err != nil {
		return EmailConfig{}, fmt.Errorf("error loading .env file: %v", err)
	}

	return EmailConfig{
		SMTPServer:   os.Getenv("SMTP_SERVER"),
		SMTPPort:     parseEnvInt("SMTP_PORT"),
		SMTPUsername: os.Getenv("SMTP_USERNAME"),
		SMTPPassword: os.Getenv("SMTP_PASSWORD"),
		FromAddress:  os.Getenv("SMTP_FROM_ADDRESS"),
		FromName:     os.Getenv("SMTP_FROM_NAME"),
	}, nil
}

// SendHTMLEmail sends an HTML email
func SendTemplateEmail(emailRequest EmailRequest) (bool, error) {
	config, err := LoadEmailConfig()
	if err != nil {
		return false, fmt.Errorf("error loading email configuration: %v", err)
	}

	// Create a new message
	m := gomail.NewMessage()
	from := fmt.Sprintf("%s <%s>", config.FromName, config.FromAddress)
	m.SetHeader("From", from)
	m.SetHeader("To", emailRequest.To...)
	m.SetHeader("Cc", emailRequest.Cc...)
	m.SetHeader("Subject", emailRequest.Subject)

	// Parse the HTML template
	t, err := template.ParseFiles(emailRequest.Template)
	if err != nil {
		return false, fmt.Errorf("error parsing template: %v", err)
	}

	// Create a buffer to execute the template
	var bodyBuffer = new(strings.Builder)
	if err := t.Execute(bodyBuffer, emailRequest.Body); err != nil {
		return false, fmt.Errorf("error executing template: %v", err)
	}

	// Set the HTML body of the email
	m.SetBody("text/html", bodyBuffer.String())

	if len(emailRequest.Attachment) > 0 {
		for _, value := range emailRequest.Attachment {
			m.Attach(value)
		}
	}

	// Set up the SMTP dialer
	d := gomail.NewDialer(config.SMTPServer, config.SMTPPort, config.SMTPUsername, config.SMTPPassword)

	// Send the email
	if err := d.DialAndSend(m); err != nil {
		//maintain logs here...
		return false, fmt.Errorf("error sending email: %v", err)
	}

	return true, nil
}

func SendFailureEmail(emailRequest EmailRequest) (bool, error) {
	config, err := LoadEmailConfig()
	if err != nil {
		return false, fmt.Errorf("error loading email configuration: %v", err)
	}

	// Create a new message
	m := gomail.NewMessage()
	m.SetHeader("From", config.FromAddress)
	m.SetHeader("To", emailRequest.To...)
	m.SetHeader("Cc", emailRequest.Cc...)
	m.SetHeader("Subject", emailRequest.Subject)

	// Construct the HTML body dynamically
	var bodyBuffer bytes.Buffer
	tpl := template.Must(template.New("emailTemplate").Parse(emailRequest.Template))
	if err := tpl.Execute(&bodyBuffer, emailRequest.Body); err != nil {
		return false, fmt.Errorf("error executing template: %v", err)
	}
	// Set the HTML body of the email
	m.SetBody("text/html", bodyBuffer.String())

	if len(emailRequest.Attachment) > 0 {
		for _, value := range emailRequest.Attachment {
			m.Attach(value)
		}
	}

	// Set up the SMTP dialer
	d := gomail.NewDialer(config.SMTPServer, config.SMTPPort, config.SMTPUsername, config.SMTPPassword)

	// Send the email
	if err := d.DialAndSend(m); err != nil {
		//maintain logs here...
		return false, fmt.Errorf("error sending email: %v", err)
	}

	return true, nil
}

// parseEnvInt parses an environment variable as an integer
func parseEnvInt(key string) int {
	valStr := os.Getenv(key)
	if valStr == "" {
		return 0
	}

	val, err := strconv.Atoi(valStr)
	if err != nil {
		return 0
	}

	return val
}
