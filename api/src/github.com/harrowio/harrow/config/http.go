package config

import (
	"crypto/rand"
	"fmt"
	"os"
	"strconv"
)

type HttpConfig struct {
	Port                 int    `json:"port"`
	Bind                 string `json:"bind"`
	UserHmacSecret       string `json:"-"`
	WebSocketPort        int    `json:"websocketPort"`
	WebSocketBind        string `json:"websocketBind"`
	MaxSimultaneousConns int
}

func (h HttpConfig) String() string {
	return fmt.Sprintf("%s:%d", h.Bind, h.Port)
}

func (h HttpConfig) WebSocket() string {
	return fmt.Sprintf("%s:%d", h.WebSocketBind, h.WebSocketPort)
}

// Returns the HTTP config, interface{} as the
// keys are a mix of strings and integers
func (c *Config) HttpConfig() HttpConfig {
	var f func(string, string) int = func(s, d string) int {
		v := os.Getenv(s)
		if len(v) == 0 {
			v = d
		}
		b, _ := strconv.ParseInt(v, 10, 0)
		return int(b)
	}

	b := make([]byte, 40)
	_, err := rand.Read(b)
	if err != nil {
		panic("can't read random bytes, erring")
	}

	return HttpConfig{
		Port:                 f("HAR_HTTP_PORT", "8080"),
		Bind:                 getEnvWithDefault("HAR_HTTP_BIND", "0.0.0.0"),
		UserHmacSecret:       getEnvWithDefault("HAR_HTTP_USER_HMAC_SECRET", string(b)),
		WebSocketPort:        f("HAR_HTTP_WEBSOCKET_PORT", "8383"),
		WebSocketBind:        getEnvWithDefault("HAR_HTTP_WEBSOCKET_BIND", "0.0.0.0"),
		MaxSimultaneousConns: f("HAR_HTTP_MAX_CONNS", "50"),
	}
}
