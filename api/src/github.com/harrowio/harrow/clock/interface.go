package clock

import "time"

var Default Interface = System

type Interface interface {
	Now() time.Time
}
