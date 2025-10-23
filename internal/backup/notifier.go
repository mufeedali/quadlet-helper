package backup

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"html/template"
	"io"
	"net/smtp"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/mufeedali/quadlet-helper/internal/config"
)

// SendNotification sends an email notification about backup status
func SendNotification(backupConfig *Config, status string, details string) error {
	if !backupConfig.Notifications.Enabled {
		return nil
	}

	// Check if we should send based on status
	if status == "failure" && !backupConfig.Notifications.OnFailure {
		return nil
	}
	if status == "success" && !backupConfig.Notifications.OnSuccess {
		return nil
	}

	emailConf := config.LoadEmailConfig()
	backupEmailConf := backupConfig.Notifications.Email

	// Read SMTP password if file is provided
	var password string
	if emailConf.PasswordFile != "" {
		data, err := os.ReadFile(emailConf.PasswordFile)
		if err != nil {
			return fmt.Errorf("failed to read SMTP password file: %w", err)
		}
		password = strings.TrimSpace(string(data))
	}

	// Prepare email subject and body
	subject := fmt.Sprintf("Backup %s: %s", strings.ToUpper(status), backupConfig.Name)
	body, err := formatEmailBody(backupConfig, status, details)
	if err != nil {
		return fmt.Errorf("failed to format email body: %w", err)
	}

	// Determine From address
	from := emailConf.From
	if backupEmailConf.From != "" {
		from = backupEmailConf.From
	}
	if from == "" {
		return fmt.Errorf("no sender email address configured")
	}

	// Determine To address
	to := emailConf.To
	if backupEmailConf.To != "" {
		to = backupEmailConf.To
	}
	if to == "" {
		return fmt.Errorf("no recipient email address configured")
	}

	// Send email
	return sendEmail(emailConf, password, from, to, subject, body)
}

// formatEmailBody creates the email body
func formatEmailBody(backupConfig *Config, status string, details string) (string, error) {
	bodyTemplate := `
	<!DOCTYPE html>
	<html>
	<head>
		<style>
			body { font-family: sans-serif; }
			.container { padding: 20px; border: 1px solid #ddd; border-radius: 5px; max-width: 600px; margin: auto; }
			.status { font-size: 20px; font-weight: bold; }
			.status.success { color: green; }
			.status.failure { color: red; }
			.status.test { color: blue; }
			.details { background-color: #f5f5f5; padding: 15px; border-radius: 3px; white-space: pre-wrap; font-family: monospace; }
			table { border-collapse: collapse; width: 100%; margin-bottom: 20px; border: 1px solid #ddd; }
			th, td { text-align: left; padding: 8px; border: 1px solid #ddd; }
			th { background-color: #f2f2f2; }
			ul { margin: 0; padding-left: 20px; }
		</style>
	</head>
	<body>
		<div class="container">
			<h2>Backup Report</h2>
			<p>
				<span class="status {{ .StatusClass }}">{{ .Status }}</span>
			</p>
			<table>
				<tr><th>Backup Name</th><td>{{ .Name }}</td></tr>
				<tr><th>Backup Type</th><td>{{ .Type }}</td></tr>
				<tr><th>Timestamp</th><td>{{ .Timestamp }}</td></tr>
				<tr><th>Sources</th><td>{{ if gt (len .Sources) 1 }}<ul>{{ range .Sources }}<li>{{ . }}</li>{{ end }}</ul>{{ else }}{{ index .Sources 0 }}{{ end }}</td></tr>
				<tr><th>Destination</th><td>{{ .Destination }}</td></tr>
				<tr><th>Schedule</th><td>{{ .Schedule }}</td></tr>
			</table>

			{{ if .Details }}
			<h3>Details:</h3>
			<pre class="details">{{ .Details }}</pre>
			{{ end }}

			<p><small>This is an automated message from quadlet-helper.</small></p>
		</div>
	</body>
	</html>`

	tmpl, err := template.New("email").Parse(bodyTemplate)
	if err != nil {
		return "", err
	}

	data := struct {
		Status      string
		StatusClass string
		Name        string
		Type        string
		Timestamp   string
		Sources     []string
		Destination string
		Schedule    string
		Details     string
	}{
		Status:      strings.ToUpper(status),
		StatusClass: status, // "success" or "failure"
		Name:        backupConfig.Name,
		Type:        string(backupConfig.Type),
		Timestamp:   time.Now().Format(time.RFC1123Z),
		Sources:     backupConfig.Source,
		Destination: backupConfig.GetDestination(),
		Schedule:    backupConfig.Schedule,
		Details:     details,
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		return "", err
	}

	return body.String(), nil
}

// extractEmailAddress extracts the email address from a string that may contain a display name
// e.g., "Display Name" <email@example.com> -> email@example.com
func extractEmailAddress(addr string) string {
	addr = strings.TrimSpace(addr)
	// Check if the address contains angle brackets
	if strings.Contains(addr, "<") && strings.Contains(addr, ">") {
		start := strings.Index(addr, "<")
		end := strings.Index(addr, ">")
		if start < end {
			return strings.TrimSpace(addr[start+1 : end])
		}
	}
	return addr
}

// sendEmail sends an email using SMTP
func sendEmail(emailConf config.EmailConfig, password, from, to, subject, body string) error {
	toList := []string{extractEmailAddress(to)}
	fromAddr := extractEmailAddress(from)

	// Prepare message with HTML content type
	msg := []byte(fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"+
		"%s\r\n", from, to, subject, body))

	addr := fmt.Sprintf("%s:%d", emailConf.Host, emailConf.Port)

	// Setup authentication
	var auth smtp.Auth
	if emailConf.Username != "" {
		auth = smtp.PlainAuth("", emailConf.Username, password, emailConf.Host)
	}

	// Send email
	if emailConf.TLS {
		return sendEmailTLS(addr, auth, fromAddr, toList, msg, emailConf.Host)
	}

	return smtp.SendMail(addr, auth, fromAddr, toList, msg)
}

// sendEmailTLS sends email using TLS
func sendEmailTLS(addr string, auth smtp.Auth, from string, to []string, msg []byte, serverName string) error {
	tlsConfig := &tls.Config{
		ServerName: serverName,
	}

	// For port 465 (SMTPS), use direct TLS connection
	// For port 587 (STARTTLS), use smtp.Dial then StartTLS
	var client *smtp.Client
	var err error

	// Try direct TLS connection first (for port 465)
	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err == nil {
		client, err = smtp.NewClient(conn, serverName)
		if err != nil {
			conn.Close()
			return fmt.Errorf("failed to create SMTP client: %w", err)
		}
	} else {
		// Fall back to STARTTLS (for port 587)
		client, err = smtp.Dial(addr)
		if err != nil {
			return fmt.Errorf("failed to connect to SMTP server: %w", err)
		}

		if err = client.StartTLS(tlsConfig); err != nil {
			client.Close()
			return fmt.Errorf("failed to start TLS: %w", err)
		}
	}
	defer client.Close()

	// Authenticate
	if auth != nil {
		if err = client.Auth(auth); err != nil {
			return fmt.Errorf("failed to authenticate: %w", err)
		}
	}

	// Set sender
	if err = client.Mail(from); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	// Set recipients
	for _, addr := range to {
		if err = client.Rcpt(addr); err != nil {
			return fmt.Errorf("failed to set recipient: %w", err)
		}
	}

	// Send message
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to get data writer: %w", err)
	}

	_, err = w.Write(msg)
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("failed to close data writer: %w", err)
	}

	return client.Quit()
}

// SendTestEmail sends a test email to verify configuration
func SendTestEmail(backupConfig *Config) error {
	subject := fmt.Sprintf("Test Email from quadlet-helper: %s", backupConfig.Name)
	body, err := formatEmailBody(backupConfig, "test", "This is a test email to verify your notification settings.")
	if err != nil {
		return fmt.Errorf("failed to format test email body: %w", err)
	}

	emailConf := config.LoadEmailConfig()
	backupEmailConf := backupConfig.Notifications.Email

	// Read SMTP password
	var password string
	if emailConf.PasswordFile != "" {
		data, err := os.ReadFile(emailConf.PasswordFile)
		if err != nil {
			return fmt.Errorf("failed to read SMTP password file: %w", err)
		}
		password = strings.TrimSpace(string(data))
	}

	from := emailConf.From
	if backupEmailConf.From != "" {
		from = backupEmailConf.From
	}

	return sendEmail(emailConf, password, from, backupEmailConf.To, subject, body)
}

// GetLastBackupLog retrieves the last backup log from systemd journal
func GetLastBackupLog(backupName string) (string, error) {
	serviceName := fmt.Sprintf("%s-backup.service", backupName)
	cmd := fmt.Sprintf("journalctl --user -u %s -n 100 --no-pager", serviceName)

	// Use sh to execute the command
	output, err := execCommand("/bin/sh", "-c", cmd)
	if err != nil {
		return "", fmt.Errorf("failed to get logs: %w", err)
	}

	return output, nil
}

// execCommand is a helper to execute commands and return output
func execCommand(name string, args ...string) (string, error) {
	cmd := &exec.Cmd{
		Path: name,
		Args: append([]string{name}, args...),
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}

	if err := cmd.Start(); err != nil {
		return "", err
	}

	output, err := io.ReadAll(stdout)
	if err != nil {
		return "", err
	}

	if err := cmd.Wait(); err != nil {
		return string(output), err
	}

	return string(output), nil
}
