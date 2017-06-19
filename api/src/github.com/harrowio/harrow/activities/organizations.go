package activities

import (
	"encoding/json"

	"github.com/harrowio/harrow/domain"
)

type OrganizationWithBillingPlan struct {
	Organization *domain.Organization
	BillingPlan  *domain.BillingPlan
}

func (self *OrganizationWithBillingPlan) UnmarshalJSON(data []byte) error {
	billingPlan := domain.BillingPlan{}
	json.Unmarshal(data, &billingPlan)
	if billingPlan.PricePerMonth.Currency.String() == "" {
		dest := &struct {
			Organization *domain.Organization
			BillingPlan  *domain.BillingPlan
		}{}
		if err := json.Unmarshal(data, dest); err != nil {
			return err
		}
		self.Organization = dest.Organization
		self.BillingPlan = dest.BillingPlan
		return nil
	} else {
		self.BillingPlan = &billingPlan
		return nil
	}
}

func init() {
	registerPayload(OrganizationCreated(&domain.Organization{}, nil))
	registerPayload(OrganizationArchived(&domain.Organization{}))
}

func OrganizationCreated(organization *domain.Organization, plan *domain.BillingPlan) *domain.Activity {
	return &domain.Activity{
		Name:  "organization.created",
		Extra: map[string]interface{}{},
		Payload: &OrganizationWithBillingPlan{
			Organization: organization,
			BillingPlan:  plan,
		},
		OccurredOn: Clock.Now(),
	}
}

func OrganizationArchived(organization *domain.Organization) *domain.Activity {
	return &domain.Activity{
		Name:       "organization.archived",
		Extra:      map[string]interface{}{},
		Payload:    organization,
		OccurredOn: Clock.Now(),
	}
}
