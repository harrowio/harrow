package domain

import (
	"fmt"
	"strings"
	"time"
)

var (
	SegmentCold     = "cold"
	SegmentLimbo    = "limbo"
	SegmentReturned = "returned"
	SegmentFailed   = "failed"
	SegmentLoner    = "loner"
	SegmentIdeal    = "ideal"
	UserSegments    = []string{SegmentCold, SegmentLimbo, SegmentReturned, SegmentFailed, SegmentLoner, SegmentIdeal}
)

type Set struct {
	items map[string]bool
}

func NewSet() *Set { return &Set{items: map[string]bool{}} }

// Includes returns true if item is contained in this set.
func (self *Set) Includes(item string) bool {
	_, found := self.items[item]
	return found
}

// Add adds item to the set
func (self *Set) Add(item string) *Set {
	self.items[item] = true
	return self
}

// Size returns the number of items in the set
func (self *Set) Size() int { return len(self.items) }

// Intersect returns a new set containing the items that are part of
// this set and the other set.
func (self *Set) Intersect(other *Set) *Set {
	result := NewSet()

	for item := range self.items {
		if other.Includes(item) {
			result.Add(item)
		}
	}

	return result
}

type UserHistory struct {
	userUuid  string
	isActive  bool
	today     time.Time
	yesterday time.Time

	lastReportedAsActive time.Time
	returned             bool

	invitedAnotherUser bool

	jobsAdded *Set
	jobsRun   *Set

	previousSegment string

	verificationEmailsRequestedToday int
	emailAddressVerified             bool
	reportedAsActive                 int
}

func NewUserHistory(userUuid string, today time.Time) *UserHistory {
	return &UserHistory{
		userUuid:                         userUuid,
		isActive:                         false,
		today:                            today,
		yesterday:                        today.Add(-24 * time.Hour),
		verificationEmailsRequestedToday: 0,
		jobsAdded:                        NewSet(),
		jobsRun:                          NewSet(),
		invitedAnotherUser:               false,
		previousSegment:                  SegmentLimbo,
	}
}

// HandleActivity updates the history for this user.
func (self *UserHistory) HandleActivity(activity *Activity) error {
	if activity.ContextUserUuid == nil {
		return nil
	}

	if *activity.ContextUserUuid != self.userUuid {
		return nil
	}

	if strings.HasPrefix(activity.Name, "user.entered-segment-") {
		self.previousSegment = activity.Name[len("user.entered-segment-"):]
		return nil
	}

	switch activity.Name {
	case "invitation.created":
		self.invitedAnotherUser = true
	case "job.scheduled":
		schedule, ok := activity.Payload.(*Schedule)
		if !ok {
			return fmt.Errorf("Invalid activity payload type: want %T, got %T, activity id=%d\n", schedule, activity.Payload, activity.Id)
		}
		self.jobsRun.Add(schedule.JobUuid)
	case "job.added":
		job, ok := activity.Payload.(*Job)
		if !ok {
			return fmt.Errorf("Invalid activity payload type: want %T, got %T, activity id=%d\n", job, activity.Payload, activity.Id)
		}
		self.jobsAdded.Add(job.Uuid)
	case "user.email-verified":
		self.emailAddressVerified = true
	case "user.requested-verification-email":
		if activity.OccurredOn.After(self.yesterday) {
			self.verificationEmailsRequestedToday++
		}
	case "user.reported-as-active":
		self.reportedAsActive++

		if activity.OccurredOn.Sub(self.lastReportedAsActive) < 48*time.Hour {
			self.returned = true
		}
		self.lastReportedAsActive = activity.OccurredOn
		if activity.OccurredOn.After(self.yesterday) {
			self.isActive = true
		}
	}

	return nil
}

// VerificationEmailsRequestedToday returns the number verification
// emails that have been requested for this user today.
func (self *UserHistory) VerificationEmailsRequestedToday() int {
	return self.verificationEmailsRequestedToday
}

// IsActive returns whether this user has been reported as active
// today already.
func (self *UserHistory) IsActive() bool {
	return self.isActive
}

// Segment returns the KPI segment the user is in.  See the various
// Segment variables for which user falls into which segment.
func (self *UserHistory) Segment() string {

	if self.jobsAdded.Size() > 0 {
		if self.jobsAdded.Intersect(self.jobsRun).Size() == 0 {
			return SegmentFailed
		}

		if self.invitedAnotherUser {
			return SegmentIdeal
		}

		return SegmentLoner
	}

	if self.returned {
		return SegmentReturned
	}

	if self.reportedAsActive > 1 {
		return SegmentLimbo
	}

	if self.emailAddressVerified {
		return SegmentLimbo
	}

	return SegmentCold
}

// PreviousSegment returns the segment the user was in previously,
// according to segment-entered and segment-left activities.  If no
// such activities exist for a user, SegmentLimbo is returned.
func (self *UserHistory) PreviousSegment() string {
	return self.previousSegment
}
