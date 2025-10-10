package backup

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/smtp"
	"os"
	"os/exec"
	"strings"
	"time"
)

// SendNotification sends an email notification about backup status
func SendNotification(config *Config, status string, details string) error {
	if !config.Notifications.Enabled {
		return nil
	}

	// Check if we should send based on status
	if status == "failure" && !config.Notifications.OnFailure {
		return nil
	}
	if status == "success" && !config.Notifications.OnSuccess {
		return nil
	}

	emailConfig := config.Notifications.Email

	// Read SMTP password if file is provided
	var password string
	if emailConfig.SMTP.PasswordFile != "" {
		data, err := os.ReadFile(emailConfig.SMTP.PasswordFile)
		if err != nil {
			return fmt.Errorf("failed to read SMTP password file: %w", err)
		}
		password = strings.TrimSpace(string(data))
	}

	// Prepare email subject and body
	subject := fmt.Sprintf("Backup %s: %s", strings.ToUpper(status), config.Name)
	body := formatEmailBody(config, status, details)

	// Send email
	return sendEmail(emailConfig, password, subject, body)
}

// formatEmailBody creates the email body
func formatEmailBody(config *Config, status string, details string) string {
	var body strings.Builder

	body.WriteString(fmt.Sprintf("Backup Status: %s\n", strings.ToUpper(status)))
	body.WriteString(fmt.Sprintf("Backup Name: %s\n", config.Name))
	body.WriteString(fmt.Sprintf("Backup Type: %s\n", config.Type))
	body.WriteString(fmt.Sprintf("Timestamp: %s\n", time.Now().Format(time.RFC3339)))
	body.WriteString("\n")

	body.WriteString("Configuration:\n")
	body.WriteString(fmt.Sprintf("  Sources: %s\n", strings.Join(config.Source, ", ")))
	body.WriteString(fmt.Sprintf("  Destination: %s\n", config.GetDestination()))
	body.WriteString(fmt.Sprintf("  Schedule: %s\n", config.Schedule))
	body.WriteString("\n")

	if details != "" {
		body.WriteString("Details:\n")
		body.WriteString(details)
		body.WriteString("\n")
	}

	body.WriteString("\n---\n")
	body.WriteString("This is an automated message from quadlet-helper backup system.\n")

	return body.String()
}

// sendEmail sends an email using SMTP
func sendEmail(emailConfig EmailConfig, password, subject, body string) error {
	from := emailConfig.From
	to := []string{emailConfig.To}

	// Prepare message
	msg := []byte(fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"\r\n"+
		"%s\r\n", from, emailConfig.To, subject, body))

	addr := fmt.Sprintf("%s:%d", emailConfig.SMTP.Host, emailConfig.SMTP.Port)

	// Setup authentication
	var auth smtp.Auth
	if emailConfig.SMTP.Username != "" {
		auth = smtp.PlainAuth("", emailConfig.SMTP.Username, password, emailConfig.SMTP.Host)
	}

	// Send email
	if emailConfig.SMTP.TLS {
		return sendEmailTLS(addr, auth, from, to, msg, emailConfig.SMTP.Host)
	}

	return smtp.SendMail(addr, auth, from, to, msg)
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
func SendTestEmail(config *Config) error {
	subject := fmt.Sprintf("Test Email from quadlet-helper: %s", config.Name)
	body := fmt.Sprintf("This is a test email to verify the email notification configuration for backup '%s'.\n\n"+
		"If you received this email, your notification settings are configured correctly.\n\n"+
		"Backup configuration:\n"+
		"  Type: %s\n"+
		"  Sources: %s\n"+
		"  Destination: %s\n"+
		"  Schedule: %s\n",
		config.Name, config.Type, strings.Join(config.Source, ", "), config.GetDestination(), config.Schedule)

	emailConfig := config.Notifications.Email

	// Read SMTP password
	var password string
	if emailConfig.SMTP.PasswordFile != "" {
		data, err := os.ReadFile(emailConfig.SMTP.PasswordFile)
		if err != nil {
			return fmt.Errorf("failed to read SMTP password file: %w", err)
		}
		password = strings.TrimSpace(string(data))
	}

	return sendEmail(emailConfig, password, subject, body)
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
