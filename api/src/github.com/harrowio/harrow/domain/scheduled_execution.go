package domain

import (
	"sort"
	"time"

	"github.com/gorhill/cronexpr"
)

const (
	ScheduledExecutionDefaultInterval = 24 * time.Hour
	ScheduledExecutionDefaultN        = 20
)

type ScheduledExecution struct {
	defaultSubject
	Time        time.Time `json:"time"`
	JobUuid     string    `json:"jobUuid"`
	Spec        string    `json:"spec"`
	Description string    `json:"description"`
}

func NewScheduledExecution(time time.Time, jobUuid, spec, description string) *ScheduledExecution {
	return &ScheduledExecution{
		Time:        time,
		JobUuid:     jobUuid,
		Spec:        spec,
		Description: description,
	}
}

func (self *ScheduledExecution) Embedded() map[string][]Subject {
	return self.embedded
}

func (self *ScheduledExecution) OwnUrl(requestScheme, requestBaseUri string) string {
	return ""
}

func (self *ScheduledExecution) Links(response map[string]map[string]string, requestScheme, requestBaseUri string) map[string]map[string]string {
	return nil
}

func (self *ScheduledExecution) AuthorizationName() string { return "schedule" }

type expr struct {
	cron        *cronexpr.Expression
	jobUuid     string
	spec        string
	description string
}

// sort.Interface
type ExecutionsByTime []*ScheduledExecution

func (a ExecutionsByTime) Len() int           { return len(a) }
func (a ExecutionsByTime) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ExecutionsByTime) Less(i, j int) bool { return a[i].Time.Before(a[j].Time) }

func ExecutionsBetween(from, to time.Time, n int, schedules []*Schedule) ([]*ScheduledExecution, error) {
	exprs := make([]expr, 0, len(schedules))
	var executions []*ScheduledExecution
	if n > 0 {
		executions = make([]*ScheduledExecution, 0, n)
	} else {
		executions = make([]*ScheduledExecution, 0)
	}
	// save the latest one time schedule that is within the given range (see below)
	var lastOneTime time.Time
	for _, schedule := range schedules {
		if schedule.Disabled != nil && len(*schedule.Disabled) > 0 {
			continue
		}
		if schedule.Cronspec != nil {
			cron, err := cronexpr.Parse(*schedule.Cronspec)
			if err != nil {
				return nil, err
			}
			exprs = append(exprs, expr{cron: cron, jobUuid: schedule.JobUuid, spec: *schedule.Cronspec, description: schedule.Description})
			continue
		}
		if schedule.Timespec != nil {
			err := schedule.setupOneTime()
			if err != nil {
				return nil, err
			}
			runOnceAt := *schedule.RunOnceAt
			if (from.Before(runOnceAt) || from.Equal(runOnceAt)) && (to.Equal(runOnceAt) || to.After(runOnceAt)) {
				executions = append(executions, NewScheduledExecution(runOnceAt, schedule.JobUuid, *schedule.Timespec, schedule.Description))
				if lastOneTime.Before(runOnceAt) {
					lastOneTime = runOnceAt
				}
			}
		}
	}
	for t := from; t.Equal(to) || t.Before(to); t = t.Add(1 * time.Minute) {
		for _, expr := range exprs {
			// need to subtract a small amount to find out if t itself is an execution
			exe := expr.cron.Next(t.Add(-1))
			// test if execution is between t and t+1m
			// but not after `to`
			if exe.Before(t.Add(1*time.Minute)) && exe.Before(to) {
				executions = append(executions, NewScheduledExecution(exe, expr.jobUuid, expr.spec, expr.description))
				// If we already have at least n executions *and* we are past the last one time schedule, exit early
				// The last check is important so we keep filling up the holes between the given one time schedules before
				// cutting to n.
				// Otherwise, we would just return one time schedules if n < len(one time schedules)
				if n > 0 && len(executions) >= n && t.After(lastOneTime) {
					sort.Sort(ExecutionsByTime(executions))
					return executions[:n], nil
				}
			}
		}
	}

	sort.Sort(ExecutionsByTime(executions))
	if n > 0 && n < len(executions) {
		return executions[:n], nil
	} else {
		return executions, nil
	}
}
