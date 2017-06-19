package zob

import (
	"fmt"

	"github.com/harrowio/harrow/bus/broadcast"
	"github.com/harrowio/harrow/config"
	"github.com/rs/zerolog/log"

	"encoding/json"
	"time"

	"github.com/lib/pq"
)

const ProgramName = "zob"

func Main() {

	c := config.GetConfig()
	db, err := c.DB()
	if err != nil {
		log.Fatal().Msgf("Error Opening Database Handle: %s\n", err)
	}
	if err = db.Ping(); err != nil {
		log.Fatal().Msgf("Database Dead: %s\n", err)
	}

	bus := broadcast.NewAMQPTransport(c.AmqpConnectionString(), "zob")
	defer bus.Close()

	datasource, err := c.PgDataSourceName()
	if err != nil {
		log.Fatal().Msgf("Failed to request PgDataSourceName: %s", err)
	}

	zob := NewZOB(bus, datasource)

	go zob.RouteMessages("broadcast/create", broadcast.Create)
	go zob.RouteMessages("broadcast/change", broadcast.Change)

	select {}
}

type ZOB struct {
	bus        broadcast.Sink
	datasource string
}

func NewZOB(bus broadcast.Sink, datasource string) *ZOB {
	return &ZOB{
		bus:        bus,
		datasource: datasource,
	}
}

func (self *ZOB) RouteMessages(pgChannelName string, kind broadcast.BroadcastMessageType) {

	var message *pq.Notification

	listener := pq.NewListener(self.datasource, 10*time.Second, time.Minute, nil)
	if err := listener.Listen(pgChannelName); err != nil {
		log.Fatal().Msgf("Error listening on pg channel %q: %s", pgChannelName, err)
	}
	log.Info().Msgf("Listening for NOTIFY messages on pg channel: %s", pgChannelName)

	for {
		select {
		case message = <-listener.Notify:
			var dbBroadcast = struct {
				Table string
				New   struct {
					Uuid string
					Id   *int
				}
			}{}

			if err := json.Unmarshal([]byte(message.Extra), &dbBroadcast); err != nil {
				log.Error().Msgf("Failed to parse pg message: %s\nBody:\n%s\n", err, message.Extra)
				continue
			}

			table := dbBroadcast.Table
			id := dbBroadcast.New.Uuid
			log.Debug().Msgf("%s %s %s", kind, table, id)
			if table == "activities" {
				// NOTE(dh): if this fails, we have other problems.
				id = fmt.Sprintf("%d", *dbBroadcast.New.Id)
			}

			if err := self.bus.Publish(string(kind), table, id); err != nil {
				log.Error().Msgf("Failed to publish %s/%s@%s: %s",
					kind, table, id, err)
			}
		}
	}
}
