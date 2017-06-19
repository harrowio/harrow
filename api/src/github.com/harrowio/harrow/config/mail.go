package config

import "os"

type MailConfig struct {
	FromAddress string `json:"fromAddress"`
}

func (c *Config) MailConfig() MailConfig {
	return MailConfig{
		FromAddress: os.Getenv("HAR_MAIL_FROM_ADDRESS"),
	}
}
