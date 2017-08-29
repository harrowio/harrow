package config

type LimitsStoreConfig struct {
	CachePath string
	Port      string
	Bind      string
	FailMode  string
	Enabled   bool
}

func (c *Config) LimitsStoreConfig() LimitsStoreConfig {
	return LimitsStoreConfig{
		CachePath: getEnvWithDefault("HAR_LIMIT_STORE_CACHE_DIR", "/tmp/harrow/limits/"),
		Port:      getEnvWithDefault("HAR_LIMIT_STORE_PORT", "5371"),
		Bind:      getEnvWithDefault("HAR_LIMIT_STORE_BIND", "0.0.0.0"),
		FailMode:  getEnvWithDefault("HAR_LIMIT_STORE_FAIL_MODE", "assume_allowed"),
		Enabled:   getEnvBoolWithDefault("HAR_LIMITS_ENABLED", false),
	}
}
