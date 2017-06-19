package http

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/gorilla/mux"
	"github.com/harrowio/harrow/activities"
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/stores"
)

func Test_ProjectMemberHandler_Routing(t *testing.T) {
	r := mux.NewRouter()

	MountProjectMemberHandler(r, nil)

	spec := &routingSpec{
		{"DELETE", "/project-members/:uuid", "project-member-remove"},
		{"PUT", "/project-members", "project-member-update"},
	}

	spec.run(r, t)
}

func Test_ProjectMemberHandler_Remove_removesProjectMembership(t *testing.T) {
	h := NewHandlerTest(MountProjectMemberHandler, t)
	defer h.Cleanup()

	membership := h.World().ProjectMembership("project-member-private")
	member := &domain.ProjectMember{
		User:           &domain.User{Uuid: membership.UserUuid},
		MembershipUuid: &membership.Uuid,
	}
	h.LoginAs("project-owner")
	h.Subject(member)
	h.Do("DELETE", h.UrlFor("self"), url.Values{
		"projectUuid": []string{membership.ProjectUuid},
	})

	if got, want := h.Response().StatusCode, http.StatusOK; got != want {
		t.Errorf("h.Response().StatusCode = %d; want %d", got, want)
	}

	memberships := stores.NewDbProjectMembershipStore(h.Tx())
	_, err := memberships.FindByUuid(membership.Uuid)
	if verr, ok := err.(*domain.NotFoundError); !ok {
		t.Fatalf("memberships.FindByUuid: err.(type) = %T; want %T", err, verr)
	}

}

func Test_ProjectMemberHandler_Remove_emitsProjectLeftActivity_ifUserRemovesHerself(t *testing.T) {
	h := NewHandlerTest(MountProjectMemberHandler, t)
	defer h.Cleanup()

	membership := h.World().ProjectMembership("project-member-private")
	member := &domain.ProjectMember{
		User:           &domain.User{Uuid: membership.UserUuid},
		MembershipUuid: &membership.Uuid,
	}
	h.LoginAs("project-member")
	h.Subject(member)
	h.Do("DELETE", h.UrlFor("self"), url.Values{
		"projectUuid": []string{membership.ProjectUuid},
	})

	check := func(activity *domain.Activity) {
		payload, ok := activity.Payload.(*activities.UserProjectPayload)
		if !ok {
			t.Errorf("payload.(type) = %T; want %T", activity.Payload, payload)
		}

		if got, want := payload.User.Name, h.World().User("project-member").Name; got != want {
			t.Errorf(`payload.User.Name = %v; want %v`, got, want)
		}

		if got, want := payload.Project.Uuid, h.World().Project("private").Uuid; got != want {
			t.Errorf(`payload.Project.Uuid = %v; want %v`, got, want)
		}
	}

	for _, activity := range h.Activities() {
		t.Logf("Activity: %s\n", activity.Name)
		if activity.Name == "user.left-project" {
			check(activity)
			return
		}
	}

	t.Fatalf("Activity %q not found", "user.left-project")

}

func Test_ProjectMemberHandler_Remove_emitsUserRemovedFromProjectActivity_ifUserRemovedBySomeoneElse(t *testing.T) {
	h := NewHandlerTest(MountProjectMemberHandler, t)
	defer h.Cleanup()

	membership := h.World().ProjectMembership("project-member-private")
	member := &domain.ProjectMember{
		User:           &domain.User{Uuid: membership.UserUuid},
		MembershipUuid: &membership.Uuid,
	}
	h.LoginAs("project-owner")
	h.Subject(member)
	h.Do("DELETE", h.UrlFor("self"), url.Values{
		"projectUuid": []string{membership.ProjectUuid},
	})

	check := func(activity *domain.Activity) {
		payload, ok := activity.Payload.(*activities.UserProjectPayload)
		if !ok {
			t.Errorf("payload.(type) = %T; want %T", activity.Payload, payload)
		}

		if got, want := payload.User.Name, h.World().User("project-member").Name; got != want {
			t.Errorf(`payload.User.Name = %v; want %v`, got, want)
		}

		if got, want := payload.Project.Uuid, h.World().Project("private").Uuid; got != want {
			t.Errorf(`payload.Project.Uuid = %v; want %v`, got, want)
		}
	}

	for _, activity := range h.Activities() {
		t.Logf("Activity: %s\n", activity.Name)
		if activity.Name == "user.removed-from-project" {
			check(activity)
			return
		}
	}

	t.Fatalf("Activity %q not found", "user.removed-from-project")

}
