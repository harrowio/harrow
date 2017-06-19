package stores_test

import (
	"testing"

	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/stores"
	"github.com/harrowio/harrow/test_helpers"
)

func Test_ProjectMembershipStore_Create_Succeeds(t *testing.T) {
	tx := test_helpers.GetDbTx(t)
	defer tx.Rollback()
	world := test_helpers.MustNewWorld(tx, t)
	store := stores.NewDbProjectMembershipStore(tx)
	membership := &domain.ProjectMembership{
		ProjectUuid:    world.Project("public").Uuid,
		UserUuid:       world.User("non-member").Uuid,
		MembershipType: domain.MembershipTypeMember,
	}

	uuid, err := store.Create(membership)

	if err != nil {
		t.Fatal(err)
	}
	if uuid == "" {
		t.Fatal("Empty uuid returned")
	}
}

func Test_ProjectMembershipStore_FindByUuid_Succeeds(t *testing.T) {
	tx := test_helpers.GetDbTx(t)
	defer tx.Rollback()
	world := test_helpers.MustNewWorld(tx, t)
	store := stores.NewDbProjectMembershipStore(tx)

	uuid, err := store.Create(&domain.ProjectMembership{
		ProjectUuid:    world.Project("public").Uuid,
		UserUuid:       world.User("non-member").Uuid,
		MembershipType: domain.MembershipTypeMember,
	})
	if err != nil {
		t.Fatal(err)
	}

	membership, err := store.FindByUuid(uuid)
	if err != nil {
		t.Fatal(err)
	}

	if membership.Uuid != uuid {
		t.Fatalf("Expected uuid %s, got %s\n", uuid, membership.Uuid)
	}
}

func Test_ProjectMembershipStore_ArchivedAtIsSetWhenProjectGetsArchived(t *testing.T) {
	tx := test_helpers.GetDbTx(t)
	defer tx.Rollback()
	world := test_helpers.MustNewWorld(tx, t)
	projectStore := stores.NewDbProjectStore(tx)
	membershipStore := stores.NewDbProjectMembershipStore(tx)

	membership := world.ProjectMembership("member")
	if err := projectStore.ArchiveByUuid(membership.ProjectUuid); err != nil {
		t.Fatal(err)
	}

	membership, err := membershipStore.FindByUuid(membership.Uuid)
	if _, ok := err.(*domain.NotFoundError); !ok {
		t.Fatalf("Expected *domain.NotFoundError, got %#v\n", err)
	}
}
