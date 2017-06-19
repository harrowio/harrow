package logger

import (
	"io/ioutil"

	"github.com/rs/zerolog"
)

type Logger interface {
	Debug() *zerolog.Event
	Error() *zerolog.Event
	Fatal() *zerolog.Event
	Info() *zerolog.Event
	Panic() *zerolog.Event
	Warn() *zerolog.Event
}

var Discard Logger = zerolog.New(ioutil.Discard)
