package domain

import (
	"fmt"
	"sort"
	"strings"
	"testing"
	"time"
)

func Test_ProjectMember_Links_ContainsLinkToProfilePicture(t *testing.T) {
	member := &ProjectMember{
		User: &User{Email: "vagrant+test@localhost"},
	}

	links := map[string]map[string]string{}

	member.Links(links, "http", "example.com")

	profilePictureUrl := links["profilePicture"]["href"]
	expectedUrl := "https://secure.gravatar.com/avatar/d8772a23405a3610461a0d60891776aa"

	if profilePictureUrl != expectedUrl {
		t.Fatalf("Expected %q to equal %q", profilePictureUrl, expectedUrl)
	}
}

type requiredCapabilityTest struct {
	role         string
	capabilities []string
	expected     []string
	forbidden    []string
}

func (self *requiredCapabilityTest) has(cap string) bool {
	n := sort.SearchStrings(self.capabilities, cap)
	if n == len(self.capabilities) {
		return false
	}

	return self.capabilities[n] == cap
}

func (self *requiredCapabilityTest) run(t *testing.T) {
	sort.Strings(self.capabilities)
	for _, cap := range self.expected {
		if !self.has(cap) {
			t.Errorf("%s cannot %s", self.role, cap)
		}
	}
	for _, cap := range self.forbidden {
		if self.has(cap) {
			t.Errorf("%s can %s", self.role, cap)
		}
	}
}

func Test_ProjectMember_Capabilities_GuestsHaveOnlyReadCapabilities(t *testing.T) {
	member := &ProjectMember{
		MembershipType: MembershipTypeGuest,
	}

	for _, capability := range member.Capabilities() {
		if !strings.HasPrefix(capability, "read") {
			t.Errorf("guest can %s\n", capability)
		}
	}
}

func Test_ProjectMember_Capabilities_ExpectedForMember(t *testing.T) {
	member := &ProjectMember{
		MembershipType: MembershipTypeMember,
	}

	expected := newCapabilityList().
		reads(
			"project",
			"project-member",
			"delivery",
			"webhook",
			"operation",
			"environment",
			"job",
			"task",
			"schedule",
			"repository",
		).
		writesFor("schedule").
		writesFor("subscription").
		strings()

	forbidden := newCapabilityList().
		writesFor("project").
		writesFor("environment").
		writesFor("job").
		writesFor("webhook").
		writesFor("delivery").
		writesFor("repository").
		writesFor("task").
		writesFor("project-member").
		strings()

	test := requiredCapabilityTest{
		role:         "member",
		capabilities: member.Capabilities(),
		expected:     expected,
		forbidden:    forbidden,
	}

	test.run(t)
}

func Test_ProjectMember_Capabilities_ExpectedForManager(t *testing.T) {
	member := &ProjectMember{
		MembershipType: MembershipTypeManager,
	}

	expected := newCapabilityList().
		reads(
			"project",
			"project-member",
			"delivery",
			"webhook",
			"operation",
			"environment",
			"job",
			"task",
			"schedule",
			"repository",
		).
		writesFor("schedule").
		writesFor("environment").
		writesFor("job").
		writesFor("webhook").
		writesFor("repository").
		writesFor("task").
		writesFor("invitation").
		writesFor("project-member").
		strings()

	forbidden := newCapabilityList().
		archives("project").
		strings()

	test := requiredCapabilityTest{
		role:         MembershipTypeManager,
		capabilities: member.Capabilities(),
		expected:     expected,
		forbidden:    forbidden,
	}

	test.run(t)
}

func Test_ProjectMember_Capabilities_ExpectedForOwner(t *testing.T) {
	owner := &ProjectMember{
		MembershipType: MembershipTypeOwner,
	}

	manager := &ProjectMember{MembershipType: MembershipTypeManager}
	expected := newCapabilityList().
		add(manager.Capabilities()).
		archives("project").
		strings()

	test := requiredCapabilityTest{
		role:         MembershipTypeOwner,
		capabilities: owner.Capabilities(),
		expected:     expected,
	}

	test.run(t)
}

func Test_ProjectMember_Promote_requiresHigherLevelMembership(t *testing.T) {
	testcases := []struct {
		promoterMembershipType  string
		promotedMembershipType  string
		resultingMembershipType string
		err                     error
	}{
		{MembershipTypeOwner, MembershipTypeManager, MembershipTypeOwner, nil},
		{MembershipTypeManager, MembershipTypeMember, MembershipTypeManager, nil},
		{MembershipTypeMember, MembershipTypeGuest, MembershipTypeMember, nil},
		{MembershipTypeOwner, MembershipTypeMember, MembershipTypeManager, nil},

		{MembershipTypeMember, MembershipTypeManager, MembershipTypeManager, NewValidationError("membershipType", "too_low")},
	}

	for _, test := range testcases {
		promoter := &ProjectMember{MembershipType: test.promoterMembershipType}
		promoted := &ProjectMember{MembershipType: test.promotedMembershipType}

		err := promoter.Promote(promoted)

		if got, want := fmt.Sprintf("%s", err), fmt.Sprintf("%s", test.err); got != want {
			t.Errorf("err = %q; want %q", got, want)
			continue

		}

		if got, want := promoted.MembershipType, test.resultingMembershipType; got != want {
			t.Errorf("promoted.MembershipType = %q; want %q", got, want)
		}
	}
}

func Test_ProjectMember_NewProjectMember_makesOrganizationOwnerProjectOwner(t *testing.T) {
	user := &User{
		Uuid: "ec539978-153c-4911-a3cc-d0fef05f1b59",
		Name: "Test User",
	}

	orgMembership := &OrganizationMembership{
		UserUuid:  user.Uuid,
		Type:      MembershipTypeOwner,
		CreatedAt: time.Now(),
	}

	project := &Project{
		Name: "Test Project",
		Uuid: "1245e0f4-7b83-4bb0-b7f8-8c8ac5d20550",
	}

	member := NewProjectMember(user, project, nil, orgMembership)
	if got, want := member.MembershipType, MembershipTypeOwner; got != want {
		t.Errorf("member.MembershipType = %q; want %q", got, want)
	}
}

func Test_ProjectMember_ToMembership_setsCreatedAt_fromOrganizationMembership(t *testing.T) {
	user := &User{
		Uuid: "ec539978-153c-4911-a3cc-d0fef05f1b59",
		Name: "Test User",
	}

	orgMembership := &OrganizationMembership{
		UserUuid:  user.Uuid,
		Type:      MembershipTypeMember,
		CreatedAt: time.Now(),
	}

	project := &Project{
		Name: "Test Project",
		Uuid: "1245e0f4-7b83-4bb0-b7f8-8c8ac5d20550",
	}

	member := NewProjectMember(user, project, nil, orgMembership)
	if got, want := member.CreatedAt, orgMembership.CreatedAt; !got.Equal(want) {
		t.Errorf("member.CreatedAt = %q; want %q", got, want)
	}
}

func Test_ProjectMember_ToMembership_setsCreatedAt_fromProjectMembership_ifAvailable(t *testing.T) {
	user := &User{
		Uuid: "ec539978-153c-4911-a3cc-d0fef05f1b59",
		Name: "Test User",
	}

	orgMembership := &OrganizationMembership{
		UserUuid:  user.Uuid,
		Type:      MembershipTypeMember,
		CreatedAt: time.Now().Add(-24 * time.Hour),
	}

	project := &Project{
		Name: "Test Project",
		Uuid: "1245e0f4-7b83-4bb0-b7f8-8c8ac5d20550",
	}

	projectMembership := &ProjectMembership{
		Uuid:           "c21666a3-7102-4fbc-b2cb-5c2186620697",
		UserUuid:       user.Uuid,
		ProjectUuid:    project.Uuid,
		MembershipType: MembershipTypeMember,
		CreatedAt:      time.Now(),
	}

	member := NewProjectMember(user, project, projectMembership, orgMembership)
	if got, want := member.CreatedAt, projectMembership.CreatedAt; !got.Equal(want) {
		t.Errorf("member.CreatedAt = %q; want %q", got, want)
	}
}

func Test_ProjectMember_ToMembership_setsCreatedAt(t *testing.T) {
	membershipUuid := "f7b53677-ac3e-4485-ad7b-4fefd56d1882"
	member := &ProjectMember{
		User:           &User{Uuid: "a1787222-1672-471f-a1c3-404ecbb17b5c"},
		MembershipUuid: &membershipUuid,
		ProjectUuid:    "7c27a371-85a4-44d0-8c67-f32c13cf645f",
		MembershipType: MembershipTypeMember,
		CreatedAt:      time.Now(),
	}
	membership := member.ToMembership()
	if got, want := membership.CreatedAt, member.CreatedAt; !got.Equal(want) {
		t.Errorf("membership.CreatedAt = %q; want = %q", got, want)
	}
}

func Test_ProjectMember_Remove_requiresManagerOrOwner(t *testing.T) {
	testcases := []struct {
		remover string
		removed string
		err     error
	}{
		{MembershipTypeOwner, MembershipTypeMember, nil},
		{MembershipTypeManager, MembershipTypeMember, nil},
		{MembershipTypeManager, MembershipTypeManager, NewValidationError("membershipType", "too_low")},
		{MembershipTypeOwner, MembershipTypeManager, nil},
		{MembershipTypeOwner, MembershipTypeOwner, nil},
		{MembershipTypeMember, MembershipTypeGuest, NewValidationError("membershipType", "too_low")},
		{MembershipTypeGuest, MembershipTypeGuest, NewValidationError("membershipType", "too_low")},
	}

	projectMemberships := NewMockArchiver()
	membershipUuid := "888767af-17ab-4892-859d-64ef70089050"
	projectMemberships.archived[membershipUuid] = 0
	for _, test := range testcases {

		remover := &ProjectMember{MembershipType: test.remover}
		toRemove := &ProjectMember{MembershipType: test.removed, MembershipUuid: &membershipUuid}
		err := remover.Remove(toRemove, projectMemberships)
		if got, want := fmt.Sprintf("%s", err), fmt.Sprintf("%s", test.err); got != want {
			t.Errorf("err = %q; want %q", got, want)
		}
	}
}

func Test_ProjectMember_Remove_archivesMembership(t *testing.T) {
	remover := &ProjectMember{
		MembershipType: MembershipTypeOwner,
	}
	toRemoveUuid := "69140030-3f05-40e8-bf0e-9ca07e4a4853"
	toRemove := &ProjectMember{
		MembershipUuid: &toRemoveUuid,
		MembershipType: MembershipTypeMember,
	}
	projectMemberships := NewMockArchiver()
	projectMemberships.archived[toRemoveUuid] = 0

	if err := remover.Remove(toRemove, projectMemberships); err != nil {
		t.Fatal(err)
	}

	if got, want := projectMemberships.archived[toRemoveUuid], 1; got != want {
		t.Errorf("projectMemberships.archived[toRemoveUuid] = %d; want %d", got, want)
	}
}

func TestProjectMember_Remove_allowsUserToRemoveHerselfRegardlessOfMembershipType(t *testing.T) {
	toRemoveUuid := "69140030-3f05-40e8-bf0e-9ca07e4a4853"
	membershipUuid := "9da258a5-ad2f-41e7-8ba9-c2f7b251c03b"
	remover := &ProjectMember{
		User: &User{
			Uuid: toRemoveUuid,
		},
		MembershipUuid: &membershipUuid,
		MembershipType: MembershipTypeMember,
	}
	toRemove := &ProjectMember{
		User: &User{
			Uuid: toRemoveUuid,
		},
		MembershipUuid: &membershipUuid,
		MembershipType: MembershipTypeMember,
	}
	projectMemberships := NewMockArchiver()
	projectMemberships.archived[membershipUuid] = 0

	if err := remover.Remove(toRemove, projectMemberships); err != nil {
		t.Fatal(err)
	}

	if got, want := projectMemberships.archived[membershipUuid], 1; got != want {
		t.Errorf("projectMemberships.archived[membershipUuid] = %d; want %d", got, want)
	}
}
