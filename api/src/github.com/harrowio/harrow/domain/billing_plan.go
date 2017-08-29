package domain

import (
	"fmt"
)

const (
	FreePlanUuid     = "b99a21cc-b108-466e-aa4d-bde10ebbe1f3"
	PlatinumPlanUuid = "f975a385-3625-4883-b353-8f1febeb5b3e"
)

var (
	FreePlan = &BillingPlan{
		Uuid:                   FreePlanUuid,
		Name:                   "free",
		PrivateCodeAvailable:   true,
		PricePerMonth:          Money{0, USD},
		UsersIncluded:          1,
		ProjectsIncluded:       1,
		PricePerAdditionalUser: Money{0, USD},
		NumberOfConcurrentJobs: 1,
	}
	PlatinumPlan = &BillingPlan{
		Uuid:                   PlatinumPlanUuid,
		Name:                   "Platinum",
		PrivateCodeAvailable:   true,
		PricePerMonth:          Money{12900, USD},
		UsersIncluded:          10,
		ProjectsIncluded:       10,
		PricePerAdditionalUser: Money{1900, USD},
		NumberOfConcurrentJobs: 1,
	}
)

type BillingPlan struct {
	defaultSubject
	Uuid                   string `json:"uuid" db:"uuid"`
	Name                   string `json:"name" db:"name"`
	ProviderName           string `json:"providerName" db:"provider_name"`
	ProviderPlanId         string `json:"providerPlanId" db:"provider_plan_id"`
	PrivateCodeAvailable   bool   `json:"privateCodeAvailable" db:"private_code_available"`
	PricePerMonth          Money  `json:"pricePerMonth" db:"price_per_month"`
	UsersIncluded          int    `json:"usersIncluded" db:"users_included"`
	ProjectsIncluded       int    `json:"projectsIncluded" db:"projects_included"`
	PricePerAdditionalUser Money  `json:"pricePerAdditionalUser" db:"price_per_additional_user"`
	NumberOfConcurrentJobs int    `json:"numberOfConcurrentJobs" db:"number_of_concurrent_jobs"`
}

func NewBillingPlan(uuid string) *BillingPlan {
	return &BillingPlan{
		Uuid: uuid,
	}
}

func (self *BillingPlan) OwnUrl(requestScheme, requestBase string) string {
	return fmt.Sprintf("%s://%s/billing-plans/%s", requestScheme, requestBase, self.Uuid)
}

func (self *BillingPlan) Links(response map[string]map[string]string, requestScheme, requestBase string) map[string]map[string]string {
	response["self"] = map[string]string{"href": self.OwnUrl(requestScheme, requestBase)}
	return response
}

func (self *BillingPlan) EnsureDefaultPrice() {
	if self.PricePerMonth.Amount != 0 {
		return
	}

	self.PricePerMonth = Money{0, USD}

	if self.UsersIncluded >= 2 {
		self.PricePerMonth = Money{2900, USD}
	}
	if self.UsersIncluded >= 5 {
		self.PricePerMonth = Money{5900, USD}
	}
	if self.UsersIncluded >= 10 {
		self.PricePerMonth = Money{12900, USD}
	}
}

// UsersExceedingLimit returns the number of users by which members
// exceeds the limits imposed by this plan.
func (self *BillingPlan) UsersExceedingLimit(members int) int {
	if members > self.UsersIncluded {
		diff := members - self.UsersIncluded

		if diff < 0 {
			return 0
		}

		return diff
	}

	return 0
}

// ProjectsExceedingLimit returns the number of projects by which
// projects exceeds the limits imposed by this plan.
func (self *BillingPlan) ProjectsExceedingLimit(projects int) int {
	if projects > self.ProjectsIncluded {
		diff := projects - self.ProjectsIncluded
		if diff < 0 {
			return 0
		}

		return diff
	}

	return 0
}
