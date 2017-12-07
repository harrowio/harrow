package opMetrics

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/rs/zerolog"

	"github.com/harrowio/harrow/config"
)

const ProgramName = "op-metrics"

var log zerolog.Logger = zerolog.New(os.Stdout).With().Str("harrow", ProgramName).Timestamp().Logger()

func Main() {

	var (
		webhookKey string
		metricsID  string
		watchdogID string
	)

	flag.StringVar(&webhookKey, "webhook-key", "*******", "value for X-WEBHOOK-KEY header")
	flag.StringVar(&metricsID, "metrics-id", "*******", "value for metrics id (url)")
	flag.StringVar(&watchdogID, "watchdog-id", "*******", "value for watchdog id (url)")

	flag.Parse()

	if len(webhookKey) == 0 || len(metricsID) == 0 || len(watchdogID) == 0 {
		flag.Usage()
		os.Exit(123)
	}

	c := config.GetConfig()
	db, err := c.DB()
	if err != nil {
		log.Fatal().Err(err)
	}

	db.MustExec(`CREATE OR REPLACE FUNCTION round_time(TIMESTAMP WITH TIME ZONE)
	RETURNS TIMESTAMP WITH TIME ZONE AS $$
	  SELECT date_trunc('hour', $1) + INTERVAL '5 min' * ROUND(date_part('minute', $1) / 5.0)
		$$ LANGUAGE SQL;`)

	type TimeoutCount struct {
		Count int       `db:"count"`
		Date  time.Time `db:"bucket"`
	}
	var timeouts []TimeoutCount
	err = db.Select(&timeouts, "SELECT COUNT(*), round_time(timed_out_at) AS bucket FROM operations WHERE timed_out_at IS NOT NULL AND created_at > NOW() - interval '2 hours' GROUP BY round_time(timed_out_at);")
	if err != nil {
		panic(err)
	}

	var pendingOpCount int
	err = db.Get(&pendingOpCount, "SELECT COUNT(*) FROM operations WHERE (started_at IS NULL) AND (canceled_at IS NULL AND timed_out_at IS NULL AND failed_at IS NULL AND finished_at IS NULL AND archived_at IS NULL);")
	if err != nil {
		panic(err)
	}

	type WaitTimeStat struct {
		Pct00 int       `db:"pct00_sec"`
		Pct99 int       `db:"pct99_sec"`
		Pct95 int       `db:"pct95_sec"`
		Mean  int       `db:"mean_sec"`
		Date  time.Time `db:"bucket"`
	}

	var q string = `
		SELECT
		floor(date_part('epoch', percentile_disc(1) WITHIN GROUP (ORDER BY started_at - created_at)))::integer AS pct00_sec,
		floor(date_part('epoch', percentile_disc(.99) WITHIN GROUP (ORDER BY started_at - created_at)))::integer AS pct99_sec,
		floor(date_part('epoch', percentile_disc(.95) WITHIN GROUP (ORDER BY started_at - created_at)))::integer AS pct95_sec,
		floor(date_part('epoch', AVG(started_at - created_at)))::integer AS mean_sec,
		round_time(created_at) AS bucket
		FROM operations WHERE created_at > NOW() - interval '2h' AND started_at IS NOT NULL GROUP BY round_time(created_at);
	`
	var waitTimeStats []WaitTimeStat

	err = db.Select(&waitTimeStats, q)
	if err != nil {
		panic(err)
	}

	type Metric struct {
		T int    `json:"t,omitempty"`
		X int64  `json:"x,omitempty"`
		Y int    `json:"y,omitempty"`
		C string `json:"c,omitempty"`
	}

	type Payload struct {
		Metrics map[string][]Metric `json:"metrics"`
	}

	var (
		p                   Payload  = Payload{make(map[string][]Metric)}
		timeoutMetrics      []Metric = []Metric{}
		waitTime00Metrics   []Metric = []Metric{}
		waitTime99Metrics   []Metric = []Metric{}
		waitTime95Metrics   []Metric = []Metric{}
		waitTimeMeanMetrics []Metric = []Metric{}
		pendingOpCounts     []Metric = []Metric{{X: time.Now().Unix(), Y: pendingOpCount}}
	)

	for _, t := range timeouts {
		timeoutMetrics = append(timeoutMetrics, Metric{Y: t.Count, X: t.Date.Unix()})
	}

	for _, t := range waitTimeStats {
		waitTime00Metrics = append(waitTime00Metrics, Metric{Y: t.Pct00, X: t.Date.Unix()})
		waitTime99Metrics = append(waitTime99Metrics, Metric{Y: t.Pct99, X: t.Date.Unix()})
		waitTime95Metrics = append(waitTime95Metrics, Metric{Y: t.Pct95, X: t.Date.Unix()})
		waitTimeMeanMetrics = append(waitTimeMeanMetrics, Metric{Y: t.Mean, X: t.Date.Unix()})
	}

	if len(timeoutMetrics) > 0 {
		p.Metrics["timeouts"] = timeoutMetrics
	}
	p.Metrics["wait_time_00pct"] = waitTime00Metrics
	p.Metrics["wait_time_99pct"] = waitTime99Metrics
	p.Metrics["wait_time_95pct"] = waitTime95Metrics
	p.Metrics["wait_time_mean"] = waitTimeMeanMetrics
	p.Metrics["pending_op_count"] = pendingOpCounts

	b, err := json.MarshalIndent(&p, "", "\t")

	var (
		metricsURL  string = fmt.Sprintf("https://status.harrow.io/state_webhook/metrics/%s", metricsID)
		watchdogURL string = fmt.Sprintf("https://status.harrow.io/state_webhook/watchdog/%s", watchdogID)
	)

	req, err := http.NewRequest("POST", metricsURL, bytes.NewReader(b))
	req.Header.Set("X-WEBHOOK-KEY", webhookKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	fmt.Println("Posting Metrics To:", metricsURL)
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	fmt.Print(string(b))

	lastWaitTimeStat := waitTimeStats[len(waitTimeStats)-1]
	if lastWaitTimeStat.Pct95 > 300 { // 5m
		fmt.Println("last wait time too high, setting service degraded")
		req, err := http.NewRequest("POST", watchdogURL, bytes.NewReader(b))
		req.Header.Set("X-WEBHOOK-KEY", webhookKey)

		q := req.URL.Query()
		q.Add("status", "0")
		req.URL.RawQuery = q.Encode()

		client := &http.Client{}
		fmt.Println("Posting Watchdog To:", watchdogURL)
		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()

		fmt.Println("response Status:", resp.Status)
		fmt.Print(string(b))

	} else {
		fmt.Printf("Last Wait Time: 00pct: %d 99pct: %d 95pct: %d Mean: %d\n", lastWaitTimeStat.Pct00, lastWaitTimeStat.Pct99, lastWaitTimeStat.Pct95, lastWaitTimeStat.Mean)
		req, err := http.NewRequest("POST", watchdogURL, bytes.NewReader(b))
		req.Header.Set("X-WEBHOOK-KEY", webhookKey)

		q := req.URL.Query()
		q.Add("status", "1")
		req.URL.RawQuery = q.Encode()

		client := &http.Client{}
		fmt.Println("Posting Watchdog To:", watchdogURL)
		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()
	}

}
