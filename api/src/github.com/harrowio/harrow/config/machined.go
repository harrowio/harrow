package config

import (
	"os"
	"strconv"
	"time"
)

const InstanceDeadline = 2 * time.Hour

type MachinedConfig struct {
	AcquisitionImpl string
	AsName          string
	OperatorEnv     string
	UsePublicIp     bool
}

func (c *Config) MachinedConfig() MachinedConfig {
	var f func(string) bool = func(s string) bool {
		b, _ := strconv.ParseBool(os.Getenv(s))
		return b
	}
	return MachinedConfig{
		AcquisitionImpl: os.Getenv("HAR_MACHINED_ACQUISITION_IMPL"),
		AsName:          os.Getenv("HAR_MACHINED_AWS_AS_NAME"),
		OperatorEnv:     os.Getenv("HAR_MACHINED_OPERATOR_ENV"),
		UsePublicIp:     f("HAR_MACHINED_USE_PUBLIC_IP"),
	}
}
