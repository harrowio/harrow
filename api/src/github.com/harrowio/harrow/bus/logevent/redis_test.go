package logevent

import (
	"bytes"
	"testing"

	"github.com/harrowio/harrow/config"
	"github.com/harrowio/harrow/logger"
	"github.com/harrowio/harrow/loxer"
	"github.com/rs/zerolog"

	redis "gopkg.in/redis.v2"
)

const (
	n = 10000
)

func Test_RedisSource(t *testing.T) {
	t.Skip("blocks indefinitely for some reason")
	c := config.GetConfig()
	redisClient := redis.NewTCPClient(c.RedisConnOpts(0))
	_, err := redisClient.FlushAll().Result()
	if err != nil {
		panic(err)
	}
	var buf bytes.Buffer
	var log logger.Logger = zerolog.New(&buf)
	transport := NewRedisTransport(redisClient, log)
	defer transport.Close()
	msgs, err := transport.Consume("1234")
	if err != nil {
		panic(err)
	}
	go func() {
		for i := 0; i < n; i++ {
			transport.Publish("1234", 0, 000000000000000, &loxer.TextEvent{loxer.EventData{Type: "text", Text: "ab", Offset: 0}})
		}
		transport.Close()
	}()
	cnt := 0
	for _ = range msgs {
		cnt++
	}
	if cnt != n {
		t.Fatalf("cnt = %d; want %d", cnt, n)
	}
}
