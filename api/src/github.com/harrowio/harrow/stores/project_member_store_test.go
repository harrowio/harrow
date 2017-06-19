package stores_test

import (
	"encoding/json"
	"testing"

	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/stores"
	"github.com/harrowio/harrow/test_helpers"
)

func Test_ProjectMemberStore_FindAllByProjectUuid_returnsOrgMembersAsWell(t *testing.T) {

	t.Skip("we changed behaviour here, is the test wrong?")

	tx := test_helpers.GetDbTx(t)
	defer tx.Rollback()
	world := test_helpers.MustNewWorld(tx, t)
	project := world.Project("public")

	members := []*domain.ProjectMember{
		{User: &domain.User{Uuid: world.OrganizationMembership("member").UserUuid}},
		{User: &domain.User{Uuid: world.OrganizationMembership("owner").UserUuid}},
	}

	store := stores.NewDbProjectMemberStore(tx)
	projectMembers, err := store.FindAllByProjectUuid(project.Uuid)
	if err != nil {
		t.Fatal(err)
	}

	memberCount := len(projectMembers)
	if memberCount != len(members) {
		t.Logf("Expected %d results, got %d", len(members), memberCount)
		t.Logf("Results:\n")
		for i, m := range projectMembers {
			mj, _ := json.MarshalIndent(m, "", " ")
			t.Logf("[%d] %s\n", i, mj)
		}
		t.FailNow()
	}

	for i, m := range projectMembers {
		if members[i].Uuid != m.Uuid {
			t.Errorf("[%d/%d] Expected %q, got %q", i+1, memberCount, members[i].Uuid, m.Uuid)
		}
	}
}
