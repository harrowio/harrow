package domain

import "testing"

func Test_OrganizationMember_Links_containsLinkToProfilePicture(t *testing.T) {
	member := &OrganizationMember{
		User: &User{
			Email: "vagrant+TEST@locahost",
		},
	}
	response := map[string]map[string]string{}
	member.Links(response, "http", "example.com")

	expected := newGravatarUrl(member.User.Email).String()
	actual := response["profilePicture"]["href"]
	if expected != actual {
		t.Fatalf("Expected %q to be %q", actual, expected)
	}
}

func Test_OrganizationMember_Capabilities_ExpectedGuestCapabilities(t *testing.T) {
	member := &OrganizationMember{
		MembershipType: MembershipTypeGuest,
	}

	expected := newCapabilityList().
		reads("organization", "project-member", "organization-member").
		strings()

	forbidden := newCapabilityList().
		writesFor("organization").
		writesFor("project").
		writesFor("project-member").
		strings()

	test := requiredCapabilityTest{
		role:         member.MembershipType,
		capabilities: member.Capabilities(),
		expected:     expected,
		forbidden:    forbidden,
	}

	test.run(t)
}

func Test_OrganizationMember_Capabilities_ExpectedMemberCapabilities(t *testing.T) {
	member := &OrganizationMember{
		MembershipType: MembershipTypeMember,
	}

	expected := newCapabilityList().
		reads("organization", "project-member", "organization-member").
		creates("project").
		strings()

	forbidden := newCapabilityList().
		archives("project").
		updates("project").
		strings()

	test := requiredCapabilityTest{
		role:         member.MembershipType,
		capabilities: member.Capabilities(),
		expected:     expected,
		forbidden:    forbidden,
	}

	test.run(t)
}

func Test_OrganizationMember_Capabilities_ExpectedManagerCapabilities(t *testing.T) {
	member := &OrganizationMember{
		MembershipType: MembershipTypeManager,
	}

	expected := newCapabilityList().
		reads("organization", "project-member", "organization-member").
		writesFor("project").
		writesFor("organization-member").
		strings()

	forbidden := newCapabilityList().
		archives("organization").
		strings()

	test := requiredCapabilityTest{
		role:         member.MembershipType,
		capabilities: member.Capabilities(),
		expected:     expected,
		forbidden:    forbidden,
	}

	test.run(t)

}

func Test_OrganizationMember_Capabilities_ExpectedOwnerCapabilities(t *testing.T) {
	member := &OrganizationMember{
		MembershipType: MembershipTypeOwner,
	}

	expected := newCapabilityList().
		reads("organization", "project-member", "organization-member").
		writesFor("project").
		writesFor("organization-member").
		writesFor("organization").
		strings()

	test := requiredCapabilityTest{
		role:         member.MembershipType,
		capabilities: member.Capabilities(),
		expected:     expected,
	}

	test.run(t)

}
