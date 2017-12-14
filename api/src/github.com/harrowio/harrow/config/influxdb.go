package config

import "time"

type InfluxDBConfig struct {
	Addr     string
	Username string
	Password string
	Database string
	Timeout  time.Duration
}

func (c *Config) InfluxDBConfig() InfluxDBConfig {
	return InfluxDBConfig{
		Addr:     getEnvWithDefault("HAR_INFLUXDB_ADDR", "http://localhost/"),
		Username: getEnvWithDefault("HAR_INFLUXDB_USERNAME", ""),
		Password: getEnvWithDefault("HAR_INFLUXDB_PASSWORD", ""),
		Database: getEnvWithDefault("HAR_INFLUXDB_PASSWORD", "har_no_database_specified"),
		Timeout:  getEnvDurationWithDefault("HAR_INFLUXDB_TIMEOUT", 5*time.Second),
	}
}
