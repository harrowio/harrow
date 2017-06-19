package stores_test

import (
	"reflect"
	"sort"
	"time"

	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/stores"

	"testing"

	helpers "github.com/harrowio/harrow/test_helpers"
)

func Test_UserStore_CreatingANewUserSuccessfully(t *testing.T) {

	tx := helpers.GetDbTx(t)
	defer tx.Rollback()

	store := stores.NewDbUserStore(tx, helpers.GetConfig(t))

	u := &domain.User{
		Name:     "Max Mustermann",
		Email:    "max@musterma.nn",
		Password: "changeme123",
		UrlHost:  "localhost:1234",
	}

	uuid, err := store.Create(u)
	if err != nil {
		t.Fatal(err)
	}

	u, err = store.FindByUuid(uuid)
	if u == nil {
		t.Fatalf("Expected u == nil, got %v", u)
	}
	if len(u.Uuid) == 0 {
		t.Fatalf("Expected len(u.Uuid) == 0, got %v", len(u.Uuid))
	}
	if len(u.Password) > 0 {
		t.Fatalf("Expected len(u.Password) > 0, got %v", len(u.Password))
	}
}

func Test_UserStore_FailingToCreateUserDueToShortEmail(t *testing.T) {

	tx := helpers.GetDbTx(t)
	defer tx.Rollback()

	store := stores.NewDbUserStore(tx, helpers.GetConfig(t))

	u := &domain.User{
		Name:     "Max Mustermann",
		Email:    "",
		Password: "changeme123",
		UrlHost:  "localhost:1234",
	}

	_, err := store.Create(u)
	if ve, ok := err.(*domain.ValidationError); ok {
		if ve.Errors["email"][0] != "too_short" {
			t.Fatal("Expected err.Errors[\"email\"] to be `too_short' got:", ve.Errors["email"])
		}
	} else {
		t.Fatal("Expected to get a domain.ValidationError, got:", err)
	}

}

func Test_UserStore_FailingToCreateUserDueToShortPassword(t *testing.T) {

	tx := helpers.GetDbTx(t)
	defer tx.Rollback()

	store := stores.NewDbUserStore(tx, helpers.GetConfig(t))

	u := &domain.User{
		Name:     "Max Mustermann",
		Email:    "max@musterma.nn",
		Password: "shortie",
		UrlHost:  "localhost:1234",
	}

	_, err := store.Create(u)

	if ve, ok := err.(*domain.ValidationError); ok {
		if ve.Errors["password"][0] != "too_short" {
			t.Fatal("Expected err.Errors[\"password\"] to be `too_short' got:", ve.Errors["password"])
		}
	} else {
		t.Fatal("Expected to get a domain.ValidationError, got:", err)
	}
}

func Test_UserStore_FailingToCreateUserDueToEmailConflict(t *testing.T) {

	tx := helpers.GetDbTx(t)
	defer tx.Rollback()

	store := stores.NewDbUserStore(tx, helpers.GetConfig(t))

	u := &domain.User{
		Name:     "Max Mustermann",
		Email:    "max@musterma.nn",
		Password: "changeme123",
		UrlHost:  "localhost:1234",
	}

	_, err := store.Create(u)
	if err != nil {
		t.Fatal(err)
	}

	u.Uuid = "" // Forces the UUID to be re-generated, the first Create blesses a user with one.
	_, err = store.Create(u)
	if ve, ok := err.(*domain.ValidationError); ok {
		if ve.Errors["email"][0] != "unique_violation" {
			t.Fatal("Expected err.Errors[\"email\"] to be `unique_violation' got: ", ve.Errors["email"])
		}
	} else {
		t.Fatal("Expected to get a domain.ValidationError, got:", err)
	}

}

func Test_UserStore_FailingToUpdateAnUnsavedUser(t *testing.T) {

	tx := helpers.GetDbTx(t)
	defer tx.Rollback()

	store := stores.NewDbUserStore(tx, helpers.GetConfig(t))

	u := &domain.User{
		Name:     "Max Mustermann",
		Email:    "max@musterma.nn",
		Password: "changeme123",
		UrlHost:  "localhost:1234",
	}

	err := store.Update(u)
	if nerr, ok := err.(*domain.NotFoundError); !ok {
		t.Errorf("err.(type) = %T; want %T", err, nerr)
	}
}

func Test_UserStore_UpdatingAUserSuccessfully(t *testing.T) {

	tx := helpers.GetDbTx(t)
	defer tx.Rollback()

	store := stores.NewDbUserStore(tx, helpers.GetConfig(t))

	u := &domain.User{
		Name:     "Max Mustermann",
		Email:    "max@musterma.nn",
		Password: "changeme123",
		UrlHost:  "localhost:1234",
	}

	uuid, err := store.Create(u)
	if err != nil {
		t.Fatal(err)
	}

	u, err = store.FindByUuid(uuid)
	if err != nil {
		t.Fatal(err)
	}

	u.Name = "Anne Mustermann"
	err = store.Update(u)
	if err != nil {
		t.Fatalf("failed to save first user: %#v", err)
	}

	u, err = store.FindByUuid(uuid)
	if err != nil {
		t.Fatal(err)
	}
	if u.Name != "Anne Mustermann" {
		t.Fatal("Name was not saved")
	}

}

func Test_UserStore_FailingToFindAUserByUuid(t *testing.T) {

	tx := helpers.GetDbTx(t)
	defer tx.Rollback()

	store := stores.NewDbUserStore(tx, helpers.GetConfig(t))

	_, err := store.FindByUuid("11111111-1111-4111-a111-111111119999")
	if _, ok := err.(*domain.NotFoundError); !ok {
		t.Fatal("Expected to get a *domain.NotFoundError, got:", err)
	}

}

func Test_UserStore_FailingToUpdateANotFoundUser(t *testing.T) {

	tx := helpers.GetDbTx(t)
	defer tx.Rollback()

	store := stores.NewDbUserStore(tx, helpers.GetConfig(t))

	u := &domain.User{
		Uuid:  "11111111-1111-4111-a111-111111111111",
		Email: "anne@musterma.nn",
		Name:  "Anne Mustermann",
	}

	err := store.Update(u)
	if _, ok := err.(*domain.NotFoundError); !ok {
		t.Fatal("Expected to get a *domain.NotFoundError, got:", err)
	}

}

func Test_UserStore_FailingToUpdateAUserDueToShortEmail(t *testing.T) {

	tx := helpers.GetDbTx(t)
	defer tx.Rollback()

	store := stores.NewDbUserStore(tx, helpers.GetConfig(t))

	u := &domain.User{
		Uuid: "11111111-1111-4111-a111-111111111111",
		Name: "Anne Mustermann",
	}

	err := store.Update(u)
	if _, ok := err.(*domain.ValidationError); !ok {
		t.Fatal("Expected to get a domain.ValidationError, got:", err)
	}

}

func Test_UserStore_FailingToUpdateAUserDueToShortPassword(t *testing.T) {

	tx := helpers.GetDbTx(t)
	defer tx.Rollback()

	store := stores.NewDbUserStore(tx, helpers.GetConfig(t))

	u := &domain.User{
		Name:     "Anne Mustermann",
		Email:    "anne@musuterma.nn",
		Password: "changeme123",
		UrlHost:  "localhost:1234",
	}

	uuid, err := store.Create(u)
	if err != nil {
		t.Fatal(err)
	}

	u.Uuid = uuid
	u.Password = "shortie" // must be len() > 0

	err = store.Update(u)
	if _, ok := err.(*domain.ValidationError); !ok {
		t.Fatal("Expected to get a domain.ValidationError, got:", err)
	}

}

func Test_UserStore_SucessfullyFindingAUserByEmailAndPassword(t *testing.T) {

	tx := helpers.GetDbTx(t)
	defer tx.Rollback()

	store := stores.NewDbUserStore(tx, helpers.GetConfig(t))

	u := &domain.User{
		Name:     "Anne Mustermann",
		Email:    "anne@musterma.nn",
		Password: "0123456789",
		UrlHost:  "localhost:1234",
	}

	storedUuid, err := store.Create(u)
	if err != nil {
		t.Fatal(err)
	}

	foundUuid, err := store.FindUuidByEmailAddressAndPassword("anne@musterma.nn", "0123456789")
	if err != nil {
		t.Fatal(err)
	}

	if storedUuid != foundUuid {
		t.Fatal("UUIDs don't match", storedUuid, foundUuid)
	}

}

func Test_UserStore_FailingToFindAUserDueToIncorrectPassword(t *testing.T) {

	tx := helpers.GetDbTx(t)
	defer tx.Rollback()

	store := stores.NewDbUserStore(tx, helpers.GetConfig(t))

	u := &domain.User{
		Name:     "Anne Mustermann",
		Email:    "anne@musterma.nn",
		Password: "0123456789",
		UrlHost:  "localhost:1234",
	}

	_, err := store.Create(u)
	if err != nil {
		t.Fatal(err)
	}

	_, err = store.FindUuidByEmailAddressAndPassword("anne@musterma.nn", "00000000")
	if _, ok := err.(*domain.NotFoundError); !ok {
		t.Fatal("Expected to get a *domain.NotFoundError, got:", err)
	}

}

func Test_UserStore_FailingToFindAUserDueToWithoutPasswordSet(t *testing.T) {

	tx := helpers.GetDbTx(t)
	defer tx.Rollback()

	store := stores.NewDbUserStore(tx, helpers.GetConfig(t))

	u := &domain.User{
		Name:            "Anne Mustermann",
		Email:           "anne@musterma.nn",
		UrlHost:         "localhost:1234",
		WithoutPassword: true,
	}

	_, err := store.Create(u)
	if err != nil {
		t.Fatal(err)
	}

	_, err = store.FindUuidByEmailAddressAndPassword("anne@musterma.nn", "00000000")
	if _, ok := err.(*domain.NotFoundError); !ok {
		t.Fatal("Expected to get a *domain.NotFoundError, got:", err)
	}

}

func Test_UserStore_FindAllUsersForProject(t *testing.T) {
	tx := helpers.GetDbTx(t)
	defer tx.Rollback()

	store := stores.NewDbUserStore(tx, helpers.GetConfig(t))
	pw := "long-password"

	u1 := helpers.MustCreateUser(t, tx, &domain.User{
		Email:    "a@foo",
		Password: pw,
	})
	u2 := helpers.MustCreateUser(t, tx, &domain.User{
		Email:    "b@foo",
		Password: pw,
	})

	o := helpers.MustCreateOrganization(t, tx, &domain.Organization{Name: "Example"})

	helpers.MustCreateOrganizationMembership(t, tx, &domain.OrganizationMembership{
		OrganizationUuid: o.Uuid,
		UserUuid:         u1.Uuid,
		Type:             "member",
	})

	helpers.MustCreateOrganizationMembership(t, tx, &domain.OrganizationMembership{
		OrganizationUuid: o.Uuid,
		UserUuid:         u2.Uuid,
		Type:             "member",
	})

	p := helpers.MustCreateProject(t, tx, &domain.Project{
		Name:             "Example Project 1",
		OrganizationUuid: o.Uuid,
	})

	users, err := store.FindAllForProject(p.Uuid)
	if err != nil {
		t.Fatal(err)
	}

	expected := []string{u1.Email, u2.Email}
	actual := []string{users[0].Email, users[1].Email}
	sort.Strings(actual)

	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("Expected %s, got %s", expected, actual)
	}
}

func Test_UserStore_FindAllSubscribers(t *testing.T) {
	tx := helpers.GetDbTx(t)
	defer tx.Rollback()

	store := stores.NewDbUserStore(tx, helpers.GetConfig(t))
	subscriptions := stores.NewDbSubscriptionStore(tx)
	world := helpers.MustNewWorld(tx, t)

	job := &domain.Job{
		Uuid: "164b8a0a-bfa4-4c23-82b6-e8c555bb44b1",
	}

	world.User("default").SubscribeTo(job, domain.EventOperationStarted, subscriptions)
	world.User("other").SubscribeTo(job, domain.EventOperationStarted, subscriptions)

	subscribers, err := store.FindAllSubscribers(job.Id(), domain.EventOperationStarted)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(subscribers, []*domain.User{
		world.User("default"),
		world.User("other"),
	}) && !reflect.DeepEqual(subscribers, []*domain.User{
		world.User("other"),
		world.User("default"),
	}) {
		t.Fatalf("Expected subscribers to include %q and %q, got %#v\n", "default", "other", subscribers)
	}
}

func Test_UserStore_Update_savesTotpEnabledAt(t *testing.T) {
	tx := helpers.GetDbTx(t)
	defer tx.Rollback()

	store := stores.NewDbUserStore(tx, helpers.GetConfig(t))
	world := helpers.MustNewWorld(tx, t)
	user := world.User("default")
	user.GenerateTotpSecret()
	user.EnableTotp(user.CurrentTotpToken())
	if err := store.Update(user); err != nil {
		t.Fatal(err)
	}

	got := (*time.Time)(nil)
	if err := tx.Get(&got, `SELECT totp_enabled_at FROM users WHERE uuid = $1`, user.Uuid); err != nil {
		t.Fatal(err)
	}

	if got == nil {
		t.Fatalf("totp_enabled_at = nil; want not nil")
	}
}

func Test_UserStore_RemovesWithoutPasswordOnUpdate(t *testing.T) {

	tx := helpers.GetDbTx(t)
	defer tx.Rollback()

	store := stores.NewDbUserStore(tx, helpers.GetConfig(t))

	u := &domain.User{
		Name:            "Anne Mustermann",
		Email:           "anne@musterma.nn",
		UrlHost:         "localhost:1234",
		WithoutPassword: true,
	}

	_, err := store.Create(u)
	if err != nil {
		t.Fatal(err)
	}

	u.Password = "1234567890"
	err = store.Update(u)
	if err != nil {
		t.Fatal(err)
	}

	_, err = store.FindUuidByEmailAddressAndPassword("anne@musterma.nn", "1234567890")
	if err != nil {
		t.Fatal(err)
	}

}
