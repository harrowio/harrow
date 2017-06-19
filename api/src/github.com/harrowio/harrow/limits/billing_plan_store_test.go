package limits

import "github.com/harrowio/harrow/domain"

type DummyBillingPlanStore struct {
	byId map[string]*domain.BillingPlan
}

func NewDummyBillingPlanStore() *DummyBillingPlanStore {
	return &DummyBillingPlanStore{
		byId: map[string]*domain.BillingPlan{},
	}
}

func (self *DummyBillingPlanStore) Add(plan *domain.BillingPlan) *DummyBillingPlanStore {
	self.byId[plan.Uuid] = plan
	return self
}

func (self *DummyBillingPlanStore) FindByUuid(billingPlanUuid string) (*domain.BillingPlan, error) {
	if plan, found := self.byId[billingPlanUuid]; found {
		return plan, nil
	} else {
		return nil, new(domain.NotFoundError)
	}
}
