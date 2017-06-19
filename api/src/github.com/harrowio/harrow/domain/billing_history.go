package domain

import "time"

type BillingPlanSubscription struct {
	Id               string `json:"id"`
	PlanUuid         string `json:"planUuid"`
	OrganizationUuid string `json:"organizationUuid"`
	UserUuid         string `json:"userUuid"`
	Status           string `json:"status"`
}

func (self *BillingPlanSubscription) IsCancellable() bool {
	return self.PlanUuid != FreePlanUuid
}

func NewBillingPlanSubscription() *BillingPlanSubscription {
	return &BillingPlanSubscription{}
}

type BillingHistory struct {
	Subscriptions map[string]*BillingPlanSubscription `json:"subscriptions"`

	// OrganizationToPlanUuids maps organization uuids to the plan
	// uuid of the most recent subscription created for that
	// organization.
	OrganizationToPlanUuids map[string]string `json:"organizationToPlanUuids"`

	// OrganizationUuidsToSubscriptions maps organization uuids to
	// the currently active subscription for the organization.
	OrganizationUuidsToSubscriptions map[string]string

	// OrganizationUuidsToCreditCards maps organization uuids to
	// credit cards registered for that organization.
	OrganizationUuidsToCreditCards map[string][]*CreditCard

	// OrganizationUuidsToExtraProjects maps organization uuids to
	// increases in project limits an organization should be
	// granted.
	OrganizationUuidsToExtraProjects map[string]int

	// OrganizationUuidsToExtraProjects maps organization uuids to
	// increases in user limits an organization should be granted.
	OrganizationUuidsToExtraUsers map[string]int

	// OrganizationUuidsToExtras maps organization uuids to a
	// history of discounts that have been applied to an
	// organization.
	OrganizationUuidsToExtras map[string][]*BillingEvent

	// Version is the time of the last event that has been
	// processed by this instance.
	Version time.Time `json:"version"`
}

func NewBillingHistory() *BillingHistory {
	return &BillingHistory{
		Subscriptions:                    map[string]*BillingPlanSubscription{},
		OrganizationToPlanUuids:          map[string]string{},
		OrganizationUuidsToSubscriptions: map[string]string{},
		OrganizationUuidsToCreditCards:   map[string][]*CreditCard{},
		OrganizationUuidsToExtraUsers:    map[string]int{},
		OrganizationUuidsToExtraProjects: map[string]int{},
		OrganizationUuidsToExtras:        map[string][]*BillingEvent{},
	}
}

// Subscription returns the billing plan subscription for the given id
// or nil if no such subscription exists.
func (self *BillingHistory) Subscription(subscriptionId string) *BillingPlanSubscription {
	return self.Subscriptions[subscriptionId]
}

// HandleEvent ...
func (self *BillingHistory) HandleEvent(event *BillingEvent) {
	switch event.EventName {
	case "plan-selected":
		self.addSubscription(event)
	case "credit-card-added":
		self.addCreditCard(event)
	case "plan-subscription-changed":
		self.changeSubscription(event)
	case "extra-limits-granted":
		self.addLimits(event)
	}

	self.Version = event.OccurredOn
}

func (self *BillingHistory) addLimits(event *BillingEvent) {
	data := event.Data.(*BillingExtraLimitsGranted)
	if self.OrganizationUuidsToExtraProjects == nil {
		self.OrganizationUuidsToExtraProjects = map[string]int{}
	}

	if self.OrganizationUuidsToExtraUsers == nil {
		self.OrganizationUuidsToExtraUsers = map[string]int{}
	}

	if self.OrganizationUuidsToExtras == nil {
		self.OrganizationUuidsToExtras = map[string][]*BillingEvent{}
	}

	self.OrganizationUuidsToExtraUsers[event.OrganizationUuid] = data.Users

	self.OrganizationUuidsToExtraProjects[event.OrganizationUuid] = data.Projects

	self.OrganizationUuidsToExtras[event.OrganizationUuid] = append(self.OrganizationUuidsToExtras[event.OrganizationUuid], event)
}

func (self *BillingHistory) addSubscription(event *BillingEvent) {
	data := event.Data.(*BillingPlanSelected)
	if data.SubscriptionId == "" {
		return
	}

	subscription := NewBillingPlanSubscription()
	subscription.Id = data.SubscriptionId
	subscription.PlanUuid = data.PlanUuid
	subscription.OrganizationUuid = event.OrganizationUuid
	subscription.UserUuid = data.UserUuid
	subscription.Status = "active"
	self.Subscriptions[subscription.Id] = subscription
	self.OrganizationToPlanUuids[event.OrganizationUuid] = data.PlanUuid
	self.OrganizationUuidsToSubscriptions[event.OrganizationUuid] = data.SubscriptionId
}

func (self *BillingHistory) changeSubscription(event *BillingEvent) {
	data := event.Data.(*BillingPlanSubscriptionChanged)
	if data.Status == "canceled" {
		delete(self.Subscriptions, data.SubscriptionId)
		delete(self.OrganizationToPlanUuids, event.OrganizationUuid)
		delete(self.OrganizationUuidsToSubscriptions, event.OrganizationUuid)
	}
}

func (self *BillingHistory) addCreditCard(event *BillingEvent) {
	data := event.Data.(*BillingCreditCardAdded)
	creditCard := *data.CreditCard
	existingCards := self.OrganizationUuidsToCreditCards[event.OrganizationUuid]

	for _, existingCard := range existingCards {
		if existingCard.CardId == creditCard.CardId {
			return
		}
	}

	for _, existingCard := range existingCards {
		if creditCard.IsDefault {
			existingCard.IsDefault = false
		}
	}
	self.OrganizationUuidsToCreditCards[event.OrganizationUuid] = append(
		existingCards,
		&creditCard,
	)
}

func (self *BillingHistory) PlanUuidFor(organizationUuid string) string {
	return self.OrganizationToPlanUuids[organizationUuid]
}

func (self *BillingHistory) SubscriptionFor(organizationUuid string) *BillingPlanSubscription {
	subscriptionId, found := self.OrganizationUuidsToSubscriptions[organizationUuid]
	if !found {
		return nil
	}
	return self.Subscriptions[subscriptionId]
}

func (self *BillingHistory) CreditCardsFor(organizationUuid string) []*CreditCard {
	cards, found := self.OrganizationUuidsToCreditCards[organizationUuid]
	if found {
		return cards
	}
	return []*CreditCard{}
}

func (self *BillingHistory) ExtraProjectsFor(organizationUuid string) int {
	return self.OrganizationUuidsToExtraProjects[organizationUuid]
}

func (self *BillingHistory) ExtraUsersFor(organizationUuid string) int {
	return self.OrganizationUuidsToExtraUsers[organizationUuid]
}

func (self *BillingHistory) ExtrasGrantedTo(organizationUuid string) []*BillingEvent {
	events := self.OrganizationUuidsToExtras[organizationUuid]
	if events == nil {
		return []*BillingEvent{}
	}

	return events
}
