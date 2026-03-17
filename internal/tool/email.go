// Package tool provides built-in tools
package tool

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"
)

// EmailSendTool sends emails via SMTP
type EmailSendTool struct {
	config EmailConfig
}

// EmailConfig holds email configuration
type EmailConfig struct {
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	FromAddress  string
}

// NewEmailSendTool creates a new email send tool
func NewEmailSendTool(config EmailConfig) *EmailSendTool {
	return &EmailSendTool{config: config}
}

func (t *EmailSendTool) Name() string {
	return "email_send"
}

func (t *EmailSendTool) Description() string {
	return "Send an email via SMTP"
}

func (t *EmailSendTool) Parameters() interface{} {
	return ObjectParam{
		Type: "object",
		Properties: map[string]interface{}{
			"to": StringParam{
				Type:        "string",
				Description: "Recipient email address",
			},
			"subject": StringParam{
				Type:        "string",
				Description: "Email subject",
			},
			"body": StringParam{
				Type:        "string",
				Description: "Email body content",
			},
			"cc": StringParam{
				Type:        "string",
				Description: "CC recipients (comma-separated, optional)",
			},
			"bcc": StringParam{
				Type:        "string",
				Description: "BCC recipients (comma-separated, optional)",
			},
			"html": BooleanParam{
				Type:        "boolean",
				Description: "Whether body is HTML (default: false)",
			},
		},
		Required: []string{"to", "subject", "body"},
	}
}

func (t *EmailSendTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	to, ok := params["to"].(string)
	if !ok {
		return nil, fmt.Errorf("to parameter is required")
	}

	subject, ok := params["subject"].(string)
	if !ok {
		return nil, fmt.Errorf("subject parameter is required")
	}

	body, ok := params["body"].(string)
	if !ok {
		return nil, fmt.Errorf("body parameter is required")
	}

	// Build email message
	from := t.config.FromAddress
	if from == "" {
		from = t.config.SMTPUsername
	}

	// Build headers
	headers := make(map[string]string)
	headers["From"] = from
	headers["To"] = to
	headers["Subject"] = subject

	if cc, ok := params["cc"].(string); ok && cc != "" {
		headers["Cc"] = cc
	}

	// Content type
	contentType := "text/plain"
	if html, ok := params["html"].(bool); ok && html {
		contentType = "text/html"
	}
	headers["Content-Type"] = contentType + "; charset=UTF-8"

	// Build message
	var msg strings.Builder
	for k, v := range headers {
		msg.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	msg.WriteString("\r\n")
	msg.WriteString(body)

	// Build recipients
	recipients := []string{to}
	if cc, ok := params["cc"].(string); ok && cc != "" {
		recipients = append(recipients, strings.Split(cc, ",")...)
	}
	if bcc, ok := params["bcc"].(string); ok && bcc != "" {
		recipients = append(recipients, strings.Split(bcc, ",")...)
	}

	// Send email
	addr := fmt.Sprintf("%s:%d", t.config.SMTPHost, t.config.SMTPPort)
	auth := smtp.PlainAuth("", t.config.SMTPUsername, t.config.SMTPPassword, t.config.SMTPHost)

	err := smtp.SendMail(addr, auth, from, recipients, []byte(msg.String()))
	if err != nil {
		return nil, fmt.Errorf("failed to send email: %w", err)
	}

	return map[string]interface{}{
		"status":  "sent",
		"to":      to,
		"subject": subject,
	}, nil
}

// EmailReadTool reads emails (placeholder for IMAP)
type EmailReadTool struct {
	config EmailConfig
}

// NewEmailReadTool creates a new email read tool
func NewEmailReadTool(config EmailConfig) *EmailReadTool {
	return &EmailReadTool{config: config}
}

func (t *EmailReadTool) Name() string {
	return "email_read"
}

func (t *EmailReadTool) Description() string {
	return "Read emails from inbox (requires IMAP configuration)"
}

func (t *EmailReadTool) Parameters() interface{} {
	return ObjectParam{
		Type: "object",
		Properties: map[string]interface{}{
			"folder": StringParam{
				Type:        "string",
				Description: "Email folder to read (default: INBOX)",
			},
			"limit": NumberParam{
				Type:        "number",
				Description: "Maximum number of emails to return (default: 10)",
			},
			"unread_only": BooleanParam{
				Type:        "boolean",
				Description: "Only return unread emails (default: false)",
			},
		},
	}
}

func (t *EmailReadTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	// This is a placeholder - IMAP implementation would go here
	// For production, use a library like github.com/emersion/go-imap

	folder := "INBOX"
	if f, ok := params["folder"].(string); ok {
		folder = f
	}

	limit := 10
	if l, ok := params["limit"].(float64); ok {
		limit = int(l)
	}

	return map[string]interface{}{
		"message": "IMAP email reading requires additional configuration",
		"folder":  folder,
		"limit":   limit,
		"emails":  []interface{}{},
	}, nil
}

// EmailListTool lists email folders
type EmailListTool struct {
	config EmailConfig
}

// NewEmailListTool creates a new email list tool
func NewEmailListTool(config EmailConfig) *EmailListTool {
	return &EmailListTool{config: config}
}

func (t *EmailListTool) Name() string {
	return "email_list_folders"
}

func (t *EmailListTool) Description() string {
	return "List email folders"
}

func (t *EmailListTool) Parameters() interface{} {
	return ObjectParam{
		Type:       "object",
		Properties: map[string]interface{}{},
	}
}

func (t *EmailListTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	// Placeholder for IMAP folder listing
	return map[string]interface{}{
		"folders": []string{"INBOX", "Sent", "Drafts", "Trash", "Spam"},
		"message": "IMAP folder listing requires additional configuration",
	}, nil
}
