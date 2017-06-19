package clock

import "time"

var System = &SystemClock{}

type SystemClock struct{}

func (self *SystemClock) Now() time.Time { return time.Now() }
