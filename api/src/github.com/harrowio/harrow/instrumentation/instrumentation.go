package instrumentation

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/harrowio/harrow/bus/metrics"
)

var (
	bus metrics.Sink = metrics.Discard
)

func init() {
	bus = metrics.Discard
}

func Now() time.Time {
	return time.Now()
}

func TimeExecution(startedAt time.Time) {
	pc, _, _, _ := runtime.Caller(1)
	elapsed := time.Since(startedAt)
	elapsedMicro := float64(elapsed / time.Microsecond)
	function := runtime.FuncForPC(pc)
	file, line := function.FileLine(pc)
	name := function.Name()
	points := [2]*metrics.Point{
		metrics.NewPoint("instrumentation."+name).
			AddField("elapsed_micro", elapsedMicro).
			AddField("line", line).
			AddTag("file", file).
			At(time.Now()),
		metrics.NewPoint("instrumentation").
			AddField("elapsed_micro", elapsedMicro).
			AddTag("function", name).
			AddField("line", line).
			AddTag("file", file).
			At(time.Now()),
	}
	bus.Report(points[:])
}

func appendToFile(filename string, text string) {
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	fmt.Fprintf(f, text)
}

func WithMeasurementsTo(file string, tag string, do func()) {
	before := time.Now()
	appendToFile(file, "BEGIN "+tag+"\n")
	do()
	after := time.Now()
	message := fmt.Sprintf("END %s %.2fs\n", tag, after.Sub(before).Seconds())
	appendToFile(file, message)
}
