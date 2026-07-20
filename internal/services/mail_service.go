package services

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/mail"
	"net/smtp"
	"strings"
	"time"

	"github.com/AI2HU/gego/internal/db"
	"github.com/AI2HU/gego/internal/models"
)

var (
	ErrSMTPNotConfigured = errors.New("SMTP is not configured")
	ErrSMTPDisabled      = errors.New("SMTP is disabled")
)

// MailService sends email via configured SMTP settings.
type MailService struct {
	db db.Database
}

func NewMailService(database db.Database) *MailService {
	return &MailService{db: database}
}

func (s *MailService) GetSMTPSettings(ctx context.Context) (*models.SMTPSettingsResponse, error) {
	settings, err := s.db.GetSMTPSettings(ctx)
	if err != nil {
		return nil, err
	}
	resp := models.ToSMTPSettingsResponse(settings)
	return &resp, nil
}

func (s *MailService) UpdateSMTPSettings(ctx context.Context, req models.UpdateSMTPSettingsRequest) (*models.SMTPSettingsResponse, error) {
	if err := validateSMTPFields(req.Host, req.Port, req.FromEmail, req.Enabled); err != nil {
		return nil, err
	}

	existing, err := s.db.GetSMTPSettings(ctx)
	if err != nil {
		return nil, err
	}

	password := req.Password
	if password == "" {
		password = existing.Password
	}

	settings := &models.SMTPSettings{
		Host:      strings.TrimSpace(req.Host),
		Port:      req.Port,
		Username:  strings.TrimSpace(req.Username),
		Password:  password,
		FromEmail: strings.TrimSpace(req.FromEmail),
		FromName:  strings.TrimSpace(req.FromName),
		UseTLS:    req.UseTLS,
		Enabled:   req.Enabled,
		CreatedAt: existing.CreatedAt,
	}

	if err := s.db.UpsertSMTPSettings(ctx, settings); err != nil {
		return nil, err
	}

	resp := models.ToSMTPSettingsResponse(settings)
	return &resp, nil
}

// Send sends an email using the saved SMTP configuration.
func (s *MailService) Send(ctx context.Context, req models.SendEmailRequest) error {
	settings, err := s.db.GetSMTPSettings(ctx)
	if err != nil {
		return err
	}
	if !settings.Enabled {
		return ErrSMTPDisabled
	}
	if settings.Host == "" || settings.FromEmail == "" {
		return ErrSMTPNotConfigured
	}
	return s.sendWithSettings(ctx, settings, req.To, req.Subject, req.Body)
}

// TestSMTP validates connectivity and optionally sends a test message.
func (s *MailService) TestSMTP(ctx context.Context, req models.TestSMTPRequest) error {
	settings, err := s.resolveTestSettings(ctx, req)
	if err != nil {
		return err
	}

	to := strings.TrimSpace(req.To)
	if to == "" {
		to = settings.FromEmail
	}
	if to == "" {
		return fmt.Errorf("test recipient is required")
	}

	return s.sendWithSettings(ctx, settings, []string{to}, "Gego SMTP test", "This is a test email from Gego. Your SMTP configuration is working.")
}

func (s *MailService) resolveTestSettings(ctx context.Context, req models.TestSMTPRequest) (*models.SMTPSettings, error) {
	existing, err := s.db.GetSMTPSettings(ctx)
	if err != nil {
		return nil, err
	}

	settings := *existing
	if strings.TrimSpace(req.Host) != "" {
		settings.Host = strings.TrimSpace(req.Host)
	}
	if req.Port > 0 {
		settings.Port = req.Port
	}
	if req.Username != "" || req.Host != "" {
		settings.Username = strings.TrimSpace(req.Username)
	}
	if req.Password != "" {
		settings.Password = req.Password
	}
	if strings.TrimSpace(req.FromEmail) != "" {
		settings.FromEmail = strings.TrimSpace(req.FromEmail)
	}
	if req.FromName != "" || req.Host != "" {
		settings.FromName = strings.TrimSpace(req.FromName)
	}
	if req.UseTLS != nil {
		settings.UseTLS = *req.UseTLS
	}

	if err := validateSMTPFields(settings.Host, settings.Port, settings.FromEmail, true); err != nil {
		return nil, err
	}
	return &settings, nil
}

func validateSMTPFields(host string, port int, fromEmail string, enabled bool) error {
	host = strings.TrimSpace(host)
	fromEmail = strings.TrimSpace(fromEmail)

	if port < 1 || port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}

	if !enabled && host == "" && fromEmail == "" {
		return nil
	}

	if host == "" {
		return fmt.Errorf("host is required")
	}
	if fromEmail == "" {
		return fmt.Errorf("from email is required")
	}
	if _, err := mail.ParseAddress(fromEmail); err != nil {
		return fmt.Errorf("invalid from email: %w", err)
	}
	return nil
}

func (s *MailService) sendWithSettings(ctx context.Context, settings *models.SMTPSettings, to []string, subject, body string) error {
	if len(to) == 0 {
		return fmt.Errorf("at least one recipient is required")
	}
	for _, addr := range to {
		if _, err := mail.ParseAddress(addr); err != nil {
			return fmt.Errorf("invalid recipient %q: %w", addr, err)
		}
	}

	fromHeader := settings.FromEmail
	if settings.FromName != "" {
		fromHeader = (&mail.Address{Name: settings.FromName, Address: settings.FromEmail}).String()
	}

	msg := buildMessage(fromHeader, to, subject, body)
	addr := fmt.Sprintf("%s:%d", settings.Host, settings.Port)

	dialer := &net.Dialer{Timeout: 15 * time.Second}
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}

	var client *smtp.Client
	if settings.UseTLS && settings.Port == 465 {
		tlsConn := tls.Client(conn, &tls.Config{ServerName: settings.Host, MinVersion: tls.VersionTLS12})
		if err := tlsConn.HandshakeContext(ctx); err != nil {
			_ = conn.Close()
			return fmt.Errorf("TLS handshake failed: %w", err)
		}
		client, err = smtp.NewClient(tlsConn, settings.Host)
	} else {
		client, err = smtp.NewClient(conn, settings.Host)
	}
	if err != nil {
		_ = conn.Close()
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Close()

	if settings.UseTLS && settings.Port != 465 {
		if ok, _ := client.Extension("STARTTLS"); ok {
			if err := client.StartTLS(&tls.Config{ServerName: settings.Host, MinVersion: tls.VersionTLS12}); err != nil {
				return fmt.Errorf("STARTTLS failed: %w", err)
			}
		}
	}

	if settings.Username != "" {
		auth := smtp.PlainAuth("", settings.Username, settings.Password, settings.Host)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("SMTP authentication failed: %w", err)
		}
	}

	if err := client.Mail(settings.FromEmail); err != nil {
		return fmt.Errorf("SMTP MAIL FROM failed: %w", err)
	}
	for _, recipient := range to {
		if err := client.Rcpt(recipient); err != nil {
			return fmt.Errorf("SMTP RCPT TO failed for %s: %w", recipient, err)
		}
	}

	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("SMTP DATA failed: %w", err)
	}
	if _, err := writer.Write(msg); err != nil {
		_ = writer.Close()
		return fmt.Errorf("failed to write message: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close message writer: %w", err)
	}

	return client.Quit()
}

func buildMessage(from string, to []string, subject, body string) []byte {
	var b strings.Builder
	b.WriteString("From: ")
	b.WriteString(from)
	b.WriteString("\r\n")
	b.WriteString("To: ")
	b.WriteString(strings.Join(to, ", "))
	b.WriteString("\r\n")
	b.WriteString("Subject: ")
	b.WriteString(subject)
	b.WriteString("\r\n")
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	b.WriteString("\r\n")
	b.WriteString(body)
	return []byte(b.String())
}
