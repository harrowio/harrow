package config

import (
	"net/url"
	"os"
)

type URLConfig struct {
	Host string
}

func (c *Config) URLConfig() *URLConfig {
	return &URLConfig{
		Host: os.Getenv("HAR_URL_HOST"),
	}
}

func (self *URLConfig) Base() *url.URL {
	return &url.URL{
		Scheme: "https",
		Host:   self.Host,
	}
}
