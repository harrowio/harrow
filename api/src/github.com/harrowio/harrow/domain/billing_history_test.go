package domain

import (
	"reflect"
	"testing"
	"time"
)

func TestBillingHistory_HandleEvent_PlanSelected_RecordsUserAndOrganizationBySubscriptionId(t *testing.T) {
	history := NewBillingHistory()
	planUuid := "64683135-b45f-412b-a221-7c1a18baf87e"
	userId := "2d0b1e8e-d3b6-496c-9d39-0c2e8705ece1"
	organizationId := "7f3d461d-e206-4cad-ac51-d7f39056b6fa"
	subscriptionId := "79mh9g"
	planSelected := &BillingEvent{
		EventName:        "plan-selected",
		OrganizationUuid: organizationId,
		Data: &BillingPlanSelected{
			UserUuid:       userId,
			PlanUuid:       planUuid,
			SubscriptionId: subscriptionId,
		},
	}
	history.HandleEvent(planSelected)

	subscription := &BillingPlanSubscription{
		UserUuid:         userId,
		Id:               subscriptionId,
		PlanUuid:         planUuid,
		OrganizationUuid: organizationId,
		Status:           "active",
	}

	if got, want := history.Subscription(subscriptionId), subscription; !reflect.DeepEqual(got, want) {
		t.Errorf(`history.Subscription(subscriptionId) = %v; want %v`, got, want)
	}
}

func TestBillingHistory_HandleEvent_updatesVersionToEventOccurrenceTime(t *testing.T) {
	history := NewBillingHistory()
	planUuid := "64683135-b45f-412b-a221-7c1a18baf87e"
	userId := "2d0b1e8e-d3b6-496c-9d39-0c2e8705ece1"
	organizationId := "7f3d461d-e206-4cad-ac51-d7f39056b6fa"
	subscriptionId := "79mh9g"
	occurredOn := time.Date(2015, 12, 11, 12, 0, 0, 0, time.UTC)
	planSelected := &BillingEvent{
		EventName:        "plan-selected",
		OccurredOn:       occurredOn,
		OrganizationUuid: organizationId,
		Data: &BillingPlanSelected{
			UserUuid:       userId,
			PlanUuid:       planUuid,
			SubscriptionId: subscriptionId,
		},
	}
	history.HandleEvent(planSelected)

	if got, want := history.Version, occurredOn; !got.Equal(want) {
		t.Errorf(`history.Version = %v; want %v`, got, want)
	}
}

func TestBillingHistory_HandleEvent_PlanSelected_recordsCurrentPlanForOrganization(t *testing.T) {
	history := NewBillingHistory()
	planUuid := "64683135-b45f-412b-a221-7c1a18baf87e"
	userId := "2d0b1e8e-d3b6-496c-9d39-0c2e8705ece1"
	organizationId := "7f3d461d-e206-4cad-ac51-d7f39056b6fa"
	subscriptionId := "79mh9g"
	planSelected := &BillingEvent{
		EventName:        "plan-selected",
		OrganizationUuid: organizationId,
		Data: &BillingPlanSelected{
			UserUuid:       userId,
			PlanUuid:       planUuid,
			SubscriptionId: subscriptionId,
		},
	}
	history.HandleEvent(planSelected)

	if got, want := history.PlanUuidFor(organizationId), planUuid; got != want {
		t.Errorf(`history.PlanUuidFor(organizationId) = %v; want %v`, got, want)
	}
}

func TestBillingHistory_HandleEvent_PlanSelected_doesNotRecordAnythingIfNoSubscriptionIdIsProvided(t *testing.T) {
	history := NewBillingHistory()
	planUuid := "64683135-b45f-412b-a221-7c1a18baf87e"
	userId := "2d0b1e8e-d3b6-496c-9d39-0c2e8705ece1"
	organizationId := "7f3d461d-e206-4cad-ac51-d7f39056b6fa"
	planSelected := &BillingEvent{
		EventName:        "plan-selected",
		OrganizationUuid: organizationId,
		Data: &BillingPlanSelected{
			UserUuid: userId,
			PlanUuid: planUuid,
		},
	}
	history.HandleEvent(planSelected)

	if got, want := history.SubscriptionFor(organizationId), (*BillingPlanSubscription)(nil); got != want {
		t.Errorf(`history.SubscriptionFor(organizationUuid) = %v; want %v`, got, want)
	}

	if got, want := history.PlanUuidFor(organizationId), ""; got != want {
		t.Errorf(`history.PlanUuidFor(organizationId) = %v; want %v`, got, want)
	}
}

func TestBillingHistory_HandleEvent_PlanSelected_recordsCurrentSubscriptionForOrganization(t *testing.T) {
	history := NewBillingHistory()
	planUuid := "64683135-b45f-412b-a221-7c1a18baf87e"
	userId := "2d0b1e8e-d3b6-496c-9d39-0c2e8705ece1"
	organizationId := "7f3d461d-e206-4cad-ac51-d7f39056b6fa"
	subscriptionId := "79mh9g"
	planSelected := &BillingEvent{
		EventName:        "plan-selected",
		OrganizationUuid: organizationId,
		Data: &BillingPlanSelected{
			UserUuid:       userId,
			PlanUuid:       planUuid,
			SubscriptionId: subscriptionId,
		},
	}
	history.HandleEvent(planSelected)

	subscription := history.SubscriptionFor(organizationId)
	if got := subscription; got == nil {
		t.Fatalf(`subscription is nil`)
	}
}

func TestBillingHistory_HandleEvent_PlanSubscriptionChanged(t *testing.T) {
	history := NewBillingHistory()
	planUuid := "64683135-b45f-412b-a221-7c1a18baf87e"
	userId := "2d0b1e8e-d3b6-496c-9d39-0c2e8705ece1"
	organizationId := "7f3d461d-e206-4cad-ac51-d7f39056b6fa"
	subscriptionId := "79mh9g"
	planSelected := &BillingEvent{
		EventName:        "plan-selected",
		OrganizationUuid: organizationId,
		Data: &BillingPlanSelected{
			UserUuid:       userId,
			PlanUuid:       planUuid,
			SubscriptionId: subscriptionId,
		},
	}
	planCancelled := &BillingEvent{
		EventName:        "plan-subscription-changed",
		OrganizationUuid: organizationId,
		Data: &BillingPlanSubscriptionChanged{
			UserUuid:       userId,
			PlanId:         planUuid,
			SubscriptionId: subscriptionId,
			Status:         "canceled",
		},
	}
	history.HandleEvent(planSelected)
	history.HandleEvent(planCancelled)

	subscription := history.SubscriptionFor(organizationId)
	if got, want := subscription, (*BillingPlanSubscription)(nil); got != want {
		t.Errorf(`subscription = %v; want %v`, got, want)
	}
}

func newExampleCreditCard(id string) *CreditCard {
	return &CreditCard{
		CardholderName: "John Doe",
		CardId:         id,
		IsDefault:      true,
	}
}

func TestBillingHistory_HandleEvent_CreditCardAdded_tracksAllCreditCardsByOrganizationUuid(t *testing.T) {
	history := NewBillingHistory()
	testCards := []*CreditCard{
		newExampleCreditCard("card-1"),
		newExampleCreditCard("card-2"),
	}

	userId := "a902d519-f577-4543-b853-bee183e5b403"
	organizationId := "a19737e3-d2f4-48f0-9dfc-ee035833c111"
	for _, creditCard := range testCards {
		creditCardAdded := &BillingEvent{
			EventName:        "credit-card-added",
			OrganizationUuid: organizationId,
			Data: &BillingCreditCardAdded{
				UserUuid:   userId,
				CreditCard: creditCard,
			},
		}
		history.HandleEvent(creditCardAdded)
	}

	creditCards := history.CreditCardsFor(organizationId)
	if got, want := len(creditCards), 2; got != want {
		t.Fatalf(`len(creditCards) = %v; want %v`, got, want)
	}

	if got, want := creditCards[0].CardId, "card-1"; got != want {
		t.Errorf(`creditCards[0].CardId = %v; want %v`, got, want)
	}

	if got, want := creditCards[1].CardId, "card-2"; got != want {
		t.Errorf(`creditCards[1].CardId = %v; want %v`, got, want)
	}

}

func TestBillingHistory_HandleEvent_CreditCardAdded_marksOnlyTheMostRecentlyAddedCardAsDefault(t *testing.T) {
	history := NewBillingHistory()
	testCards := []*CreditCard{
		newExampleCreditCard("card-1"),
		newExampleCreditCard("card-2"),
	}

	userId := "a902d519-f577-4543-b853-bee183e5b403"
	organizationId := "a19737e3-d2f4-48f0-9dfc-ee035833c111"
	for _, creditCard := range testCards {
		creditCardAdded := &BillingEvent{
			EventName:        "credit-card-added",
			OrganizationUuid: organizationId,
			Data: &BillingCreditCardAdded{
				UserUuid:   userId,
				CreditCard: creditCard,
			},
		}
		history.HandleEvent(creditCardAdded)
	}

	creditCards := history.CreditCardsFor(organizationId)
	if got, want := len(creditCards), 2; got != want {
		t.Fatalf(`len(creditCards) = %v; want %v`, got, want)
	}

	if got, want := creditCards[0].IsDefault, false; got != want {
		t.Errorf(`creditCards[0].IsDefault = %v; want %v`, got, want)
	}

	if got, want := creditCards[1].IsDefault, true; got != want {
		t.Errorf(`creditCards[1].IsDefault = %v; want %v`, got, want)
	}
}

func TestBillingHistory_HandleEvent_CreditCardAdded_DoesNotAddCardWithSameIdAgain(t *testing.T) {
	history := NewBillingHistory()
	testCards := []*CreditCard{
		newExampleCreditCard("card-1"),
		newExampleCreditCard("card-1"),
	}

	userId := "a902d519-f577-4543-b853-bee183e5b403"
	organizationId := "a19737e3-d2f4-48f0-9dfc-ee035833c111"
	for _, creditCard := range testCards {
		creditCardAdded := &BillingEvent{
			EventName:        "credit-card-added",
			OrganizationUuid: organizationId,
			Data: &BillingCreditCardAdded{
				UserUuid:   userId,
				CreditCard: creditCard,
			},
		}
		history.HandleEvent(creditCardAdded)
	}

	creditCards := history.CreditCardsFor(organizationId)
	if got, want := len(creditCards), 1; got != want {
		t.Fatalf(`len(creditCards) = %v; want %v`, got, want)
	}

	if got, want := creditCards[0].CardId, "card-1"; got != want {
		t.Errorf(`creditCards[0].CardId = %v; want %v`, got, want)
	}

}

func TestBillingHistory_HandleEvent_ExtraLimitsGranted_sets_number_of_extra_projects_and_users_for_organization(t *testing.T) {
	history := NewBillingHistory()
	organizationId := "3fb449e1-c05c-4df1-baa3-47239b236bad"
	limitsGranted := &BillingEvent{
		EventName:        "extra-limits-granted",
		OrganizationUuid: organizationId,
		Data: &BillingExtraLimitsGranted{
			Projects: 5,
			Users:    1,
		},
	}

	history.HandleEvent(limitsGranted)

	if got, want := history.ExtraProjectsFor(organizationId), 5; got != want {
		t.Errorf(`history.ExtraProjectsFor(organizationId) = %v; want %v`, got, want)
	}

	if got, want := history.ExtraUsersFor(organizationId), 1; got != want {
		t.Errorf(`history.ExtraUsersFor(organizationId) = %v; want %v`, got, want)
	}
}

func TestBillingHistory_ExtrasGranted_returns_a_list_of_extras_that_have_been_granted_to_the_organization(t *testing.T) {
	history := NewBillingHistory()
	organizationId := "3fb449e1-c05c-4df1-baa3-47239b236bad"
	limitsGranted := &BillingEvent{
		EventName:        "extra-limits-granted",
		OrganizationUuid: organizationId,
		Data: &BillingExtraLimitsGranted{
			Projects: 5,
			Users:    1,
		},
	}

	history.HandleEvent(limitsGranted)

	if got, want := len(history.ExtrasGrantedTo(organizationId)), 1; got != want {
		t.Fatalf(`len(history.ExtrasGrantedTo(organizationId)) = %v; want %v`, got, want)
	}

	if got, want := history.ExtrasGrantedTo(organizationId)[0], limitsGranted; !reflect.DeepEqual(got, want) {
		t.Errorf(`history.ExtrasGrantedTo(organizationId)[0] = %v; want %v`, got, want)
	}
}
