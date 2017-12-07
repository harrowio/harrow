package http

import (
	"encoding/json"
	"testing"

	"github.com/gorilla/mux"
	"github.com/harrowio/harrow/test_helpers"
	braintree "github.com/lionelbarrow/braintree-go"
)

func Test_BillingPlanHandler_Routing(t *testing.T) {
	r := mux.NewRouter()
	MountBillingPlanHandler(r, nil)

	spec := routingSpec{
		{"GET", "/billing-plans", "billing-plan-list"},
		{"GET", "/billing-plans/:plan-uuid", "billing-plan-show"},
		{"POST", "/billing-plans/braintree/webhooks", "billing-plan-braintree-handle-notification"},
		{"POST", "/billing-plans/braintree/client-token", "billing-plan-braintree-client-token"},
		{"POST", "/billing-plans/braintree/credit-cards", "billing-plan-braintree-add-credit-card"},
		{"POST", "/billing-plans/braintree/purchase", "billing-plan-braintree-purchase"},
	}

	spec.run(r, t)
}

func Test_BillingPlanHandler_NewBraintreeClientToken_returnsANewBraintreeClientToken(t *testing.T) {
	h := NewHandlerTest(MountBillingPlanHandler, t)
	defer h.Cleanup()

	clientToken := "client-token"
	generateBraintreeToken = func() (string, error) {
		return clientToken, nil
	}

	result := &ClientTokenResponse{}
	h.ResultTo(result)
	h.Do("POST", h.Url("/billing-plans/braintree/client-token"), nil)

	if got, want := result.ClientToken, clientToken; got != want {
		t.Errorf(`resultC.lientToken = %v; want %v`, got, want)
	}
}

func Test_BillingPlanHandler_List_returnsListOfAllPlans(t *testing.T) {
	if !test_helpers.IsBraintreeProxyAvailable() {
		t.Skip("braintree-proxy not running")
		return
	}

	h := NewHandlerTest(MountBillingPlanHandler, t)
	defer h.Cleanup()

	result := struct {
		Collection []struct {
			Subject struct {
				Name string
			}
		}
	}{}

	expectedPlanNames := []string{
		"Free",
		"Silver",
		"Gold",
		"Platinum",
	}

	h.ResultTo(&result)
	h.Do("GET", h.Url("/billing-plans"), nil)

	if got, want := len(result.Collection), len(expectedPlanNames); got != want {
		t.Errorf("len(result.Collection) = %d; want %d", got, want)
	}

	for i, billingPlan := range result.Collection {
		if i > len(expectedPlanNames) {
			break
		}

		if got, want := billingPlan.Subject.Name, expectedPlanNames[i]; got != want {
			t.Errorf("billingPlan.Subject.Name[%d] = %q; want %q", i, got, want)
		}
	}

}

const (
	SubscriptionCancelledWebhookAsJSON = `{"XMLName":{"Space":"","Local":"notification"},"Timestamp":"2015-11-10T12:35:07Z","Kind":"subscription_canceled","Subject":{"XMLName":{"Space":"","Local":"subject"},"APIErrorResponse":null,"Disbursement":null,"Subscription":{"XMLName":"","Id":"79mh9g","Balance":"0.00","BillingDayOfMonth":"10","BillingPeriodEndDate":"","BillingPeriodStartDate":"","CurrentBillingCycle":"0","DaysPastDue":"","DiscountList":[null],"FailureCount":"0","FirstBillingDate":"2016-02-10","MerchantAccountId":"harrowinc","NeverExpires":"true","NextBillAmount":"29.00","NextBillingPeriodAmount":"29.00","NextBillingDate":"2016-02-10","NumberOfBillingCycles":"","PaidThroughDate":"","PaymentMethodToken":"8f67jg","PlanId":"2015-07-21-silver","Price":"29.00","Status":"Canceled","TrialDuration":"3","TrialDurationUnit":"month","TrialPeriod":"true","Transactions":{"Transaction":null},"Options":null},"MerchantAccount":null,"Transaction":null}}`
)

func MustNewWebhookNotification() *braintree.WebhookNotification {
	notification := new(braintree.WebhookNotification)
	if err := json.Unmarshal([]byte(SubscriptionCancelledWebhookAsJSON), notification); err != nil {
		panic(err)
	}

	return notification
}

func TestBillingHandler_BraintreeNotificationToBillingEvent_createsABillingEventForSubscriptions(t *testing.T) {
	t.Skip("not relevant anymore")
	notification := MustNewWebhookNotification()
	subscriptionCanceled, err := BraintreeNotificationToBillingEventData(notification)
	if err != nil {
		t.Fatal(err)
	}

	if got, want := subscriptionCanceled.PlanId, "2015-07-21-silver"; got != want {
		t.Errorf(`subscriptionCanceled.PlanId = %v; want %v`, got, want)
	}

	if got, want := subscriptionCanceled.SubscriptionId, "79mh9g"; got != want {
		t.Errorf(`subscriptionCanceled.SubscriptionId = %v; want %v`, got, want)
	}

	if got, want := subscriptionCanceled.Status, "canceled"; got != want {
		t.Errorf(`subscriptionCanceled.Status = %v; want %v`, got, want)
	}
}
