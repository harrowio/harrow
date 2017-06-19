package clock

import "time"

type Static struct {
	time.Time
}

func At(t time.Time) *Static {
	return &Static{t}
}

func (self *Static) Now() time.Time { return self.Time }
