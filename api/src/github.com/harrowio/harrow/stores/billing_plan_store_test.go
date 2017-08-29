package stores_test

import (
	"testing"

	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/stores"
	"github.com/harrowio/harrow/test_helpers"
	braintree "github.com/lionelbarrow/braintree-go"
)

type BraintreeMock struct {
	CallsToFindAllPlans int
	Plans               []*braintree.Plan
}

func (self *BraintreeMock) FindAllPlans() ([]*braintree.Plan, error) {
	self.CallsToFindAllPlans++
	return self.Plans, nil
}

func TestDbBillingPlanStore_FindAll_ReturnsAllPlansValidForTheCurrentDate(t *testing.T) {
	tx := test_helpers.GetDbTx(t)
	defer tx.Rollback()

	store := stores.NewDbBillingPlanStore(tx)

	tx.MustExec(`
INSERT INTO  provider_plan_availabilities_and_limits
       (name, provider_name, provider_plan_id, availability, limits)
 VALUES ('test', 'braintree', 'test-plan', '[,today)', '{}');
`)

	plans, err := store.FindAll()
	if err != nil {
		t.Fatal(err)
	}

	for _, plan := range plans {
		if plan.ProviderPlanId == "test-plan" {
			t.Fatalf("plan.ProviderPlanId = %q present; want not present", "test-plan")
		}
	}
}

func TestDbBillingPlanStore_FindAll_UnpacksLimitsFromJSON(t *testing.T) {
	tx := test_helpers.GetDbTx(t)
	defer tx.Rollback()

	store := stores.NewDbBillingPlanStore(tx)

	tx.MustExec(`
INSERT INTO  provider_plan_availabilities_and_limits
       (name, provider_name, provider_plan_id, availability, limits)
VALUES ('test', 'braintree', 'test-plan', '[today,]', '{"usersIncluded":100}');
`)

	plans, err := store.FindAll()
	if err != nil {
		t.Fatal(err)
	}

	found := (*domain.BillingPlan)(nil)
	for _, plan := range plans {
		if plan.ProviderPlanId == "test-plan" {
			found = plan
			break
		}

	}

	if found == nil {
		t.Fatalf("plan.ProviderPlanId = %q not found", "test-plan")
	}

	if got, want := found.UsersIncluded, 100; got != want {
		t.Errorf("found.UsersIncluded = %d; want %d", got, want)
	}
}
