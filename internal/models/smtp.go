package models

import "time"

// SMTPSettings holds outbound email SMTP configuration.
type SMTPSettings struct {
	ID        string    `json:"id"`
	Host      string    `json:"host"`
	Port      int       `json:"port"`
	Username  string    `json:"username"`
	Password  string    `json:"password"`
	FromEmail string    `json:"from_email"`
	FromName  string    `json:"from_name"`
	UseTLS    bool      `json:"use_tls"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SMTPSettingsResponse is the API representation with password masked.
type SMTPSettingsResponse struct {
	Host        string `json:"host"`
	Port        int    `json:"port"`
	Username    string `json:"username"`
	FromEmail   string `json:"from_email"`
	FromName    string `json:"from_name"`
	UseTLS      bool   `json:"use_tls"`
	Enabled     bool   `json:"enabled"`
	HasPassword bool   `json:"has_password"`
}

// UpdateSMTPSettingsRequest is the payload to upsert SMTP configuration.
// Empty Password leaves the stored password unchanged.
type UpdateSMTPSettingsRequest struct {
	Host      string `json:"host"`
	Port      int    `json:"port"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	FromEmail string `json:"from_email"`
	FromName  string `json:"from_name"`
	UseTLS    bool   `json:"use_tls"`
	Enabled   bool   `json:"enabled"`
}

// TestSMTPRequest optionally overrides saved settings and/or sets a test recipient.
type TestSMTPRequest struct {
	To        string `json:"to"`
	Host      string `json:"host"`
	Port      int    `json:"port"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	FromEmail string `json:"from_email"`
	FromName  string `json:"from_name"`
	UseTLS    *bool  `json:"use_tls"`
}

// SendEmailRequest is used by the mail service to send a message.
type SendEmailRequest struct {
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	Body    string   `json:"body"`
}

// ToSMTPSettingsResponse maps domain settings to a safe API response.
func ToSMTPSettingsResponse(s *SMTPSettings) SMTPSettingsResponse {
	if s == nil {
		return SMTPSettingsResponse{Port: 587, UseTLS: true}
	}
	return SMTPSettingsResponse{
		Host:        s.Host,
		Port:        s.Port,
		Username:    s.Username,
		FromEmail:   s.FromEmail,
		FromName:    s.FromName,
		UseTLS:      s.UseTLS,
		Enabled:     s.Enabled,
		HasPassword: s.Password != "",
	}
}
