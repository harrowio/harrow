package config

import "strings"

func (c *Config) AmqpConnectionString() string {
	return strings.TrimSpace(getEnvWithDefault("HAR_RABBITMQ_DSN", "amqp://guest:guest@localhost:5672/"))
}
