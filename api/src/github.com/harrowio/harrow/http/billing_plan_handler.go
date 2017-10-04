package http

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gorilla/mux"
	braintree "github.com/lionelbarrow/braintree-go"

	"github.com/harrowio/harrow/activities"
	"github.com/harrowio/harrow/config"
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/stores"
)

var (
	generateBraintreeToken = braintreeClientToken
	findBraintreeCustomer  = braintreeFindCustomer
	// cancelBraintreeSubscription = braintreeCancelSubscription
	// createBraintreeSubscription = braintreeCreateSubscription
	createBraintreeCustomer = braintreeCreateCustomer
	braintreeConfig         *config.BraintreeConfig
)

const (
	ErrorMessageFromBraintreeAboutCanceledSubscription = "Subscription has already been canceled."
)

func init() {
	braintreeConfig = config.GetConfig().Braintree()
}

func braintreeFindCustomer(id string) (*braintree.Customer, error) {
	client := braintreeConfig.NewClient()
	return client.Customer().Find(id)
}

func braintreeCreateCreditCard(user *domain.User, nonce string) (*braintree.CreditCard, error) {
	client := braintreeConfig.NewClient()

	t := &[]bool{true}[0]

	creditCard := &braintree.CreditCard{
		CustomerId:         user.Uuid,
		PaymentMethodNonce: nonce,
		Options: &braintree.CreditCardOptions{
			VerifyCard:  t,
			MakeDefault: true,
		},
	}

	return client.CreditCard().Create(creditCard)
}

func braintreeCreateCustomer(user *domain.User) (*braintree.Customer, error) {
	client := braintreeConfig.NewClient()

	firstName, lastName := "", ""
	nameFields := strings.Split(user.Name, " ")
	if len(nameFields) > 1 {
		firstName = nameFields[0]
		lastName = strings.Join(nameFields[1:], " ")
	}
	customer := &braintree.Customer{
		Id:        user.Uuid,
		FirstName: firstName,
		LastName:  lastName,
	}

	return client.Customer().Create(customer)
}

func braintreeClientToken() (string, error) {
	return braintreeConfig.NewClient().ClientToken().Generate()
}

type BraintreePurchaseRequest struct {
	PlanUuid         string `json:"planUuid"`
	OrganizationUuid string `json:"organizationUuid"`
}

type ClientTokenResponse struct {
	ClientToken string `json:"clientToken"`
	Error       string `json:"error,omitempty"`
}

type billingPlanHandler struct{}

func MountBillingPlanHandler(r *mux.Router, ctxt ServerContext) {

	h := &billingPlanHandler{}

	// Collection
	root := r.PathPrefix("/billing-plans").Subrouter()
	root.PathPrefix("/{uuid}").Methods("GET").Handler(
		HandlerFunc(ctxt, h.Show),
	).Name(
		"billing-plan-show",
	)
	root.PathPrefix("/braintree/webhooks").Methods("POST").Handler(
		HandlerFunc(ctxt, h.BraintreeHandleNotification),
	).Name(
		"billing-plan-braintree-handle-notification",
	)
	root.PathPrefix("/braintree/client-token").Methods("POST").Handler(
		HandlerFunc(ctxt, h.NewBraintreeClientToken),
	).Name(
		"billing-plan-braintree-client-token",
	)
	root.PathPrefix("/braintree/credit-cards").Methods("POST").Handler(
		HandlerFunc(ctxt, h.BraintreeCreditCards),
	).Name(
		"billing-plan-braintree-add-credit-card",
	)
	root.PathPrefix("/braintree/purchase").Methods("POST").Handler(
		HandlerFunc(ctxt, h.BraintreePurchase),
	).Name(
		"billing-plan-braintree-purchase",
	)

	root.Methods("GET").Handler(HandlerFunc(ctxt, h.List)).
		Name("billing-plan-list")

}

func (self *billingPlanHandler) List(ctxt RequestContext) error {
	billingPlanStore := stores.NewDbBillingPlanStore(ctxt.Tx())
	plans, err := billingPlanStore.FindAll()
	if err != nil {
		return err
	}

	interfacePlans := []interface{}{}
	for _, plan := range plans {
		interfacePlans = append(interfacePlans, plan)
	}

	writeCollectionPageAsJson(ctxt, &CollectionPage{
		Total:      len(interfacePlans),
		Count:      len(interfacePlans),
		Collection: interfacePlans,
	})

	return nil
}

func (self *billingPlanHandler) NewBraintreeClientToken(ctxt RequestContext) error {

	token, _ := generateBraintreeToken()

	json.NewEncoder(ctxt.W()).Encode(&ClientTokenResponse{
		ClientToken: token,
	})

	return nil
}

func (self *billingPlanHandler) BraintreePurchase(ctxt RequestContext) error {

	billingEvents := stores.NewDbBillingEventStore(ctxt.Tx())

	params := &BraintreePurchaseRequest{}
	if err := json.NewDecoder(ctxt.R().Body).Decode(&params); err != nil {
		return err
	}

	if params.PlanUuid == "" {
		return domain.NewValidationError("planUuid", "required")
	}

	if params.OrganizationUuid == "" {
		return domain.NewValidationError("organizationUuid", "required")
	}

	plans := stores.NewDbBillingPlanStore(ctxt.Tx())
	plan, err := plans.FindByUuid(params.PlanUuid)
	if err != nil {
		return err
	}

	organizations := stores.NewDbOrganizationStore(ctxt.Tx())
	organization, err := organizations.FindByUuid(params.OrganizationUuid)
	if err != nil {
		switch err.(type) {
		case *domain.NotFoundError:
			return domain.NewValidationError("organizationUuid", "not_found")
		default:
			return err
		}
	}

	history, err := stores.NewDbBillingHistoryStore(ctxt.Tx(), ctxt.KeyValueStore()).Load()
	if err != nil {
		ctxt.Log().Warn().Msgf("failed to load billing history: %s\n", err)
		return err
	}

	existingSubscription := history.SubscriptionFor(organization.Uuid)
	ctxt.Log().Info().Msgf("existingSubscription(%q):\n%s\n",
		organization.Uuid,
		(func() []byte { data, _ := json.Marshal(existingSubscription); return data })(),
	)

	customer, err := findBraintreeCustomer(ctxt.User().Uuid)
	if err != nil {
		customer, err = createBraintreeCustomer(ctxt.User())
		if err != nil {
			return err
		}
	}

	creditCard := customer.DefaultCreditCard()
	if creditCard == nil {
		return domain.NewValidationError("credit-card", "no_default_set")
	}

	event := &domain.BillingPlanSelected{}
	event.FillFromPlan(plan)
	event.UserUuid = ctxt.User().Uuid

	if plan.Uuid == domain.FreePlanUuid {
		event.SubscriptionId = fmt.Sprintf("%s-free-%s",
			organization.Uuid,
			time.Now().Format(time.RFC3339),
		)
	}

	if _, err := billingEvents.Create(organization.NewBillingEvent(event)); err != nil {
		return err
	}

	ctxt.EnqueueActivity(activities.BillingPlanPurchasedWithBraintree(plan), nil)

	return nil
}

func BraintreeNotificationToBillingEventData(notification *braintree.WebhookNotification) (*domain.BillingPlanSubscriptionChanged, error) {
	if notification.Subject.Subscription == nil {
		return nil, nil
	}

	return &domain.BillingPlanSubscriptionChanged{
		PlanId:         notification.Subject.Subscription.PlanId,
		SubscriptionId: notification.Subject.Subscription.Id,
		Status:         strings.ToLower(string(notification.Subject.Subscription.Status)),
	}, nil
}

func (self *billingPlanHandler) BraintreeHandleNotification(ctxt RequestContext) error {

	signature := ctxt.R().FormValue("bt_signature")
	payload := ctxt.R().FormValue("bt_payload")
	webhook, err := braintreeConfig.NewClient().WebhookNotification().Parse(
		signature,
		payload,
	)
	if err != nil {
		return err
	}

	eventData, err := BraintreeNotificationToBillingEventData(webhook)
	if err != nil {
		return err
	}

	if eventData == nil {
		return nil
	}

	history, err := stores.NewDbBillingHistoryStore(ctxt.Tx(), ctxt.KeyValueStore()).Load()
	if err != nil {
		ctxt.Log().Warn().Msgf("failed to load billing history: %s\n", err)
		return err
	}

	subscription := history.Subscription(eventData.SubscriptionId)
	if subscription != nil {
		eventData.UserUuid = subscription.UserUuid
		event := domain.NewBillingEvent(subscription.OrganizationUuid, eventData)
		if _, err := stores.NewDbBillingEventStore(ctxt.Tx()).Create(event); err != nil {
			ctxt.Log().Warn().Msgf("failed to created billing event: %s\n%#v\n", err, event)
			return err
		}
	} else {
		ctxt.Log().Warn().Msgf("subscription %q not found\n", eventData.SubscriptionId)
	}

	return nil
}

func (self *billingPlanHandler) Show(ctxt RequestContext) error {

	planUuid := ctxt.PathParameter("uuid")
	billingPlanStore := stores.NewDbBillingPlanStore(ctxt.Tx())
	plan, err := billingPlanStore.FindByUuid(planUuid)
	if err != nil {
		return err
	}

	writeAsJson(ctxt, plan)
	return nil

}

type AddCreditCardRequest struct {
	Nonce            string `json:"nonce"`
	OrganizationUuid string `json:"organizationUuid"`
}

func (self *billingPlanHandler) BraintreeCreditCards(ctxt RequestContext) error {

	if ctxt.User() == nil {
		return ErrLoginRequired
	}

	params := &AddCreditCardRequest{}
	if err := json.NewDecoder(ctxt.R().Body).Decode(&params); err != nil {
		return err
	}

	if params.Nonce == "" {
		return domain.NewValidationError("nonce", "required")
	}

	if params.OrganizationUuid == "" {
		return domain.NewValidationError("organizationUuid", "required")
	}

	_, err := findBraintreeCustomer(ctxt.User().Uuid)
	if err != nil {
		_, err = createBraintreeCustomer(ctxt.User())
		if err != nil {
			return err
		}
	}

	creditCard, err := braintreeCreateCreditCard(ctxt.User(), params.Nonce)
	if err == nil {
		billingEvents := stores.NewDbBillingEventStore(ctxt.Tx())
		creditCardAdded := &domain.BillingEvent{
			EventName:        "credit-card-added",
			OrganizationUuid: params.OrganizationUuid,
			Data: &domain.BillingCreditCardAdded{
				CreditCard: &domain.CreditCard{
					CardId:         creditCard.UniqueNumberIdentifier,
					IsDefault:      true,
					CardholderName: creditCard.CardholderName,
					SafeCardNumber: creditCard.Last4,
					Token:          creditCard.Token,
				},
				UserUuid:            ctxt.User().Uuid,
				PaymentProviderName: "braintree",
			},
		}

		if _, err := billingEvents.Create(creditCardAdded); err != nil {
			return err
		}

		return nil
	}

	if berr, ok := err.(*braintree.BraintreeError); ok {
		return NewError(StatusUnprocessableEntity, "invalid_credit_card", berr.ErrorMessage)
	} else {
		return err
	}
}
