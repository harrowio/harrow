package runner

import (
	"fmt"
	"os"
	"time"

	"github.com/harrowio/harrow/config"
	"github.com/harrowio/harrow/logger"
	"github.com/influxdata/influxdb/client/v2"
)

// Reporter is the generic reporter for metrics
type Reporter interface {
	// Waiting/polling
	PolledNoWork()
	PolledFoundWork()

	// how long is our backlog?
	WaitTime(time.Duration)

	// did we fail to connect to the host?
	SSHError(error)

	// Heartbeats
	Heartbeat()
	LongPollLost()

	// Containers
	MadeContainer()
	DestroyedContainer()
	DestroyContainerWillRetry()

	// What're we doing operationally?
	Signal(os.Signal)
	CleanShutdown()
}

type NoopReporter struct{}

func (_ NoopReporter) PolledNoWork()              {}
func (_ NoopReporter) PolledFoundWork()           {}
func (_ NoopReporter) WaitTime(_ time.Duration)   {}
func (_ NoopReporter) SSHError(error)             {}
func (_ NoopReporter) Heartbeat()                 {}
func (_ NoopReporter) LongPollLost()              {}
func (_ NoopReporter) MadeContainer()             {}
func (_ NoopReporter) DestroyedContainer()        {}
func (_ NoopReporter) DestroyContainerWillRetry() {}
func (_ NoopReporter) Signal(os.Signal)           {}
func (_ NoopReporter) CleanShutdown()             {}

func NewInfluxDBReporter(logger logger.Logger, influxConf config.InfluxDBConfig, runnerName string) influxDBReporter {
	// Create a new HTTPClient
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     influxConf.Addr,
		Username: influxConf.Username,
		Password: influxConf.Password,
		Timeout:  influxConf.Timeout,
	})
	if err != nil {
		logger.Fatal().Msgf("can't make influxdb http client: %s", err)
	}

	q := client.NewQuery(fmt.Sprintf("CREATE DATABASE %s", influxConf.Database), "", "")
	c.Query(q)

	return influxDBReporter{c, map[string]string{"runner-name": runnerName}, influxConf.Database, logger}
}

type influxDBReporter struct {
	client      client.Client
	defaultTags map[string]string
	database    string

	logger logger.Logger
}

func (i influxDBReporter) recordPoint(pointName string, fields map[string]interface{}) {
	bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  i.database,
		Precision: "s",
	})
	pt, err := client.NewPoint(pointName, i.defaultTags, fields, time.Now())
	if err != nil {
		i.logger.Error().Msgf("error making influxdb point: %s", err)
	}
	bp.AddPoint(pt)
	i.client.Write(bp)
}

func (i influxDBReporter) PolledNoWork() {
	i.recordPoint("polled_no_work", map[string]interface{}{"count": 1})
}
func (i influxDBReporter) PolledFoundWork() {
	i.recordPoint("polled_found_work", map[string]interface{}{"count": 1})
}
func (i influxDBReporter) WaitTime(d time.Duration) {
	i.recordPoint("wait_time", map[string]interface{}{"count": 1, "duration_ms": d * time.Millisecond})
}
func (i influxDBReporter) SSHError(err error) {
	i.recordPoint("ssh_error", map[string]interface{}{"count": 1, "error_str": err.Error()})
}
func (i influxDBReporter) Heartbeat() {
	i.recordPoint("heart_beat", map[string]interface{}{"count": 1})
}
func (i influxDBReporter) LongPollLost() {
	i.recordPoint("long_poll_lost", map[string]interface{}{"count": 1})
}
func (i influxDBReporter) MadeContainer() {
	i.recordPoint("made_container", map[string]interface{}{"count": 1})
}
func (i influxDBReporter) DestroyedContainer() {
	i.recordPoint("destroyed_container", map[string]interface{}{"count": 1, "is_retry": false})
}
func (i influxDBReporter) DestroyContainerWillRetry() {
	i.recordPoint("destroyed_container", map[string]interface{}{"count": 0, "is_retry": true})
}
func (i influxDBReporter) Signal(s os.Signal) {
	i.recordPoint("got_signal", map[string]interface{}{"count": 1, "signal": s.String()})
}
func (i influxDBReporter) CleanShutdown() {
	i.recordPoint("shutdown", map[string]interface{}{"count": 1})
}
