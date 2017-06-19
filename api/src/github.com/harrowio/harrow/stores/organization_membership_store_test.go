// Package stores provides storage objects, they all have a trivial interface
// for finding by UUID, updateing and deleting
package stores_test

import (
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/stores"

	"testing"

	helpers "github.com/harrowio/harrow/test_helpers"
	"github.com/jmoiron/sqlx"
)

func setupOperationMembershipStoreTest(t *testing.T) *sqlx.Tx {

	tx := helpers.GetDbTx(t)

	helpers.MustCreateUser(t, tx, &domain.User{
		Uuid:     "11111111-1111-4111-a111-111111111111",
		Name:     "Anne Mustermann",
		Email:    "anne@musterma.nn",
		Password: "changeme123",
	})

	helpers.MustCreateOrganization(t, tx, &domain.Organization{
		Uuid: "22222222-2222-4222-a222-222222222222",
		Name: "Musterorg",
	})

	return tx

}

func Test_OrganizationMembershipStore_LookingUpMembershipsByUserUuid(t *testing.T) {

	tx := setupOperationMembershipStoreTest(t)
	defer tx.Rollback()

	orgMembershipStore := stores.NewDbOrganizationMembershipStore(tx)

	orgUuid := "22222222-2222-4222-a222-222222222222"
	userUuid := "11111111-1111-4111-a111-111111111111"

	om := &domain.OrganizationMembership{
		OrganizationUuid: orgUuid,
		UserUuid:         userUuid,
		Type:             domain.MembershipTypeGuest,
	}
	err := orgMembershipStore.Create(om)
	if err != nil {
		t.Fatal(err)
	}

	oms, err := orgMembershipStore.FindAllByUserUuid(userUuid)
	if err != nil {
		t.Fatal(err)
	}

	if len(oms) != 1 {
		t.Fatalf("Expected len(oms) != 1, got %v", len(oms))
	}

	if oms[0].OrganizationUuid != orgUuid {
		t.Fatalf("Expected oms[0].OrganizationUuid != orgUuid, got %v != %v", oms[0].OrganizationUuid, orgUuid)
	}

	if oms[0].UserUuid != userUuid {
		t.Fatalf("Expected oms[0].UserUuid != userUuid, got %v != %v", oms[0].UserUuid, userUuid)
	}

	if oms[0].Type != domain.MembershipTypeGuest {
		t.Fatalf("Expected oms[0].Type != domain.MembershipTypeGuest, got %v != %v", oms[0].Type, domain.MembershipTypeGuest)
	}

}

func Test_OrganizationMembershipStore_SuccessfullyCreatingANewOrganizationMembership(t *testing.T) {

	tx := setupOperationMembershipStoreTest(t)
	defer tx.Rollback()

	orgMembershipStore := stores.NewDbOrganizationMembershipStore(tx)

	orgUuid := "22222222-2222-4222-a222-222222222222"
	userUuid := "11111111-1111-4111-a111-111111111111"

	om := &domain.OrganizationMembership{
		OrganizationUuid: orgUuid,
		UserUuid:         userUuid,
		Type:             domain.MembershipTypeGuest,
	}

	err := orgMembershipStore.Create(om)
	if err != nil {
		t.Fatal(err)
	}

	om, err = orgMembershipStore.FindByOrganizationAndUserUuids(orgUuid, userUuid)
	if err != nil {
		t.Fatal(err)
	}
	if om == nil {
		t.Fatalf("Expected om == nil, got %v == %v", om, nil)
	}

}

func Test_OrganizationMembershipStore_FailingToCreateIfTheConstraintsAreNotMet(t *testing.T) {

	tx := setupOperationMembershipStoreTest(t)
	defer tx.Rollback()

	orgMembershipStore := stores.NewDbOrganizationMembershipStore(tx)

	orgUuid := "33333333-3333-4333-a333-333333333333"
	userUuid := "11111111-1111-4111-a111-111111111111"

	om := &domain.OrganizationMembership{
		OrganizationUuid: orgUuid,
		UserUuid:         userUuid,
		Type:             domain.MembershipTypeGuest,
	}

	err := orgMembershipStore.Create(om)
	if err == nil {
		t.Fatalf("Expected err == nil, got %v == %v", err, nil)
	}
}

func Test_OrganizationMembershipStore_FindThroughProjectMemberships_returnsGuestMembership(t *testing.T) {
	tx := helpers.GetDbTx(t)
	defer tx.Rollback()

	world := helpers.MustNewWorld(tx, t)
	project := world.Project("public")
	user := world.User("other")
	projectMembership := project.NewMembership(user, domain.MembershipTypeMember)
	helpers.MustCreateProjectMembership(t, tx, projectMembership)
	store := stores.NewDbOrganizationMembershipStore(tx)
	membership, err := store.FindThroughProjectMemberships(project.OrganizationUuid, user.Uuid)

	if err != nil {
		t.Fatal(err)
	}

	if got, want := membership.Type, domain.MembershipTypeGuest; got != want {
		t.Errorf("membership.Type = %q; want %q", got, want)
	}
}

func Test_OrganizationMembershipStore_FindThroughProjectMemberships_returnsNothingIfNoProjectMembershipExists(t *testing.T) {
	tx := helpers.GetDbTx(t)
	defer tx.Rollback()

	world := helpers.MustNewWorld(tx, t)
	user := world.User("non-member")
	project := world.Project("private")
	store := stores.NewDbOrganizationMembershipStore(tx)
	membership, err := store.FindThroughProjectMemberships(project.OrganizationUuid, user.Uuid)
	if err == nil {
		t.Fatal("Expected an error")
	}

	want, ok := err.(*domain.NotFoundError)
	if !ok {
		t.Errorf("err = %T, want %T", err, want)
	}

	if membership != nil {
		t.Errorf("membership = %#v; want nil", membership)
	}
}

func Test_OrganizationMembershipStore_FindThroughProjectMemberships_ignoresArchivedProjects(t *testing.T) {
	tx := helpers.GetDbTx(t)
	defer tx.Rollback()

	world := helpers.MustNewWorld(tx, t)
	user := world.User("non-member")
	project := world.Project("private")
	store := stores.NewDbOrganizationMembershipStore(tx)
	projectMembership := project.NewMembership(user, domain.MembershipTypeMember)
	helpers.MustCreateProjectMembership(t, tx, projectMembership)
	projectStore := stores.NewDbProjectStore(tx)
	if err := projectStore.ArchiveByUuid(project.Uuid); err != nil {
		t.Fatal(err)
	}

	membership, err := store.FindThroughProjectMemberships(project.OrganizationUuid, user.Uuid)
	if err == nil {
		t.Fatal("Expected an error")
	}

	want, ok := err.(*domain.NotFoundError)
	if !ok {
		t.Errorf("err = %T, want %T", err, want)
	}

	if membership != nil {
		t.Errorf("membership = %#v; want nil", membership)
	}
}
