package domain

import "testing"

func Test_ProjectMember_and_ProjectMembership_haveSameAuthorizationName(t *testing.T) {
	member := &ProjectMember{}
	membership := &ProjectMembership{}

	if m, ms := member.AuthorizationName(), membership.AuthorizationName(); m != ms {
		t.Fatalf("Expected member(%q) to be membership(%q)", m, ms)
	}
}

func Test_OrganizationMember_and_OrganizationMembership_haveSameAuthorizationName(t *testing.T) {
	member := &OrganizationMember{}
	membership := &OrganizationMembership{}

	if m, ms := member.AuthorizationName(), membership.AuthorizationName(); m != ms {
		t.Fatalf("Expected member(%q) to be membership(%q)", m, ms)
	}
}

func Test_Schedule_and_ScheduledExecution_haveSameAuthorizationName(t *testing.T) {
	schedule := &Schedule{}
	scheduledExecution := &ScheduledExecution{}

	if s, se := schedule.AuthorizationName(), scheduledExecution.AuthorizationName(); s != se {
		t.Fatalf("Expected schedule(%q) to be scheduledExecution(%q)", s, se)
	}
}

func Test_Subscription_and_Subscriptions_haveSameAuthorizationName(t *testing.T) {
	subscription := &Subscription{}
	subscriptions := &Subscriptions{}

	if s, ss := subscription.AuthorizationName(), subscriptions.AuthorizationName(); s != ss {
		t.Fatalf("Expected subscription(%q) to be subscriptions(%q)", s, ss)
	}
}
