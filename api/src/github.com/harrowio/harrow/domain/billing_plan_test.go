package domain

import "testing"

func TestBillingPlan_EnsureDefaultPrice_ensuresThatPlansAreNotFree(t *testing.T) {
	testcases := []struct {
		UsersIncluded int
		DefaultPrice  Money
	}{
		{1, Money{0, USD}},
		{2, Money{2900, USD}},
		{5, Money{5900, USD}},
		{10, Money{12900, USD}},
	}

	for _, testcase := range testcases {
		plan := &BillingPlan{
			UsersIncluded: testcase.UsersIncluded,
		}

		plan.EnsureDefaultPrice()

		if got, want := plan.PricePerMonth, testcase.DefaultPrice; got != want {
			t.Errorf("UsersIncluded=%d plan.PricePerMonth = %q; want %q",
				testcase.UsersIncluded, got, want)
		}
	}

}

func TestBillingPlan_EnsureDefaultPrice_doesNotOverrideAnExistingPrice(t *testing.T) {
	price := Money{3100, USD}
	plan := &BillingPlan{
		UsersIncluded: 5,
		PricePerMonth: price,
	}

	plan.EnsureDefaultPrice()

	if got, want := plan.PricePerMonth, price; got != want {
		t.Errorf("plan.PricePerMonth = %q; want %q", got, want)
	}
}
