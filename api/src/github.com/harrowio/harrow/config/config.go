package config

import (
	"log"
	"os"
	"strings"

	_ "github.com/lib/pq"
)

type Config struct {
	environment string
	Root        string
}

//
// Unexported Package varaibles to hold memoized
// values
//
var c *Config
var env string = "development"

func GetConfig() *Config {
	return &Config{}
}

func (c *Config) Environment() string {
	if len(c.environment) > 0 {
		return c.environment
	} else {
		return "development"
	}
}

func getEnvWithDefault(k, d string) string {
	val := strings.TrimSpace(os.Getenv(k))
	if val == "" && d == "" {
		log.Fatalf("Config Error: No default value for %s available, please export it.\n", k)
	}
	if val == "" && d != "" {
		val = d
	}
	return val
}
