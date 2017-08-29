package stores

import (
	"encoding/json"

	"github.com/harrowio/harrow/domain"
	"github.com/jmoiron/sqlx"
	braintree "github.com/lionelbarrow/braintree-go"
)

type Braintree interface {
	FindAllPlans() ([]*braintree.Plan, error)
}

type BraintreeAPI struct {
	client *braintree.Braintree
}

func NewBraintreeAPI(client *braintree.Braintree) *BraintreeAPI {
	return &BraintreeAPI{
		client: client,
	}
}

func (self *BraintreeAPI) FindAllPlans() ([]*braintree.Plan, error) {
	return self.client.Plan().All()
}

type DbBillingPlanStore struct {
	tx *sqlx.Tx
}

func NewDbBillingPlanStore(tx *sqlx.Tx) *DbBillingPlanStore {
	return &DbBillingPlanStore{tx: tx}
}

type storedBillingPlan struct {
	*domain.BillingPlan
	PricePerAdditionalUser *domain.Money `db:"price_per_additional_user"`
	Limits                 []byte        `db:"limits"`
}

func (self *DbBillingPlanStore) FindAll() ([]*domain.BillingPlan, error) {
	envelopes := []*storedBillingPlan{}
	q := `SELECT uuid, name, provider_name, provider_plan_id, price_per_additional_user, limits FROM provider_plan_availabilities_and_limits WHERE availability @> now()::date`
	if err := self.tx.Select(&envelopes, q); err != nil {
		return nil, err
	}

	result := []*domain.BillingPlan{}
	for _, storedPlan := range envelopes {
		plan := storedPlan.BillingPlan
		if err := json.Unmarshal([]byte(storedPlan.Limits), &plan); err != nil {
			return nil, err
		}
		if storedPlan.PricePerAdditionalUser != nil {
			plan.PricePerAdditionalUser = *storedPlan.PricePerAdditionalUser
		}

		plan.EnsureDefaultPrice()

		result = append(result, plan)
	}

	return result, nil
}

func (self *DbBillingPlanStore) FindByUuid(uuid string) (*domain.BillingPlan, error) {
	all, err := self.FindAll()
	if err != nil {
		return nil, err
	}

	for _, plan := range all {
		if plan.Uuid == uuid {
			return plan, nil
		}
	}

	return nil, new(domain.NotFoundError)
}
