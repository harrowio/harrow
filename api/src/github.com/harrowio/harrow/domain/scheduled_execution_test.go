package domain

import (
	"testing"
	"time"
)

func Test_ScheduledExecution_ExecutionsBetween(t *testing.T) {
	eachTwoSpec := "*/2 * * * *"
	eachTwo := &Schedule{Cronspec: &eachTwoSpec, Description: "each two"}
	eachThreeSpec := "*/3 * * * *"
	eachThree := &Schedule{Cronspec: &eachThreeSpec, Description: "each three"}
	onSevenSpec := "7 * * * *"
	onSeven := &Schedule{Cronspec: &onSevenSpec, Description: "on seven"}
	from := must3339("2000-01-01T00:00:05Z")
	to := must3339("2000-01-01T00:08:05Z")

	expected := []*ScheduledExecution{
		&ScheduledExecution{Time: must3339("2000-01-01T00:02:00Z"), Spec: eachTwoSpec, Description: "each two"},
		&ScheduledExecution{Time: must3339("2000-01-01T00:03:00Z"), Spec: eachThreeSpec, Description: "each three"},
		&ScheduledExecution{Time: must3339("2000-01-01T00:04:00Z"), Spec: eachTwoSpec, Description: "each two"},
		&ScheduledExecution{Time: must3339("2000-01-01T00:06:00Z"), Spec: eachTwoSpec, Description: "each two"},
		&ScheduledExecution{Time: must3339("2000-01-01T00:06:00Z"), Spec: eachThreeSpec, Description: "each three"},
		&ScheduledExecution{Time: must3339("2000-01-01T00:07:00Z"), Spec: onSevenSpec, Description: "on seven"},
		&ScheduledExecution{Time: must3339("2000-01-01T00:08:00Z"), Spec: eachTwoSpec, Description: "each two"},
	}

	executions, err := ExecutionsBetween(from, to, -1, []*Schedule{eachTwo, eachThree, onSeven})
	if err != nil {
		t.Fatal(err)
	}
	if len(expected) != len(executions) {
		t.Fatalf("Expected %d executions, but got %d\n", len(expected), len(executions))
	}
	for _, exp := range expected {
		if !hasExecution(executions, exp) {
			t.Fatalf("Expected ScheduledExecution(Time = %s, Spec = %s), but could not find it\n",
				exp.Time, exp.Spec)
		}
	}
	// the number 4 is interesting, because there would be two possible outcomes if the specs were not ordered
	fourExecutions, err := ExecutionsBetween(from, to, 4, []*Schedule{eachTwo, eachThree, onSeven})
	if 4 != len(fourExecutions) {
		t.Fatalf("Expected 4 executions, but got %d\n", len(fourExecutions))
	}
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 4; i++ {
		exp := expected[i]
		if !hasExecution(fourExecutions, exp) {
			t.Fatalf("Expected ScheduledExecution(Time = %s, Spec = %s), but could not find it\n",
				exp.Time, exp.Spec)
		}
	}
}

func Test_ScheduledExecution_ExecutionsBetween_Hourly(t *testing.T) {
	hourlySpec := "0 */1 * * *"
	from := must3339("2000-01-01T00:00:00Z")
	to := must3339("2000-01-02T00:00:00Z")
	schedules := []*Schedule{
		&Schedule{Cronspec: &hourlySpec, Description: "each hour"},
	}
	executions, err := ExecutionsBetween(from, to, -1, schedules)
	if err != nil {
		t.Fatal(err)
	}
	if len(executions) != 24 {
		t.Fatalf("len(executions)=%d, want %d", len(executions), 24)
	}
	expect := must3339("2000-01-01T23:00:00Z")
	if executions[23].Time != expect {
		t.Fatalf("executions[23].Time=%s, want %s", executions[23].Time, expect)
	}
}

func Test_ScheduledExecution_ExecutionsBetween_OneTimes(t *testing.T) {
	withinSpec := "0:05 Jan 01, 2000"
	anotherWithinSpec := "0:03 Jan 01, 2000"
	notWithinSpec := "0:05 Jan 03, 2000"

	from := must3339("2000-01-01T00:00:00Z")
	to := must3339("2000-01-02T00:00:00Z")
	schedules := []*Schedule{
		&Schedule{Timespec: &withinSpec, Description: "spec within"},
		&Schedule{Timespec: &anotherWithinSpec, Description: "another spec within"},
		// this Schedule should not appear in the output
		&Schedule{Timespec: &notWithinSpec, Description: "spec outside"},
	}
	executions, err := ExecutionsBetween(from, to, -1, schedules)
	if err != nil {
		t.Fatal(err)
	}
	if len(executions) != 2 {
		t.Fatalf("len(executions)=%d, want %d", len(executions), 2)
	}
	// due to sorting, anotherWithinSpec should be first
	expect := must3339("2000-01-01T00:03:00Z")
	if executions[0].Time != expect {
		t.Fatalf("executions[0].Time=%s, want %s", executions[0].Time.Format(time.RFC3339), expect.Format(time.RFC3339))
	}
	expect = must3339("2000-01-01T00:05:00Z")
	if executions[1].Time != expect {
		t.Fatalf("executions[0].Time=%s, want %s", executions[1].Time.Format(time.RFC3339), expect.Format(time.RFC3339))
	}
}

func Test_ScheduledExecution_ExecutionsBetween_OneTimes_WithN(t *testing.T) {
	withinSpec := "0:05 Jan 01, 2000"
	anotherWithinSpec := "0:03 Jan 01, 2000"
	notWithinSpec := "0:00 Jan 03, 2000"

	from := must3339("2000-01-01T00:00:00Z")
	to := must3339("2000-01-02T00:00:00Z")
	schedules := []*Schedule{
		&Schedule{Timespec: &withinSpec, Description: "spec within"},
		&Schedule{Timespec: &anotherWithinSpec, Description: "another spec within"},
		&Schedule{Timespec: &notWithinSpec, Description: "not within"},
	}
	// N=1
	executions, err := ExecutionsBetween(from, to, 1, schedules)
	if err != nil {
		t.Fatal(err)
	}
	if len(executions) != 1 {
		t.Fatalf("len(executions)=%d, want %d", len(executions), 1)
	}
	expect := must3339("2000-01-01T00:03:00Z")
	if executions[0].Time != expect {
		t.Fatalf("executions[0].Time=%s, want %s", executions[0].Time.Format(time.RFC3339), expect.Format(time.RFC3339))
	}
}

func Test_ScheduledExecution_ExecutionsBetween_Mixed_WithN(t *testing.T) {
	withinSpec := "0:05 Jan 01, 2000"
	anotherWithinSpec := "0:03 Jan 01, 2000"
	minutelySpec := "*/2 * * * *"
	notWithinSpec := "0:00 Jan 03, 2000"

	from := must3339("2000-01-01T00:00:00Z")
	to := must3339("2000-01-02T00:00:00Z")
	schedules := []*Schedule{
		&Schedule{Timespec: &withinSpec, Description: "spec within"},
		&Schedule{Cronspec: &minutelySpec, Description: "each other minute"},
		&Schedule{Timespec: &anotherWithinSpec, Description: "another spec within"},
		&Schedule{Timespec: &notWithinSpec, Description: "not within"},
	}
	executions, err := ExecutionsBetween(from, to, 5, schedules)
	if err != nil {
		t.Fatal(err)
	}
	if len(executions) != 5 {
		t.Fatalf("len(executions)=%d, want %d", len(executions), 5)
	}
	expects := []time.Time{
		must3339("2000-01-01T00:00:00Z"),
		must3339("2000-01-01T00:02:00Z"),
		must3339("2000-01-01T00:03:00Z"),
		must3339("2000-01-01T00:04:00Z"),
		must3339("2000-01-01T00:05:00Z"),
	}
	for i, expect := range expects {
		if executions[i].Time != expect {
			t.Fatalf("executions[%d].Time=%s, want %s", i, executions[i].Time.Format(time.RFC3339), expect.Format(time.RFC3339))
		}
	}
	// N=3, cutting off before the last onetime spec
	executions, err = ExecutionsBetween(from, to, 3, schedules)
	if err != nil {
		t.Fatal(err)
	}
	if len(executions) != 3 {
		t.Fatalf("len(executions)=%d, want %d", len(executions), 3)
	}
	expects = []time.Time{
		must3339("2000-01-01T00:00:00Z"),
		must3339("2000-01-01T00:02:00Z"),
		must3339("2000-01-01T00:03:00Z"),
	}

	for i, expect := range expects {
		if executions[i].Time != expect {
			t.Fatalf("executions[%d].Time=%s, want %s", i, executions[i].Time.Format(time.RFC3339), expect.Format(time.RFC3339))
		}
	}
}

func Test_ScheduledExecution_ExecutionsBetween_WithNGreaterThanMax(t *testing.T) {
	from := must3339("2000-01-01T00:00:00Z")
	to := must3339("2000-01-02T00:00:00Z")
	// give 0 schedules
	schedules := []*Schedule{}
	executions, err := ExecutionsBetween(from, to, 5, schedules)
	if err != nil {
		t.Fatal(err)
	}
	// expect 0 executions
	if len(executions) != 0 {
		t.Fatalf("len(executions)=%d, want %d", len(executions), 0)
	}
}

func must3339(value string) time.Time {
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		panic(err)
	}
	return t
}

func hasExecution(haystack []*ScheduledExecution, needle *ScheduledExecution) bool {
	for _, exe := range haystack {
		if exe.Spec == needle.Spec && exe.Time.Equal(needle.Time) && exe.Description == needle.Description {
			return true
		}
	}
	return false
}
