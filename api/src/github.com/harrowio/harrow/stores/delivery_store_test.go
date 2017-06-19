package stores_test

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/stores"
	"github.com/harrowio/harrow/test_helpers"
	"github.com/harrowio/harrow/uuidhelper"
	"github.com/jmoiron/sqlx"
)

func Test_DbDeliveryStore_Create_generatesUuid(t *testing.T) {
	tx := test_helpers.GetDbTx(t)
	defer tx.Rollback()

	world := test_helpers.MustNewWorld(tx, t)
	delivery := mustNewDelivery(tx, t, world)

	delivery.Uuid = ""

	store := stores.NewDbDeliveryStore(tx)
	uuid, err := store.Create(delivery)
	if err != nil {
		t.Fatal(err)
	}

	found := &domain.Delivery{}

	if err := tx.Get(found, `SELECT * FROM deliveries WHERE uuid = $1`, uuid); err != nil {
		t.Fatal(err)
	}
}

func Test_DbDeliveryStore_Create_generatesDeliveredAt(t *testing.T) {
	tx := test_helpers.GetDbTx(t)
	defer tx.Rollback()

	world := test_helpers.MustNewWorld(tx, t)
	delivery := mustNewDelivery(tx, t, world)

	store := stores.NewDbDeliveryStore(tx)
	uuid, err := store.Create(delivery)
	if err != nil {
		t.Fatal(err)
	}

	found := &domain.Delivery{}

	if err := tx.Get(found, `SELECT * FROM deliveries WHERE uuid = $1`, uuid); err != nil {
		t.Fatal(err)
	}

	if found.DeliveredAt.IsZero() {
		t.Fatal("DeliveredAt is zero")
	}
}

func Test_DbDeliveryStore_Create_insertsDelivery(t *testing.T) {
	tx := test_helpers.GetDbTx(t)
	defer tx.Rollback()

	world := test_helpers.MustNewWorld(tx, t)
	delivery := mustNewDelivery(tx, t, world)

	store := stores.NewDbDeliveryStore(tx)
	uuid, err := store.Create(delivery)
	if err != nil {
		t.Fatal(err)
	}

	found := &domain.Delivery{}

	if err := tx.Get(found, `SELECT * FROM deliveries WHERE uuid = $1`, uuid); err != nil {
		t.Fatal(err)
	}
}

func Test_DbDeliveryStore_FindByUuid_returnsDomainNotFoundError_ifDeliveryDoesNotExist(t *testing.T) {
	tx := test_helpers.GetDbTx(t)
	defer tx.Rollback()

	store := stores.NewDbDeliveryStore(tx)
	doesNotExist := uuidhelper.MustNewV4()
	_, err := store.FindByUuid(doesNotExist)
	if err == nil {
		t.Fatal("Expected an error")
	}

	if derr, ok := err.(*domain.NotFoundError); !ok {
		t.Fatalf("Expected %T, got %T", derr, err)
	}
}

func Test_DbDeliveryStore_FindByUuid_returnsDelivery(t *testing.T) {
	tx := test_helpers.GetDbTx(t)
	defer tx.Rollback()

	world := test_helpers.MustNewWorld(tx, t)
	delivery := mustNewDelivery(tx, t, world)

	store := stores.NewDbDeliveryStore(tx)
	uuid, err := store.Create(delivery)
	if err != nil {
		t.Fatal(err)
	}

	found, err := store.FindByUuid(uuid)
	if err != nil {
		t.Fatal(err)
	}
	if found.Uuid != delivery.Uuid {
		t.Fatalf("Expected to find %q, got %q", delivery.Uuid, found.Uuid)
	}
}

func Test_DbDeliveryStore_FindByWebhookUuid_returnsDeliveries(t *testing.T) {
	tx := test_helpers.GetDbTx(t)
	defer tx.Rollback()

	world := test_helpers.MustNewWorld(tx, t)
	delivery := mustNewDelivery(tx, t, world)

	store := stores.NewDbDeliveryStore(tx)
	_, err := store.Create(delivery)
	if err != nil {
		t.Fatal(err)
	}

	found, err := store.FindByWebhookUuid(delivery.WebhookUuid)
	if err != nil {
		t.Fatal(err)
	}

	for _, foundDelivery := range found {
		if foundDelivery.Uuid == delivery.Uuid {
			// t.Logf("Found %q", delivery.Uuid)
			return
		}
	}

	t.Fatalf("Not found: %q", delivery.Uuid)
}

func mustNewDelivery(tx *sqlx.Tx, t *testing.T, world *test_helpers.World) *domain.Delivery {
	webhook := domain.NewWebhook(world.Project("public").Uuid,
		world.User("default").Uuid,
		world.Job("default").Uuid,
		"github",
	)

	store := stores.NewDbWebhookStore(tx)
	if _, err := store.Create(webhook); err != nil {
		t.Fatal(err)
	}

	httpReq, err := http.NewRequest(
		"POST",
		"http://example.com/webhooks/uuid/slug",
		bytes.NewBufferString(`BODY`),
	)
	if err != nil {
		t.Fatal(err)
	}

	delivery := webhook.NewDelivery(httpReq)

	return delivery
}
