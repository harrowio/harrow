package domain

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
	"time"
)

var (
	billingEventRegistryLock = &sync.Mutex{}
	registeredBillingEvents  = map[string]BillingEventData{}
)

func RegisterBillingEvent(event BillingEventData) {
	billingEventRegistryLock.Lock()
	defer billingEventRegistryLock.Unlock()
	registeredBillingEvents[event.BillingEventName()] = event
}

func NewBillingEventDataByName(eventName string) BillingEventData {
	example, found := registeredBillingEvents[eventName]
	if !found {
		return nil
	}

	return reflect.New(reflect.TypeOf(example).Elem()).Interface().(BillingEventData)
}

type BillingEvent struct {
	Uuid             string    `json:"uuid" db:"uuid"`
	EventName        string    `json:"eventName" db:"event_name"`
	OrganizationUuid string    `json:"organizationUuid" db:"organization_uuid"`
	OccurredOn       time.Time `json:"occurredOn" db:"occurred_on"`
	Data             BillingEventData
}

func NewBillingEvent(organizationUuid string, eventData BillingEventData) *BillingEvent {
	return &BillingEvent{
		OrganizationUuid: organizationUuid,
		EventName:        eventData.BillingEventName(),
		Data:             eventData,
	}
}

func (self *BillingEvent) UnmarshalJSON(data []byte) error {
	envelope := struct {
		Uuid             string
		EventName        string
		OrganizationUuid string
		OccurredOn       time.Time
		Data             json.RawMessage
	}{}

	if err := json.Unmarshal(data, &envelope); err != nil {
		return err
	}

	eventData := NewBillingEventDataByName(envelope.EventName)
	if data == nil {
		return fmt.Errorf("BillingEvent: unregistered event name %q", envelope.EventName)
	}

	if err := json.Unmarshal([]byte(envelope.Data), eventData); err != nil {
		return err
	}

	self.Uuid = envelope.Uuid
	self.EventName = envelope.EventName
	self.OrganizationUuid = envelope.OrganizationUuid
	self.OccurredOn = envelope.OccurredOn
	self.Data = eventData

	return nil
}

func init() {
	RegisterBillingEvent(&BillingPlanSelected{})
	RegisterBillingEvent(&BillingPlanSubscriptionChanged{})
	RegisterBillingEvent(&BillingCreditCardAdded{})
	RegisterBillingEvent(&BillingExtraLimitsGranted{})
}

type BillingEventData interface {
	BillingEventName() string
}

type BillingPlanSelected struct {
	UserUuid               string
	PlanUuid               string
	PlanName               string
	SubscriptionId         string
	PrivateCodeAvailable   bool
	PricePerMonth          Money
	UsersIncluded          int
	ProjectsIncluded       int
	PricePerAdditionalUser Money
	NumberOfConcurrentJobs int
}

func (self *BillingPlanSelected) BillingEventName() string { return "plan-selected" }
func (self *BillingPlanSelected) FillFromPlan(plan *BillingPlan) {
	self.PlanName = plan.Name
	self.PlanUuid = plan.Uuid
	self.PrivateCodeAvailable = plan.PrivateCodeAvailable
	self.PricePerMonth = plan.PricePerMonth
	self.UsersIncluded = plan.UsersIncluded
	self.ProjectsIncluded = plan.ProjectsIncluded
	self.PricePerAdditionalUser = plan.PricePerAdditionalUser
	self.NumberOfConcurrentJobs = plan.NumberOfConcurrentJobs
}

type BillingPlanSubscriptionChanged struct {
	UserUuid       string
	SubscriptionId string
	PlanId         string
	Status         string
}

func (self *BillingPlanSubscriptionChanged) BillingEventName() string {
	return "plan-subscription-changed"
}

type BillingCreditCardAdded struct {
	// CreditCard is the credit card that has been added.
	CreditCard *CreditCard

	// UserUuid is the uuid of the user who added the card.
	UserUuid string

	// PaymentProviderName is the name of the payment provider at
	// which the card has been registered.
	PaymentProviderName string
}

func (self *BillingCreditCardAdded) BillingEventName() string {
	return "credit-card-added"
}

type BillingExtraLimitsGranted struct {
	// Number of addtional projects granted to the organization
	Projects int

	// Number of addtional users granted to the organization
	Users int

	// Uuid of the user granting those limits
	GrantedBy string

	// Reason for granting those limits
	Reason string
}

func (self *BillingExtraLimitsGranted) BillingEventName() string {
	return "extra-limits-granted"
}
