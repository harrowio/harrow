package domain

import "time"

// An activity is something that has been done, most of times by a
// user, sometimes by a program. An activity represents that an action
// has lead to a result and captures information about the action and
// the result.
type Activity struct {
	// Id is a strictly monotonically increasing integer and
	// functions as a sequence number for ordering events.
	Id int `json:"id" db:"id"`

	// Name is used to disambiguate different activities.  Values
	// should follow the form:
	//
	//    ${OBJECT}.${EVENT}
	//
	// Examples:
	//
	//    operation.finished
	//    job.added
	Name string `json:"name" db:"name"`

	// OccurredOn is the time at which the activity has occurred.
	OccurredOn time.Time `json:"occurredOn" db:"occurred_on"`

	// CreatedAt is the time at which the activity has been
	// persisted.
	CreatedAt time.Time `json:"createdAt" db:"created_at"`

	// ContextUserUuid is the uuid of the user who caused the
	// activity.  For activities that are not triggered by a user,
	// this field is nil.
	ContextUserUuid *string `json:"contextUserUuid" db:"context_user_uuid"`

	// Payload is the context dependent data captured about the
	// activity.
	Payload interface{} `json:"payload" db:"payload"`

	// Extra is additional data added to the activity by various
	// processors of an activity.
	Extra map[string]interface{} `json:"extra" db:"extra"`
}

func NewActivity(id int, name string) *Activity {
	return &Activity{
		Id:    id,
		Name:  name,
		Extra: map[string]interface{}{},
	}
}

// Audience returns a list of user ids that should be notified about
// this activity.
//
// Usually it refers to all people having access to a project if the
// activity is emitted from something within a given project.
func (self *Activity) Audience() []string {
	audience := self.Extra["audience"]
	if audience == nil {
		return []string{}
	}

	result := []string{}

	switch members := audience.(type) {
	case []interface{}:
		for _, member := range members {
			result = append(result, member.(string))
		}
	case []string:
		result = members
	}

	return result
}

func (self *Activity) SetAudience(audience []string) *Activity {
	self.Extra["audience"] = audience
	return self
}

// SubscriptionKey returns the key used for matching notification
// subscriptions against this activity.  Usually this is the activity
// name.
func (self *Activity) SubscriptionKey() string {
	switch self.Name {
	case "operation.failed":
		return "operations.failed"
	case "operation.failed-fatally":
		return "operations.failed"
	case "operation.started":
		return "operations.started"
	case "operation.succeeded":
		return "operations.succeeded"
	case "operation.scheduled":
		return "operations.scheduled"
	default:
		return self.Name
	}
}

// ProjectUuid returns the uuid of the project this activity is
// associated with or the empty string if this activity is not
// associated with any project.
func (self *Activity) ProjectUuid() string {
	switch projectUuid := self.Extra["projectUuid"].(type) {
	case string:
		return projectUuid
	default:
		return ""
	}
}

func (self *Activity) SetProjectUuid(projectUuid string) *Activity {
	self.Extra["projectUuid"] = projectUuid
	return self
}

// JobUuid returns the uuid of the job this activity is
// associated with or the empty string if this activity is not
// associated with any job.
func (self *Activity) JobUuid() string {
	switch jobUuid := self.Extra["jobUuid"].(type) {
	case string:
		return jobUuid
	default:
		return ""
	}
}

func (self *Activity) SetJobUuid(jobUuid string) *Activity {
	self.Extra["jobUuid"] = jobUuid
	return self
}
