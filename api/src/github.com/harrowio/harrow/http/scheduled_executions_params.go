package http

import (
	"time"

	"github.com/harrowio/harrow/domain"
)

type ScheduledExecutionsParams struct {
	from time.Time
	to   time.Time
}

func NewScheduledExecutionsParams(params Getter) (*ScheduledExecutionsParams, error) {
	now := time.Now()
	self := &ScheduledExecutionsParams{
		from: now,
		to:   now.Add(domain.ScheduledExecutionDefaultInterval),
	}

	err := self.fromParams(params)

	return self, err
}

func (self *ScheduledExecutionsParams) fromParams(params Getter) error {
	if err := self.parseTime(&self.from, params.Get("from")); err != nil {
		return NewMalformedParameters("from", err)
	}

	if err := self.parseTime(&self.to, params.Get("to")); err != nil {
		return NewMalformedParameters("to", err)
	}

	return nil
}

func (self *ScheduledExecutionsParams) parseTime(dst *time.Time, src string) error {
	if len(src) == 0 {
		return nil
	}
	return dst.UnmarshalText([]byte(src))
}
