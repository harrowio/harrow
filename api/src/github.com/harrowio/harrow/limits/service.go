package limits

import (
	"github.com/harrowio/harrow/config"
	"github.com/harrowio/harrow/domain"
)

var (
	conf = config.GetConfig()
)

type OrganizationStore interface {
	FindByProjectUuid(projectUuid string) (*domain.Organization, error)
}

type BillingHistory interface {
	PlanUuidFor(organizationUuid string) string
	ExtraProjectsFor(organizationUuid string) int
	ExtraUsersFor(organizationUuid string) int
}

type BillingPlanStore interface {
	FindByUuid(billingPlanUuid string) (*domain.BillingPlan, error)
}

type LimitsStore interface {
	FindByOrganizationUuid(organizationUuid string) (*Limits, error)
}

type BelongsToProject interface {
	FindProject(projects domain.ProjectStore) (*domain.Project, error)
}

// Service takes care of coordinating the necessary objects to obtain
// an instance of limits for any domain subject.
type Service struct {
	organizations  OrganizationStore
	projects       domain.ProjectStore
	organization   *domain.Organization
	billingHistory BillingHistory
	billingPlans   BillingPlanStore
	billingPlan    *domain.BillingPlan
	limits         LimitsStore
}

func NewService(organizations OrganizationStore, projects domain.ProjectStore, billingPlans BillingPlanStore, billingHistory BillingHistory, limits LimitsStore) *Service {
	return &Service{
		organizations:  organizations,
		projects:       projects,
		billingPlans:   billingPlans,
		billingHistory: billingHistory,
		limits:         limits,
	}
}

// Exceeded returns true if the limits have been exceeded for the
// provided object.
func (self *Service) Exceeded(subject interface{}) (bool, error) {
	if !conf.FeaturesConfig().LimitsEnabled && conf.Environment() != "test" {
		return false, nil
	}

	if organization, ok := subject.(*domain.Organization); ok {
		self.organization = organization
	} else {
		if err := self.findOrganization(subject); err != nil {
			return false, err
		}
	}

	planUuid := self.billingHistory.PlanUuidFor(self.organization.Uuid)
	plan, err := self.billingPlans.FindByUuid(planUuid)
	if err != nil {
		return false, err
	} else {
		self.billingPlan = plan
	}

	limits, err := self.limits.FindByOrganizationUuid(self.organization.Uuid)
	if err != nil {
		return false, err
	}
	reported := limits.Report(plan,
		self.billingHistory.ExtraUsersFor(self.organization.Uuid),
		self.billingHistory.ExtraProjectsFor(self.organization.Uuid),
	)
	return reported.Exceeded(), nil
}

func (self *Service) findOrganization(subject interface{}) error {

	belongsToProject, ok := subject.(BelongsToProject)
	if !ok {
		return nil
	}

	project, err := belongsToProject.FindProject(self.projects)
	if err != nil {
		return err
	}

	organization, err := self.organizations.FindByProjectUuid(project.Uuid)
	if err != nil {
		return err
	}

	self.organization = organization

	return nil
}

// Organization returns the organization for which this instance
// determines and applies limits.
func (self *Service) Organization() *domain.Organization {
	return self.organization
}

// BillingPlan returns the billing plan for the organization for which
// this instance determines and applies limits.
func (self *Service) BillingPlan() *domain.BillingPlan {
	return self.billingPlan
}
