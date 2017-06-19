package stores_test

import (
	"testing"

	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/stores"
	"github.com/harrowio/harrow/test_helpers"
)

type mockInvitationStore struct {
	result *domain.Invitation
}

func newMockInvitationStore(result *domain.Invitation) *mockInvitationStore {
	return &mockInvitationStore{
		result: result,
	}
}

func (self *mockInvitationStore) FindByUserAndProjectUuid(userId, projectId string) (*domain.Invitation, error) {
	return self.result, nil
}

func Test_CachedInvitationStore_searchesFallbackOnCacheMiss(t *testing.T) {
	invitation := &domain.Invitation{}
	mock := newMockInvitationStore(invitation)
	store := stores.NewCachedInvitationStore(mock)

	found, err := store.FindByUserAndProjectUuid("ignored", "ignored")
	if err != nil {
		t.Fatal(err)
	}

	if found != mock.result {
		t.Fatalf("Expected %#v, got %#v", mock.result, found)
	}
}

func Test_InvitationStore_Update_returnsDomainNotFoundError_whenUpdatingNonExistingInvitation(t *testing.T) {
	tx := test_helpers.GetDbTx(t)
	defer tx.Rollback()

	store := stores.NewDbInvitationStore(tx)
	got := store.Update(&domain.Invitation{})
	want, ok := got.(*domain.NotFoundError)
	if !ok {
		t.Errorf("err.(type) = %T; want %T", got, want)
	}
}
