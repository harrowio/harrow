package config

import "os"

type FilesystemConfig struct {
	LogDir     string `json:"log_dir"`
	GitTempDir string `json:"git_temp_dir"`
}

func (c *Config) FilesystemConfig() FilesystemConfig {
	return FilesystemConfig{
		LogDir:     os.Getenv("HAR_FILESYSTEM_LOG_DIR"),
		GitTempDir: os.Getenv("HAR_FILESYSTEM_GIT_TMP_DIR"),
	}
}
