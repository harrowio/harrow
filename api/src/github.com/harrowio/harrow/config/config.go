package config

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

type Config struct {
	environment string
	Root        string
}

const InstanceDeadline = 2 * time.Hour

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

func getEnvBoolWithDefault(k string, d bool) bool {
	val, err := strconv.ParseBool(getEnvWithDefault(k, fmt.Sprintf("%t", d)))
	if err != nil {
		panic(fmt.Sprintf("can't parse env variable '%s' as boolean", k))
	}
	return val
}
