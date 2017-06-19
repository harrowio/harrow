package config

import "os"

type LimitsStoreConfig struct {
	CachePath string `json:"cache_path"`
}

func (c *Config) LimitsStoreConfig() LimitsStoreConfig {
	return LimitsStoreConfig{
		CachePath: os.Getenv("HAR_LIMIT_STORE_CACHE_DIR"),
	}
}
