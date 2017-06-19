package domain

import (
	"errors"
	"fmt"

	"time"

	"github.com/dhamidi/timespec"
	"github.com/gorhill/cronexpr"
	"github.com/harrowio/harrow/clock"
)

var Clock clock.Interface = clock.System

type Schedulable interface {
	// NextOperation determines the next time an operation should
	// run.  If now is non nil, the value it points to is used
	// instead of time.Now().  A zero time return value means, that no
	// next operation is scheduled.  An error is returned when
	// invalid scheduling information is used for determining the
	// time of the next operation.
	//
	// The time returned is in the timezone specified by this
	// Schedulable.
	NextOperation(now *time.Time) (time.Time, error)

	// Validate tests whether the schedule is in valid state.  If it
	// isn't, it returns an error of type (*ValidationError).
	// Validate may modify the internal state of the schedule to
	// pool, make it valid (if possible).
	Validate() error

	// Id returns the UUID of the schedule.  This information is
	// required for managing the pool of live schedules.
	Id() string

	// JobId returns the UUID of the job to execute by this
	// Schedulable.
	JobId() string

	// IsRecurring is used by the Scheduler (and possibly others)
	// to determine whether to add this to the watch pool
	// or schedule it for immediate execution
	IsRecurring() bool

	// IsDisabled returns true or false based on the presense of a non
	// empty string in Disabled
	IsDisabled() bool

	OperationParameters() *OperationParameters

	// We echo the interface of domain.Subject in order to
	// be able to render Schedulables to the API as JSON
	// Url returns the URL referencing the subject itself
	OwnUrl(requestScheme, requestBaseUri string) string
	// Links fills in the _links object in the HAL response
	Links(response map[string]map[string]string, requestScheme, requestBaseUri string) map[string]map[string]string
}

type Schedule struct {
	defaultSubject
	Uuid         string    `json:"uuid"`
	UserUuid     string    `json:"userUuid"            db:"user_uuid"`
	JobUuid      string    `json:"jobUuid"             db:"job_uuid"`
	Cronspec     *string   `json:"cronspec"`
	Description  string    `json:"description"`
	CreatedAt    time.Time `json:"createdAt"           db:"created_at"`
	Timespec     *string   `json:"timespec,omitempty"`
	TimezoneName string    `json:"location"            db:"location"`

	Parameters *OperationParameters `json:"parameters" db:"parameters"`

	NextExecutions []*ScheduledExecution `json:"nextExecutions,omitempty", db:"-"`

	ArchivedAt *time.Time `json:"archivedAt" db:"archived_at"`

	// Disabled is an enum describing the reason this Schedule is disabled
	// If it is nil, the Schedule is enabled.
	Disabled *string `json:"disabled" db:"disabled"`
	// DisabledBecause describes the error in case of ScheduleDisabledInternalError
	DisabledBecause *string `json:"-" db:"disabled_because"`
	// location contains the result of loading the timezone
	// information identified by TimezoneName.
	location *time.Location
	// RunOnceAt contains the result of parsing a timespec by at(1).
	// It is necessary to cache this value in the database, because
	// evaluating a timespec by at(1) can depend on the current time.
	// Repeatedly evaluating 'now + 2 minutes' for example yields a
	// different time everytime it is evaluated.
	RunOnceAt *time.Time `json:"runOnceAt,omitempty"   db:"run_once_at"`
}

type recurringSchedule Schedule
type oneTimeSchedule Schedule

var (
	ErrInvalidCronspecFormat = fmt.Errorf("Invalid Cronspec Format.")
	ErrInvalidTimespecFormat = fmt.Errorf("Invalid Timespec Format.")
	ErrScheduledInThePast    = fmt.Errorf("Scheduled in the past.")
	ErrInvalidTimezoneName   = fmt.Errorf("Invalid timezone name.")
	ErrNoTimeScheduled       = fmt.Errorf("Empty schedule time (Cronspec/Timespec).")
)

var (
	ScheduleDisabledInternalError = "internal_error"
	ScheduleDisabledJobArchived   = "job_archived"
	ScheduleDisabledRanOnce       = "ran_once"
)

func (s *recurringSchedule) IsDisabled() bool {
	if s.Disabled == nil {
		return false
	}

	if s.ArchivedAt != nil {
		return true
	}

	return len(*s.Disabled) > 0
}

func (s *oneTimeSchedule) IsDisabled() bool {
	if s.Disabled == nil {
		return false
	}

	if s.ArchivedAt != nil {
		return true
	}

	return len(*s.Disabled) > 0
}

func (s *recurringSchedule) IsRecurring() bool {
	return true
}

func (s *recurringSchedule) NextOperation(now *time.Time) (time.Time, error) {
	var currentTime time.Time
	if now == nil {
		currentTime = time.Now()
	} else {
		currentTime = *now
	}

	if s.Cronspec == nil {
		return time.Time{}, ErrNoTimeScheduled
	}

	expr, err := cronexpr.Parse(*s.Cronspec)

	if err != nil {
		return time.Time{}, fmt.Errorf("Invalid cronspec: %s", *s.Cronspec)
	}

	return expr.Next(currentTime.In(s.location)), nil
}

func (s *recurringSchedule) Validate() error {
	if s.Cronspec == nil {
		return NewValidationError("cronspec", "empty")
	}

	_, err := cronexpr.Parse(*s.Cronspec)
	if err != nil {
		return NewValidationError("cronspec", "invalid")
	}

	return nil
}

func (s *recurringSchedule) Id() string {
	return s.Uuid
}

func (s *recurringSchedule) JobId() string {
	return s.JobUuid
}

func (s *oneTimeSchedule) IsRecurring() bool {
	return false
}

func (s *oneTimeSchedule) NextOperation(now *time.Time) (time.Time, error) {
	var currentTime time.Time
	if now == nil {
		currentTime = Clock.Now()
	} else {
		currentTime = *now
	}

	if s.RunOnceAt == nil {
		return time.Time{}, ErrNoTimeScheduled
	} else if currentTime.UTC().After((*s.RunOnceAt).UTC()) {
		return *s.RunOnceAt, ErrScheduledInThePast
	}

	return (*s.RunOnceAt).In(s.location), nil
}

func resolveTimespec(tsStr string, referenceTime time.Time) (*time.Time, error) {
	if ts, err := timespec.Parse(tsStr); err != nil {
		return nil, err
	} else {
		t := ts.Resolve(Clock.Now())
		return &t, nil
	}
}

func (s *oneTimeSchedule) Validate() error {
	if s.Timespec == nil {
		return NewValidationError("timespec", "required")
	}

	_, err := resolveTimespec(*s.Timespec, s.CreatedAt)
	if err != nil {
		return NewValidationError("timespec", "invalid")
	}

	return nil
}

func (s *oneTimeSchedule) Id() string {
	return s.Uuid
}

func (s *oneTimeSchedule) JobId() string {
	return s.JobUuid
}

func (s *Schedule) setupOneTime() error {
	if s.RunOnceAt == nil {
		t, err := resolveTimespec(*s.Timespec, s.CreatedAt)
		if err != nil {
			return err
		}
		s.RunOnceAt = t
	}
	return nil
}

// TODO: This actually changes s via setupOneTime, but the name does not give
// that fact away. This method should probably be defined on a non-pointer,
// make a copy of s, or RunOnceAt should be a method so we don't need
// setupOneTime.
func NewSchedulable(s *Schedule) (Schedulable, error) {

	var err error
	s.location, err = time.LoadLocation(s.TimezoneName)
	if err != nil {
		return nil, ErrInvalidTimezoneName
	}

	if s.Disabled != nil {
		return nil, errors.New(*s.Disabled)
	}

	if s.Cronspec != nil {
		return (*recurringSchedule)(s), err
	}

	if s.Timespec != nil {
		err := s.setupOneTime()
		return (*oneTimeSchedule)(s), err
	}

	return nil, ErrNoTimeScheduled
}

// TODO: this potentially mutates self, see above
func (self *Schedule) Validate() error {
	schedulable, _ := NewSchedulable(self)
	return schedulable.Validate()
}

func (self *Schedule) OwnUrl(requestScheme, requestBaseUri string) string {
	return fmt.Sprintf("%s://%s/schedules/%s", requestScheme, requestBaseUri, self.Uuid)
}

func (self *oneTimeSchedule) OwnUrl(requestScheme, requestBaseUri string) string {
	return fmt.Sprintf("%s://%s/schedules/%s", requestScheme, requestBaseUri, self.Uuid)
}

func (self *recurringSchedule) OwnUrl(requestScheme, requestBaseUri string) string {
	return fmt.Sprintf("%s://%s/schedules/%s", requestScheme, requestBaseUri, self.Uuid)
}

func (self *Schedule) Links(response map[string]map[string]string, requestScheme, requestBaseUri string) map[string]map[string]string {
	response["self"] = map[string]string{"href": self.OwnUrl(requestScheme, requestBaseUri)}
	response["job"] = map[string]string{"href": fmt.Sprintf("%s://%s/jobs/%s", requestScheme, requestBaseUri, self.JobUuid)}
	response["user"] = map[string]string{"href": fmt.Sprintf("%s://%s/users/%s", requestScheme, requestBaseUri, self.UserUuid)}
	response["scheduled-executions"] = map[string]string{
		"href": fmt.Sprintf("%s://%s/schedules/%s/scheduled-executions", requestScheme, requestBaseUri, self.Uuid),
	}
	response["operations"] = map[string]string{"href": fmt.Sprintf("%s://%s/schedules/%s/operations", requestScheme, requestBaseUri, self.Uuid)}
	return response
}

func (self *oneTimeSchedule) Links(response map[string]map[string]string, requestScheme, requestBaseUri string) map[string]map[string]string {
	response["self"] = map[string]string{"href": self.OwnUrl(requestScheme, requestBaseUri)}
	response["job"] = map[string]string{"href": fmt.Sprintf("%s://%s/jobs/%s", requestScheme, requestBaseUri, self.JobUuid)}
	response["user"] = map[string]string{"href": fmt.Sprintf("%s://%s/users/%s", requestScheme, requestBaseUri, self.UserUuid)}
	response["scheduled-executions"] = map[string]string{
		"href": fmt.Sprintf("%s://%s/schedules/%s/scheduled-executions", requestScheme, requestBaseUri, self.Uuid),
	}
	response["operations"] = map[string]string{"href": fmt.Sprintf("%s://%s/schedules/%s/operations", requestScheme, requestBaseUri, self.Uuid)}
	return response
}

func (self *recurringSchedule) Links(response map[string]map[string]string, requestScheme, requestBaseUri string) map[string]map[string]string {
	response["self"] = map[string]string{"href": self.OwnUrl(requestScheme, requestBaseUri)}
	response["job"] = map[string]string{"href": fmt.Sprintf("%s://%s/jobs/%s", requestScheme, requestBaseUri, self.JobUuid)}
	response["user"] = map[string]string{"href": fmt.Sprintf("%s://%s/users/%s", requestScheme, requestBaseUri, self.UserUuid)}
	response["scheduled-executions"] = map[string]string{
		"href": fmt.Sprintf("%s://%s/schedules/%s/scheduled-executions", requestScheme, requestBaseUri, self.Uuid),
	}
	response["operations"] = map[string]string{"href": fmt.Sprintf("%s://%s/schedules/%s/operations", requestScheme, requestBaseUri, self.Uuid)}
	return response
}

// FindProject satisfies authz.BelongsToProject in order to determine
// authorization.
func (self *Schedule) FindProject(store ProjectStore) (*Project, error) {
	return store.FindByJobUuid(self.JobUuid)
}

func (self *Schedule) OwnedBy(user *User) bool {
	return user.Uuid == self.UserUuid
}

// FindUser satisies authz.BelongsToUser in order to determine
// authorization.
func (self *Schedule) FindUser(store UserStore) (*User, error) {
	return store.FindByUuid(self.UserUuid)
}

func (self *recurringSchedule) FindProject(store ProjectStore) (*Project, error) {
	return store.FindByJobUuid(self.JobUuid)
}

func (self *oneTimeSchedule) FindProject(store ProjectStore) (*Project, error) {
	return store.FindByJobUuid(self.JobUuid)
}

func (self *Schedule) AuthorizationName() string { return "schedule" }

func (self *Schedule) OperationParameters() *OperationParameters {
	self.Parameters.ScheduleDescription = self.Description
	self.Parameters.ScheduleUuid = self.Uuid
	return self.Parameters
}

func (self *oneTimeSchedule) OperationParameters() *OperationParameters {
	if self.Parameters == nil {
		self.Parameters = NewOperationParameters()
	}

	self.Parameters.ScheduleDescription = self.Description
	self.Parameters.ScheduleUuid = self.Uuid
	if self.Parameters.Reason != OperationTriggerReason("") {
		return self.Parameters
	}

	if self.Timespec != nil && *self.Timespec == "now" {
		self.Parameters.Reason = OperationTriggeredByUser
	} else {
		self.Parameters.Reason = OperationTriggeredBySchedule
	}
	return self.Parameters
}

func (self *recurringSchedule) OperationParameters() *OperationParameters {
	if self.Parameters == nil {
		self.Parameters = NewOperationParameters()
	}
	self.Parameters.ScheduleDescription = self.Description
	self.Parameters.ScheduleUuid = self.Uuid
	self.Parameters.Reason = OperationTriggeredBySchedule
	return self.Parameters
}
