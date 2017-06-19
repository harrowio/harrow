package activities

import "github.com/harrowio/harrow/clock"

// Clock is the clock used for recording the time at which an activity
// has occurred.
var Clock clock.Interface = clock.System
