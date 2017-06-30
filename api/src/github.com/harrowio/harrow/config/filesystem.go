package config

type FilesystemConfig struct {
	OpLogDir   string `json:"log_dir"`
	GitTempDir string `json:"git_temp_dir"`
}

func (c *Config) FilesystemConfig() FilesystemConfig {
	return FilesystemConfig{
		OpLogDir:   getEnvWithDefault("HAR_FILESYSTEM_OP_LOG_DIR", "/tmp/harrow/op-logs/"),
		GitTempDir: getEnvWithDefault("HAR_FILESYSTEM_GIT_TMP_DIR", "/tmp/harrow/git/"),
	}
}
