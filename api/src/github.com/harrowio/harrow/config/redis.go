package config

import (
	"fmt"
	"os"

	redis "gopkg.in/redis.v2"
)

type redisConfig struct {
	host     string
	port     string
	password string
}

func (c *Config) RedisConfig() redisConfig {
	return redisConfig{
		host:     getEnvWithDefault("HAR_REDIS_HOST", "localhost"),
		port:     getEnvWithDefault("HAR_REDIS_PORT", "6379"),
		password: os.Getenv("HAR_REDIS_PASSWORD"),
	}
}

func (c *Config) RedisConnOpts(dbNum int64) *redis.Options {
	config := c.RedisConfig()
	return &redis.Options{
		Addr:     fmt.Sprintf("%s:%s", config.host, config.port),
		Password: config.password,
		DB:       dbNum,
	}
}
