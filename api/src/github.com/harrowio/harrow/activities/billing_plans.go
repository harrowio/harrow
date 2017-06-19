package activities

import (
	"github.com/harrowio/harrow/domain"
)

func init() {
	registerPayload(BillingPlanPurchasedWithBraintree(&domain.BillingPlan{}))
}

func BillingPlanPurchasedWithBraintree(plan *domain.BillingPlan) *domain.Activity {
	return &domain.Activity{
		Name:       "billing-plan.purchased-with-braintree",
		OccurredOn: Clock.Now(),
		Payload:    plan,
		Extra:      map[string]interface{}{},
	}
}
