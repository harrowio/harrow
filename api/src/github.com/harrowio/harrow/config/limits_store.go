package config

type LimitsStoreConfig struct {
	CachePath string `json:"cache_path"`
}

func (c *Config) LimitsStoreConfig() LimitsStoreConfig {
	return LimitsStoreConfig{
		CachePath: getEnvWithDefault("HAR_LIMIT_STORE_CACHE_DIR", "/tmp/harrow/limits/"),
	}
}
