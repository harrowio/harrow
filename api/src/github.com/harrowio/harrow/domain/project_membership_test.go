package domain

import (
	"reflect"
	"testing"
	"time"
)

func Test_ProjectMembership_OwnUrl(t *testing.T) {
	now := time.Now()
	pm := &ProjectMembership{
		Uuid:           "c320e6ac-4f63-428b-bef0-538c6afd3bb3",
		ProjectUuid:    "09b4920c-d942-4735-a4c8-c3c1bf203c7f",
		UserUuid:       "1f658c71-f471-477e-8e15-013a7712a33d",
		MembershipType: MembershipTypeMember,
		CreatedAt:      now,
		ArchivedAt:     nil,
	}

	expected := "http://example.com/project-memberships/" + pm.Uuid
	if url := pm.OwnUrl("http", "example.com"); url != expected {
		t.Fatalf("Expected %#v to equal %#v\n", url, expected)
	}
}

func Test_ProjectMembership_Links(t *testing.T) {
	links := map[string]map[string]string{}
	now := time.Now()
	pm := &ProjectMembership{
		Uuid:           "c320e6ac-4f63-428b-bef0-538c6afd3bb3",
		ProjectUuid:    "09b4920c-d942-4735-a4c8-c3c1bf203c7f",
		UserUuid:       "1f658c71-f471-477e-8e15-013a7712a33d",
		MembershipType: MembershipTypeMember,
		CreatedAt:      now,
		ArchivedAt:     nil,
	}

	pm.Links(links, "http", "example.com")

	expected := map[string]map[string]string{
		"project": {
			"href": "http://example.com/projects/" + pm.ProjectUuid,
		},
		"user": {
			"href": "http://example.com/users/" + pm.UserUuid,
		},
	}

	if !reflect.DeepEqual(links, expected) {
		t.Fatalf("Expected %#v to equal %#v\n", links, expected)
	}
}
