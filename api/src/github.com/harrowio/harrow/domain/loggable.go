package domain

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
)

const (
	LoggableOK = iota
	LoggableEndOfTransmission
)

const (
	LoggableWorkspace = "workspace"
	LoggableVerbose   = "verbose"
)

const (
	intLogPattern = `([0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}Z)\s[IOE]{1}:\s`
)

var intLogMatch *regexp.Regexp = regexp.MustCompile(intLogPattern)

type LogLine struct {
	Seq int    `json:"seq"` // the line number
	Msg string `json:"msg"` // the line, without trailing \n
}

type Loggable struct {
	defaultSubject
	Uuid     string     `json:"uuid"` // Actually, this is the operation's uuid
	LogLines []*LogLine `json:"logLines"`
	Status   int        `json:"status"`
}

func (self *Loggable) AuthorizationName() string { return "log" }

// resembles a HAL type
type LoggableWrapper struct {
	Subject Loggable                     `json:"subject"`
	Links   map[string]map[string]string `json:"_links"`
}

const (
	LogSubscriptionSubscribe = iota
	LogSubscriptionUnsubscribe
)

type LogSubscription struct {
	Command       int
	OperationUuid string
	LineSeq       int
}

func NewLogLine(seq int, msg string) *LogLine {
	return &LogLine{Seq: seq, Msg: msg}
}

func LogLinesFromFile(file *os.File) []*LogLine {
	scanner := bufio.NewScanner(file)
	logLines := make([]*LogLine, 0)

	// lines are 1-indexed
	for lineSeq := 1; scanner.Scan(); lineSeq++ {
		logLine := NewLogLine(lineSeq, scanner.Text())
		logLines = append(logLines, logLine)
	}
	return logLines
}

func LogLinesFromSlice(slice []string, offset int) []*LogLine {
	logLines := make([]*LogLine, 0)

	for lineSeq, s := range slice {
		// lines are 1-indexed
		logLine := NewLogLine(lineSeq+1+offset, s)
		logLines = append(logLines, logLine)
	}
	return logLines
}

func (self *LogLine) IsInternal() bool {
	return !intLogMatch.Match([]byte(self.Msg))
}

func (self *LogLine) MustMarshal() []byte {
	j, err := json.Marshal(self)
	if err != nil {
		panic(err)
	}
	return j
}

func MustUnmarshalLogLine(j []byte) *LogLine {
	l := &LogLine{}
	err := json.Unmarshal(j, l)
	if err != nil {
		panic(err)
	}
	return l
}

func NewLoggable(uuid string, logLines []*LogLine, status int) *Loggable {
	return &Loggable{Uuid: uuid, LogLines: logLines, Status: status}
}

func (self *Loggable) Wrap() *LoggableWrapper {
	return &LoggableWrapper{
		Subject: *self,
		Links:   make(map[string]map[string]string),
	}
}

func (self *Loggable) MustMarshal() []byte {
	j, err := json.Marshal(self.Wrap())
	if err != nil {
		panic(err)
	}
	return j
}

func MustUnmarshalLoggable(j []byte) *Loggable {
	l := Loggable{}
	w := l.Wrap()
	err := json.Unmarshal(j, w)
	if err != nil {
		panic(err)
	}
	return &w.Subject
}

func (self *Loggable) OwnUrl(requestScheme, requestBaseUri string) string {
	return fmt.Sprintf("%s://%s/logs/%s", requestScheme, requestBaseUri, self.Uuid)
}

func (self *Loggable) Links(response map[string]map[string]string, requestScheme, requestBaseUri string) map[string]map[string]string {
	response["operation"] = map[string]string{"href": fmt.Sprintf("%s://%s/operations/%s", requestScheme, requestBaseUri, self.Uuid)}
	response["self"] = map[string]string{"href": self.OwnUrl(requestScheme, requestBaseUri)}
	return response
}
