package stores

import (
	"encoding/json"
	"time"

	"net/http"

	"github.com/harrowio/harrow/config"
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/logger"
	"github.com/jmoiron/sqlx"
	braintree "github.com/lionelbarrow/braintree-go"
)

type Braintree interface {
	FindAllPlans() ([]*braintree.Plan, error)
}

type CachingBraintree struct {
	next      Braintree
	lastCheck time.Time
	cacheFor  time.Duration
	cached    []*braintree.Plan
	log       logger.Logger
}

func NewCachingBraintree(next Braintree, cacheFor time.Duration) *CachingBraintree {
	self := &CachingBraintree{
		next:      next,
		cacheFor:  cacheFor,
		lastCheck: time.Time{},
		cached:    nil,
	}

	return self
}

func (self *CachingBraintree) Log() logger.Logger {
	if self.log == nil {
		self.log = logger.Discard
	}
	return self.log
}

func (self *CachingBraintree) SetLogger(l logger.Logger) {
	self.log = l
}

func (self *CachingBraintree) FindAllPlans() ([]*braintree.Plan, error) {
	if self.isExpired() || self.cached == nil {
		plans, err := self.next.FindAllPlans()
		if err != nil {
			return nil, err
		}
		self.cached = plans
		self.lastCheck = time.Now()
	}

	return self.cached, nil
}

func (self *CachingBraintree) isExpired() bool {
	return time.Now().Add(-self.cacheFor).After(self.lastCheck)
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

type BraintreeProxy struct{}

func NewBraintreeProxy() *BraintreeProxy {
	return &BraintreeProxy{}
}

func (self *BraintreeProxy) FindAllPlans() ([]*braintree.Plan, error) {
	if !config.GetConfig().FeaturesConfig().LimitsEnabled {
		return []*braintree.Plan{}, nil
	}

	req, err := http.NewRequest("GET", "http://localhost:10002/", nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	result := []*braintree.Plan{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

type DbBillingPlanStore struct {
	tx        *sqlx.Tx
	braintree Braintree
}

func NewDbBillingPlanStore(tx *sqlx.Tx, braintree Braintree) *DbBillingPlanStore {
	return &DbBillingPlanStore{
		tx:        tx,
		braintree: braintree,
	}
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

	braintreePlansById := map[string]*braintree.Plan{}
	braintreePlans, err := self.braintree.FindAllPlans()
	if err != nil {
		return nil, err
	}
	for _, braintreePlan := range braintreePlans {
		braintreePlansById[braintreePlan.Id] = braintreePlan
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

		if braintreePlan, found := braintreePlansById[plan.ProviderPlanId]; found {
			plan.BindFromBraintree(braintreePlan)
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

func NewBillingPlanStore(tx *sqlx.Tx, c *config.Config) *DbBillingPlanStore {
	client := c.Braintree().NewClient()
	return NewDbBillingPlanStore(tx, NewCachingBraintree(NewBraintreeAPI(client), 1*time.Hour))
}
