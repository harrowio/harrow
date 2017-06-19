package braintree

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	braintreeAPI "github.com/lionelbarrow/braintree-go"
)

type MockBraintreeAPI struct {
	Plans               []*braintreeAPI.Plan
	CallsToFindAllPlans int
	Error               error
}

func (self *MockBraintreeAPI) FindAllPlans() ([]*braintreeAPI.Plan, error) {
	self.CallsToFindAllPlans++
	return self.Plans, self.Error
}

func NewBraintreeMockWithOnePlan() *MockBraintreeAPI {
	return &MockBraintreeAPI{
		Plans: []*braintreeAPI.Plan{
			{Id: "test-plan"},
		},
	}
}

func TestProxy_FindAllPlans_fillsCacheIfEmpty(t *testing.T) {
	apiClient := NewBraintreeMockWithOnePlan()
	cache := NewMemoryCache()
	proxy := NewProxy(cache, apiClient)

	cache.Clear()
	plans, err := proxy.FindAllPlans()
	if err != nil {
		t.Fatal(err)
	}

	entries, err := cache.Entries()
	if err != nil {
		t.Fatal(err)
	}
	if got, want := plans, entries; !reflect.DeepEqual(got, want) {
		t.Errorf("plans = %#v; want %#v", got, want)
	}
}

func TestProxy_FindAllPlans_returnsCachedEntriesIfAvailable(t *testing.T) {
	apiClient := NewBraintreeMockWithOnePlan()
	cache := NewMemoryCache()
	proxy := NewProxy(cache, apiClient)
	cache.Set(apiClient.Plans)

	_, err := proxy.FindAllPlans()
	if err != nil {
		t.Fatal(err)
	}

	if got, want := apiClient.CallsToFindAllPlans, 0; got != want {
		t.Errorf("apiClient.CallsToFindAllPlans = %d; want %d", got, want)
	}

}

func TestProxy_ClearCache_removesAllEntriesFromTheCache(t *testing.T) {
	apiClient := NewBraintreeMockWithOnePlan()
	cache := NewMemoryCache()
	proxy := NewProxy(cache, apiClient)

	proxy.ClearCache()

	plans, err := proxy.FindAllPlans()
	if err != nil {
		t.Fatal(err)
	}

	entries, err := cache.Entries()
	if err != nil {
		t.Fatal(err)
	}

	if got, want := plans, entries; !reflect.DeepEqual(got, want) {
		t.Errorf("plans = %#v; want %#v", got, want)
	}
}

func TestProxy_ServeHTTP_DELETE_cache_clearsTheCache(t *testing.T) {
	apiClient := NewBraintreeMockWithOnePlan()
	cache := NewMemoryCache()
	cache.Set(apiClient.Plans)
	proxy := NewProxy(cache, apiClient)
	res := httptest.NewRecorder()

	req, err := http.NewRequest("DELETE", "http://example.com/cache", nil)
	if err != nil {
		t.Fatal(err)
	}

	proxy.ServeHTTP(res, req)

	_, err = cache.Entries()
	if got, want := err, ErrCacheMiss; got != want {
		t.Errorf("err = %s; want %s", got, want)
	}
}

func TestProxy_ServeHTTP_GET_returnsAllPlansAsJSON(t *testing.T) {
	apiClient := NewBraintreeMockWithOnePlan()
	cache := NewMemoryCache()
	cache.Set(apiClient.Plans)
	proxy := NewProxy(cache, apiClient)
	res := httptest.NewRecorder()

	req, err := http.NewRequest("GET", "http://example.com/", nil)
	if err != nil {
		t.Fatal(err)
	}

	proxy.ServeHTTP(res, req)

	result := []*braintreeAPI.Plan{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}

	if got, want := len(result), len(apiClient.Plans); got != want {
		t.Errorf("len(result) = %d; want %d", got, want)
	}

	if got, want := result[0].Id, apiClient.Plans[0].Id; got != want {
		t.Errorf("result[0].Id = %q; want %q", got, want)
	}
}
