package config

import "github.com/spf13/viper"

// EmailConfig holds the global email settings
type EmailConfig struct {
	Host         string
	Port         int
	Username     string
	PasswordFile string
	TLS          bool
	From         string
	To           string
}

// LoadEmailConfig loads the email configuration from Viper
func LoadEmailConfig() EmailConfig {
	return EmailConfig{
		Host:         viper.GetString("email.host"),
		Port:         viper.GetInt("email.port"),
		Username:     viper.GetString("email.user"),
		PasswordFile: viper.GetString("email.passwordfile"),
		TLS:          viper.GetBool("email.tls"),
		From:         viper.GetString("email.from"),
		To:           viper.GetString("email.to"),
	}
}
