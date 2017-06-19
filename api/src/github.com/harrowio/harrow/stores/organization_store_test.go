// Package stores provides storage objects, they all have a trivial interface
// for finding by UUID, updateing and deleting
package stores_test

import (
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/stores"

	"testing"

	helpers "github.com/harrowio/harrow/test_helpers"
)

func Test_OrganizationStore_SuccessfullyCreatingANewOrganization(t *testing.T) {

	tx := helpers.GetDbTx(t)
	defer tx.Rollback()

	orgStore := stores.NewDbOrganizationStore(tx)

	_, err := orgStore.Create(&domain.Organization{Name: "Test Organization"})
	if err != nil {
		t.Fatal(err)
	}

}

func Test_OrganizationStore_SucessfullyDeleteOrganization(t *testing.T) {

	var err error

	tx := helpers.GetDbTx(t)
	defer tx.Rollback()

	o := helpers.MustCreateOrganization(t, tx, &domain.Organization{Name: "Test Organization"})

	orgStore := stores.NewDbOrganizationStore(tx)

	err = orgStore.DeleteByUuid(o.Uuid)
	if err != nil {
		t.Fatal(err)
	}

	_, err = orgStore.FindByUuid(o.Uuid)
	if err == nil {
		t.Fatalf("Expected err == nil, got %v == %v", err, nil)
	}

}

func Test_OrganizationStore_FailToDeleteNonExistentOrganization(t *testing.T) {

	tx := helpers.GetDbTx(t)
	defer tx.Rollback()

	orgStore := stores.NewDbOrganizationStore(tx)

	err := orgStore.DeleteByUuid("11111111-1111-4111-a111-111111111111")
	if _, ok := err.(*domain.NotFoundError); !ok {
		t.Fatalf("Expected to get a domain.NotFoundError, got: %v", err)
	}

}

func Test_OrganizationStore_FailingToCreateUserDueToShortName(t *testing.T) {

	tx := helpers.GetDbTx(t)
	defer tx.Rollback()

	store := stores.NewDbOrganizationStore(tx)

	u := &domain.Organization{
		Name: "",
	}

	_, err := store.Create(u)
	if ve, ok := err.(*domain.ValidationError); ok {
		if ve.Errors["name"][0] != "too_short" {
			t.Fatalf("Expected ve.Errors[\"name\"] != \"too_short\" got ve.Errors[\"name\"] == %v", ve.Errors["name"])
		}
	} else {
		t.Fatalf("Expected ve.(type) == domain.ValidationError, got error %v", err)
	}

}

func Test_OrganizationStore_Update_returnsDomainNotFoundError_whenUpdatingNonExistingOrganization(t *testing.T) {
	tx := helpers.GetDbTx(t)
	defer tx.Rollback()

	store := stores.NewDbOrganizationStore(tx)
	got := store.Update(&domain.Organization{})
	want, ok := got.(*domain.NotFoundError)
	if !ok {
		t.Errorf("err.(type) = %T; want %T", got, want)
	}
}
