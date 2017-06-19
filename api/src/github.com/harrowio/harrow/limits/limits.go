package limits

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/harrowio/harrow/config"
	"github.com/harrowio/harrow/domain"
)

func CanCreate(thing interface{}) (bool, error) {
	return true, nil
}

type state struct {
	LastActivity                     time.Time
	OrganizationUuid                 string
	ProjectCount                     int
	ProjectsForUser                  map[string]map[string]string
	PrivateRepositories              map[string]bool
	PublicRepositories               map[string]bool
	NumberOfOperationsByYearAndMonth map[string]int
	Projects                         map[string]bool
	OrganizationCreatedAt            time.Time
}

type Limits struct {
	state state
	now   time.Time
}

func NewLimits(organizationUuid string, now time.Time) *Limits {
	return &Limits{
		state: state{
			LastActivity:                     time.Time{},
			OrganizationUuid:                 organizationUuid,
			ProjectCount:                     0,
			ProjectsForUser:                  map[string]map[string]string{},
			NumberOfOperationsByYearAndMonth: map[string]int{},
			PrivateRepositories:              map[string]bool{},
			PublicRepositories:               map[string]bool{},
			Projects:                         map[string]bool{},
		},
		now: now,
	}
}

// Version returns the occurrence date of the last activity handled by
// this instance.
func (self *Limits) Version() time.Time {
	return self.state.LastActivity
}

// NumberOfTrialDaysLeft returns the number of trial days left for
// this organization.  The default trial duration is 14 days.
func (self *Limits) NumberOfTrialDaysLeft() int {
	trialExpiresAt := self.state.OrganizationCreatedAt.Add(14 * 24 * time.Hour)
	daysRemaining := trialExpiresAt.Sub(self.now) / (24 * time.Hour)
	if trialExpiresAt.Before(self.now) {
		daysRemaining = 0
	}

	return int(daysRemaining)
}

// NumberOfMembers returns the number of members in the limits for
// this organization.
func (self *Limits) NumberOfMembers() int {
	return len(self.state.ProjectsForUser)
}

// NumberOfProjects returns the number of projects that have been
// created for this organization.
func (self *Limits) NumberOfProjects() int {
	return self.state.ProjectCount
}

// NumberOfSuccessfulOperationsIn returns the number of successful
// operations that have been run for an organization in the given year
// and month.
func (self *Limits) NumberOfSuccessfulOperationsIn(year int, month time.Month) int {
	return self.state.NumberOfOperationsByYearAndMonth[fmt.Sprintf("%04d-%02d", year, month)]
}

// NumberOfPrivateRepositories returns the number of private
// repositories used in the organization this instance accounts for.
//
// If the same repository is added with a different access mechanism
// (e.g. first SSH and then HTTPS), it is counted twice.  Likewise if
// the same repository is added in more than one project.
func (self *Limits) NumberOfPrivateRepositories() int {
	return len(self.state.PrivateRepositories)
}

// NumberOfPublicRepositories returns the number of public
// repositories used in the organization this instance accounts for.
func (self *Limits) NumberOfPublicRepositories() int {
	return len(self.state.PublicRepositories)
}

// Report returns a representation of this instance that is suitable
// for reporting it to the end user via the HTTP API.
//
// If plan is not nil, it is inspected to include the maximum values
// based on the plan.
func (self *Limits) Report(plan *domain.BillingPlan, extraUsers, extraProjects int) *domain.Limits {
	result := &domain.Limits{
		OrganizationUuid:    self.state.OrganizationUuid,
		Projects:            self.NumberOfProjects(),
		Members:             self.NumberOfMembers(),
		PublicRepositories:  self.NumberOfPublicRepositories(),
		PrivateRepositories: self.NumberOfPrivateRepositories(),
		TrialDaysLeft:       self.NumberOfTrialDaysLeft(),
		TrialEnabled:        config.GetConfig().FeaturesConfig().TrialEnabled,
		Version:             self.Version(),
	}

	if plan != nil {
		requiresUpgrade := self.NumberOfPrivateRepositories() > 0
		if plan.PrivateCodeAvailable {
			requiresUpgrade = false
		}

		result.Plan = &domain.LimitsComparedToPlan{
			UsersExceedingLimit:           plan.UsersExceedingLimit(self.NumberOfMembers() - extraUsers),
			ProjectsExceedingLimit:        plan.ProjectsExceedingLimit(self.NumberOfProjects() - extraProjects),
			RequiresUpgradeForPrivateCode: requiresUpgrade,
			UsersIncluded:                 plan.UsersIncluded + extraUsers,
			ProjectsIncluded:              plan.ProjectsIncluded + extraProjects,
		}

		if plan.Uuid == domain.PlatinumPlanUuid {
			result.Plan.ProjectsExceedingLimit = 0
		}
	}

	return result
}

// HandleActivity updates this instance of limits based on a single
// activity.
func (self *Limits) HandleActivity(activity *domain.Activity) error {
	if self.state.LastActivity.After(activity.OccurredOn) {
		return nil
	}

	handler := activityHandlers[activity.Name]
	if handler == nil {
		return nil
	}

	if err := handler(self, activity); err != nil {
		return err
	}
	self.state.LastActivity = activity.OccurredOn

	return nil
}

// addProjectForUser records the project identified by projectUuid as
// a project the user identified by userUuid is a member of.
//
// This is necessary for tracking when a user has left all projects in
// this organization and is thus not considered a member of the
// organization anymore.
func (self *Limits) addProjectForUser(userUuid, projectUuid string) *Limits {
	if self.state.ProjectsForUser[userUuid] == nil {
		self.state.ProjectsForUser[userUuid] = map[string]string{}
	}

	self.state.ProjectsForUser[userUuid][projectUuid] = projectUuid
	return self
}

// hasProject returns true if the project identified by projectUuid
// belongs to this organization.
func (self *Limits) hasProject(projectUuid string) bool {
	return self.state.Projects[projectUuid]
}

// addProject marks project as belonging to the organization this
// instance accounts for.
func (self *Limits) addProject(project *domain.Project) *Limits {
	self.state.Projects[project.Uuid] = true
	return self
}

func (self *Limits) MarshalJSON() ([]byte, error) {
	return json.Marshal(self.state)
}

func (self *Limits) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &self.state)
}
