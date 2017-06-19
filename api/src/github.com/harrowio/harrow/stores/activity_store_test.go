package stores_test

import (
	"encoding/json"
	"testing"

	"bytes"

	"github.com/harrowio/harrow/activities"
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/stores"
	"github.com/harrowio/harrow/test_helpers"
	"github.com/jmoiron/sqlx/types"
)

func TestActivityStore_Store_StoresActivities(t *testing.T) {
	tx := test_helpers.GetDbTx(t)
	defer tx.Rollback()

	activity := domain.NewActivity(1, "test.activity")
	activity.Payload = map[string]interface{}{
		"payload": "data",
	}
	activity.Extra = map[string]interface{}{
		"extra": "data",
	}

	store := stores.NewDbActivityStore(tx)
	if err := store.Store(activity); err != nil {
		t.Fatal(err)
	}

	result := struct {
		Payload types.JSONText `db:"payload"`
		Extra   types.JSONText `db:"extra"`
	}{}
	if err := tx.Get(&result, `SELECT payload,extra FROM activities WHERE id = $1`, activity.Id); err != nil {
		t.Fatal(err)
	}

	payload := map[string]interface{}{}
	if err := result.Payload.Unmarshal(&payload); err != nil {
		t.Fatal(err)
	}
	if got, want := payload["payload"].(string), "data"; got != want {
		t.Fatalf(`payload["payload"].(string) = %q; want %q`, got, want)
	}

	extra := map[string]interface{}{}
	if err := result.Extra.Unmarshal(&extra); err != nil {
		t.Fatal(err)
	}
	if got, want := extra["extra"].(string), "data"; got != want {
		t.Fatalf(`extra["extra"].(string) = %q; want %q`, got, want)
	}

}

func TestActivityStore_FindActivityById_returnsActivity(t *testing.T) {
	tx := test_helpers.GetDbTx(t)
	defer tx.Rollback()

	store := stores.NewDbActivityStore(tx)
	activity := domain.NewActivity(1, "test.activity")
	if err := store.Store(activity); err != nil {
		t.Fatal(err)
	}

	found, err := store.FindActivityById(activity.Id)
	if err != nil {
		t.Fatal(err)
	}

	if got, want := found.Id, activity.Id; got != want {
		t.Errorf(`found.Id = %v; want %v`, got, want)
	}

	if got, want := found.Name, activity.Name; got != want {
		t.Errorf(`found.Name = %v; want %v`, got, want)
	}
}

func TestActivityStore_FindActivityById_returnsNotFoundError_ifActivityDoesNotExist(t *testing.T) {
	tx := test_helpers.GetDbTx(t)
	defer tx.Rollback()

	store := stores.NewDbActivityStore(tx)
	_, err := store.FindActivityById(666)
	if err == nil {
		t.Fatal("Expected an error")
	}

	want, ok := err.(*domain.NotFoundError)
	if !ok {
		t.Errorf("err.(type) = %T; want %T", err, want)
	}
}

func TestActivityStore_FindActivityByNameAndPayloadUuid_searches_by_organization_uuid_for_organization_created_activities(t *testing.T) {
	organizations := []*domain.Organization{
		{
			Uuid: "5d4ec70b-74a2-42c1-8723-84ba4f389cde",
		},
		{
			Uuid: "f448214f-6a6c-4258-9576-2a7745671c90",
		},
	}

	tx := test_helpers.GetDbTx(t)
	defer tx.Rollback()

	store := stores.NewDbActivityStore(tx)
	for _, organization := range organizations {
		activity := activities.OrganizationCreated(organization, domain.FreePlan)
		if err := store.Store(activity); err != nil {
			t.Fatal(err)
		}
	}

	activity, err := store.FindActivityByNameAndPayloadUuid("organization.created", organizations[0].Uuid)
	if err != nil {
		t.Fatal(err)
	}

	if got := activity; got == nil {
		t.Fatalf(`activity is nil`)
	}
	{
		got, _ := json.MarshalIndent(activity.Payload, "", "  ")
		want, _ := json.MarshalIndent(activities.OrganizationCreated(organizations[0], domain.FreePlan).Payload, "", "  ")
		if !bytes.Equal(got, want) {
			t.Errorf(`activity.Payload = %s; want %s`, got, want)
		}
	}
}
