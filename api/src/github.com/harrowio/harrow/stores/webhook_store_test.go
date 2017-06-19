package stores_test

import (
	"testing"
	"time"

	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/stores"
	"github.com/harrowio/harrow/test_helpers"
	"github.com/harrowio/harrow/uuidhelper"
)

func Test_DbWebhookStore_Create_generatesUuid(t *testing.T) {
	tx := test_helpers.GetDbTx(t)
	defer tx.Rollback()

	world := test_helpers.MustNewWorld(tx, t)
	store := stores.NewDbWebhookStore(tx)
	webhook := &domain.Webhook{
		Uuid:        "",
		ProjectUuid: world.Project("public").Uuid,
		CreatorUuid: world.User("default").Uuid,
		JobUuid:     world.Job("default").Uuid,
		Name:        "github-webhook",
		Slug:        "webhook-slug",
	}

	uuid, _ := store.Create(webhook)
	if uuid == "" {
		t.Fatalf("Expected uuid to be set")
	}
}

func Test_DbWebhookStore_Create_insertsWebhook(t *testing.T) {
	tx := test_helpers.GetDbTx(t)
	defer tx.Rollback()

	world := test_helpers.MustNewWorld(tx, t)
	store := stores.NewDbWebhookStore(tx)
	webhook := &domain.Webhook{
		Uuid:        "",
		ProjectUuid: world.Project("public").Uuid,
		CreatorUuid: world.User("default").Uuid,
		JobUuid:     world.Job("default").Uuid,
		Name:        "github-webhook",
		Slug:        "webhook-slug",
	}

	uuid, err := store.Create(webhook)
	if err != nil {
		t.Fatal(err)
	}

	found := &domain.Webhook{}
	err = tx.Get(found, `SELECT * FROM webhooks WHERE uuid = $1`, uuid)
	if err != nil {
		t.Fatal(err)
	}
}

func Test_DbWebhookStore_Create_savesJobUuid(t *testing.T) {
	tx := test_helpers.GetDbTx(t)
	defer tx.Rollback()

	world := test_helpers.MustNewWorld(tx, t)
	store := stores.NewDbWebhookStore(tx)
	webhook := &domain.Webhook{
		Uuid:        "",
		ProjectUuid: world.Project("public").Uuid,
		CreatorUuid: world.User("default").Uuid,
		JobUuid:     world.Job("default").Uuid,
		Name:        "github-webhook",
		Slug:        "webhook-slug",
	}

	uuid, err := store.Create(webhook)
	if err != nil {
		t.Fatal(err)
	}

	found := &domain.Webhook{}
	err = tx.Get(found, `SELECT * FROM webhooks WHERE uuid = $1`, uuid)
	if err != nil {
		t.Fatal(err)
	}
	if found.JobUuid != world.Job("default").Uuid {
		t.Fatal("found.JobUuid=%s, want=%s", found.JobUuid, world.Job("default"))
	}
}

func Test_DbWebhookStore_Create_failsForDuplicateSlugs(t *testing.T) {
	tx := test_helpers.GetDbTx(t)
	defer tx.Rollback()

	world := test_helpers.MustNewWorld(tx, t)
	store := stores.NewDbWebhookStore(tx)
	project := world.Project("public")
	user := world.User("default")
	job := world.Job("default")
	webhook := domain.NewWebhook(project.Uuid, user.Uuid, job.Uuid, "github")

	if _, err := store.Create(webhook); err != nil {
		t.Fatal(err)
	}

	webhook.Uuid = ""
	if _, err := store.Create(webhook); err == nil {
		t.Fatalf("Expected an error.")
	}
}

func Test_DbWebhookStore_FindByUuid_findsWebhook(t *testing.T) {
	tx := test_helpers.GetDbTx(t)
	defer tx.Rollback()

	world := test_helpers.MustNewWorld(tx, t)
	store := stores.NewDbWebhookStore(tx)
	webhook := &domain.Webhook{
		Uuid:        "",
		ProjectUuid: world.Project("public").Uuid,
		CreatorUuid: world.User("default").Uuid,
		JobUuid:     world.Job("default").Uuid,
		Name:        "github-webhook",
		Slug:        "webhook-slug",
	}

	uuid, err := store.Create(webhook)
	if err != nil {
		t.Fatal(err)
	}

	found, err := store.FindByUuid(uuid)
	if err != nil {
		t.Fatal(err)
	}

	if found.Uuid != webhook.Uuid {
		t.Fatalf("Found %q, wanted %q", found.Uuid, webhook.Uuid)
	}
}

func Test_DbWebhookStore_FindByUuid_returnsDomainNotFoundError_whenNotFound(t *testing.T) {
	tx := test_helpers.GetDbTx(t)
	defer tx.Rollback()

	world := test_helpers.MustNewWorld(tx, t)
	store := stores.NewDbWebhookStore(tx)
	webhook := &domain.Webhook{
		Uuid:        "",
		ProjectUuid: world.Project("public").Uuid,
		CreatorUuid: world.User("default").Uuid,
		JobUuid:     world.Job("default").Uuid,
		Name:        "github-webhook",
		Slug:        "webhook-slug",
	}

	_, err := store.Create(webhook)
	if err != nil {
		t.Fatal(err)
	}

	notExistingUuid := uuidhelper.MustNewV4()
	_, err = store.FindByUuid(notExistingUuid)
	if err == nil {
		t.Fatalf("Expected an error")
	}

	_, ok := err.(*domain.NotFoundError)
	if !ok {
		t.Fatalf("Expected %T, got %T", &domain.NotFoundError{}, err)
	}
}

func Test_DbWebhookStore_FindByUuid_doesNotReturnArchivedWebhooks(t *testing.T) {
	tx := test_helpers.GetDbTx(t)
	defer tx.Rollback()

	world := test_helpers.MustNewWorld(tx, t)
	store := stores.NewDbWebhookStore(tx)
	webhook := &domain.Webhook{
		Uuid:        "",
		ProjectUuid: world.Project("public").Uuid,
		CreatorUuid: world.User("default").Uuid,
		JobUuid:     world.Job("default").Uuid,
		Name:        "github-webhook",
		Slug:        "webhook-slug",
	}

	uuid, err := store.Create(webhook)
	if err != nil {
		t.Fatal(err)
	}

	tx.MustExec(`UPDATE webhooks SET archived_at = NOW() WHERE uuid = $1`, uuid)

	found, err := store.FindByUuid(uuid)

	if found != nil {
		t.Fatalf("Expected no webhook to be returned.\nGot: %#v", found)
	}

	if _, ok := err.(*domain.NotFoundError); !ok {
		t.Fatal(err)
	}
}

func Test_DbWebhookStore_FindBySlug_findsWebhook(t *testing.T) {
	tx := test_helpers.GetDbTx(t)
	defer tx.Rollback()

	world := test_helpers.MustNewWorld(tx, t)
	project := world.Project("public")
	store := stores.NewDbWebhookStore(tx)
	user := world.User("default")
	job := world.Job("default")
	webhook := domain.NewWebhook(project.Uuid, user.Uuid, job.Uuid, "github")
	if _, err := store.Create(webhook); err != nil {
		t.Fatal(err)
	}

	found, err := store.FindBySlug(webhook.Slug)
	if err != nil {
		t.Fatal(err)
	}

	if got, want := found.Slug, webhook.Slug; got != want {
		t.Fatalf("found.Slug = %q; want %q", got, want)
	}
}

func Test_DbWebhookStore_FindBySlug_returnsDomainNotFoundError_whenNotFound(t *testing.T) {
	tx := test_helpers.GetDbTx(t)
	defer tx.Rollback()

	world := test_helpers.MustNewWorld(tx, t)
	project := world.Project("public")
	store := stores.NewDbWebhookStore(tx)
	user := world.User("default")
	job := world.Job("default")
	webhook := domain.NewWebhook(project.Uuid, user.Uuid, job.Uuid, "github")
	if _, err := store.Create(webhook); err != nil {
		t.Fatal(err)
	}

	found, err := store.FindBySlug(webhook.Slug + "-does-not-exist")

	if found != nil {
		t.Fatalf("Expected no webhook to be returned.\nGot: %#v", found)
	}

	if _, ok := err.(*domain.NotFoundError); !ok {
		t.Fatal(err)
	}
}

func Test_DbWebhookStore_FindBySlug_doesNotReturnArchivedWebhooks(t *testing.T) {
	tx := test_helpers.GetDbTx(t)
	defer tx.Rollback()

	world := test_helpers.MustNewWorld(tx, t)
	project := world.Project("public")
	store := stores.NewDbWebhookStore(tx)
	user := world.User("default")
	job := world.Job("default")
	webhook := domain.NewWebhook(project.Uuid, user.Uuid, job.Uuid, "github")

	uuid, err := store.Create(webhook)
	if err != nil {
		t.Fatal(err)
	}

	tx.MustExec(`UPDATE webhooks SET archived_at = NOW() WHERE uuid = $1`, uuid)

	found, err := store.FindBySlug(webhook.Slug)

	if found != nil {
		t.Fatalf("Expected no webhook to be returned.\nGot: %#v", found)
	}

	if _, ok := err.(*domain.NotFoundError); !ok {
		t.Fatal(err)
	}
}

func Test_DbWebhookStore_ArchiveByUuid_setsArchivedAt(t *testing.T) {
	tx := test_helpers.GetDbTx(t)
	defer tx.Rollback()

	world := test_helpers.MustNewWorld(tx, t)
	store := stores.NewDbWebhookStore(tx)
	webhook := &domain.Webhook{
		Uuid:        "",
		ProjectUuid: world.Project("public").Uuid,
		CreatorUuid: world.User("default").Uuid,
		JobUuid:     world.Job("default").Uuid,
		Name:        "github-webhook",
		Slug:        "webhook-slug",
	}

	uuid, err := store.Create(webhook)
	if err != nil {
		t.Fatal(err)
	}

	err = store.ArchiveByUuid(uuid)
	if err != nil {
		t.Fatal(err)
	}

	found := &domain.Webhook{}
	err = tx.Get(found, `SELECT * FROM webhooks WHERE uuid = $1`, uuid)
	if err != nil {
		t.Fatal(err)
	}

	if found.ArchivedAt == nil {
		t.Fatal("Expected ArchivedAt to be set")
	}
}

func Test_DbWebhookStore_Update_returnsNotFoundErrorIfWebhookDoesNotExist(t *testing.T) {
	tx := test_helpers.GetDbTx(t)
	defer tx.Rollback()

	world := test_helpers.MustNewWorld(tx, t)
	store := stores.NewDbWebhookStore(tx)
	webhook := &domain.Webhook{
		Uuid:        "",
		ProjectUuid: world.Project("public").Uuid,
		CreatorUuid: world.User("default").Uuid,
		JobUuid:     world.Job("default").Uuid,
		Name:        "github-webhook",
		Slug:        "webhook-slug",
	}

	err := store.Update(webhook)
	if err == nil {
		t.Fatal("Expected an error")
	}

	if _, ok := err.(*domain.NotFoundError); !ok {
		t.Fatal(err)
	}
}

func Test_DbWebhookStore_Update_updatesMutableAttributes(t *testing.T) {
	tx := test_helpers.GetDbTx(t)
	defer tx.Rollback()

	world := test_helpers.MustNewWorld(tx, t)
	store := stores.NewDbWebhookStore(tx)
	webhook := &domain.Webhook{
		Uuid:        "",
		ProjectUuid: world.Project("public").Uuid,
		CreatorUuid: world.User("default").Uuid,
		JobUuid:     world.Job("default").Uuid,
		Name:        "github-webhook",
		Slug:        "webhook-slug",
	}
	newJob := &domain.Job{
		EnvironmentUuid: world.Environment("astley").Uuid,
		TaskUuid:        world.Task("default").Uuid,
		Name:            "new job",
	}

	jobUuid, err := stores.NewDbJobStore(tx).Create(newJob)
	if err != nil {
		t.Fatal(err)
	}

	uuid, err := store.Create(webhook)
	if err != nil {
		t.Fatal(err)
	}

	webhook.Slug = "new-slug"
	webhook.Name = "new-name"
	webhook.JobUuid = jobUuid

	err = store.Update(webhook)
	if err != nil {
		t.Fatal(err)
	}

	found, err := store.FindByUuid(uuid)
	if err != nil {
		t.Fatal(err)
	}

	if got, want := found.Slug, webhook.Slug; got != want {
		t.Errorf("found.Slug = %q; want %q", got, want)
	}

	if got, want := found.Name, webhook.Name; got != want {
		t.Errorf("found.Name = %q; want %q")
	}

	if got, want := found.JobUuid, jobUuid; got != want {
		t.Errorf("found.JobUuid = %q; want %q", got, want)
	}

}

func Test_DbWebhookStore_Update_doesNotUpdateCreatedAtOrArchivedAt(t *testing.T) {
	tx := test_helpers.GetDbTx(t)
	defer tx.Rollback()

	world := test_helpers.MustNewWorld(tx, t)
	store := stores.NewDbWebhookStore(tx)
	webhook := &domain.Webhook{
		Uuid:        "",
		ProjectUuid: world.Project("public").Uuid,
		CreatorUuid: world.User("default").Uuid,
		JobUuid:     world.Job("default").Uuid,
		Name:        "github-webhook",
		Slug:        "webhook-slug",
	}

	uuid, err := store.Create(webhook)
	if err != nil {
		t.Fatal(err)
	}

	createdAt := time.Now().Add(-1 * time.Hour)
	archivedAt := createdAt
	webhook.CreatedAt = createdAt
	webhook.ArchivedAt = &archivedAt

	err = store.Update(webhook)
	if err != nil {
		t.Fatal(err)
	}

	found, err := store.FindByUuid(uuid)
	if err != nil {
		t.Fatal(err)
	}

	if found.CreatedAt.Equal(createdAt) {
		t.Errorf("CreatedAt updated")
	}

	if found.ArchivedAt != nil {
		t.Errorf("ArchivedAt updated")
	}

}

func Test_DbWebhookStore_FindByProjectUuid_returnsWebhooksForProject(t *testing.T) {
	tx := test_helpers.GetDbTx(t)
	defer tx.Rollback()

	world := test_helpers.MustNewWorld(tx, t)
	store := stores.NewDbWebhookStore(tx)
	webhook := &domain.Webhook{
		Uuid:        "",
		ProjectUuid: world.Project("public").Uuid,
		CreatorUuid: world.User("default").Uuid,
		JobUuid:     world.Job("default").Uuid,
		Name:        "github-webhook",
		Slug:        "webhook-slug",
	}

	_, err := store.Create(webhook)
	if err != nil {
		t.Fatal(err)
	}

	found, err := store.FindByProjectUuid(webhook.ProjectUuid)
	if err != nil {
		t.Fatal(err)
	}

	if len(found) == 0 {
		t.Fatalf("Expected to find at least one Webhook")
	}

	for _, foundWebhook := range found {
		if foundWebhook.Uuid == webhook.Uuid {
			// t.Logf("Found Webhook %q", webhook.Uuid)
			break
		}
	}
}

func Test_DbWebhookStore_FindByProjectUuid_doesNotReturnArchivedWebhooks(t *testing.T) {
	tx := test_helpers.GetDbTx(t)
	defer tx.Rollback()

	world := test_helpers.MustNewWorld(tx, t)
	store := stores.NewDbWebhookStore(tx)
	webhook := &domain.Webhook{
		Uuid:        "",
		ProjectUuid: world.Project("public").Uuid,
		CreatorUuid: world.User("default").Uuid,
		JobUuid:     world.Job("default").Uuid,
		Name:        "github-webhook",
		Slug:        "webhook-slug",
	}

	_, err := store.Create(webhook)
	if err != nil {
		t.Fatal(err)
	}

	if err := store.ArchiveByUuid(webhook.Uuid); err != nil {
		t.Fatal(err)
	}

	found, err := store.FindByProjectUuid(webhook.ProjectUuid)
	if err != nil {
		t.Fatal(err)
	}

	if n := len(found); n != 0 {
		t.Fatalf("Expected %d webhooks to be found, got %d", 0, n)
	}
}

func Test_DbWebhookStore_FindByProjectUuid_doesNotReturnWebhooksCreatedByJobNotifiers(t *testing.T) {
	tx := test_helpers.GetDbTx(t)
	defer tx.Rollback()

	world := test_helpers.MustNewWorld(tx, t)
	store := stores.NewDbWebhookStore(tx)
	webhook := &domain.Webhook{
		Uuid:        "",
		ProjectUuid: world.Project("public").Uuid,
		CreatorUuid: world.User("default").Uuid,
		JobUuid:     world.Job("default").Uuid,
		Name:        "urn:harrow:job-notifier:35320fe9-7598-436c-ab0d-9ecc5ee186da",
		Slug:        "webhook-slug",
	}

	_, err := store.Create(webhook)
	if err != nil {
		t.Fatal(err)
	}

	if err := store.ArchiveByUuid(webhook.Uuid); err != nil {
		t.Fatal(err)
	}

	found, err := store.FindByProjectUuid(webhook.ProjectUuid)
	if err != nil {
		t.Fatal(err)
	}

	if n := len(found); n != 0 {
		t.Fatalf("Expected %d webhooks to be found, got %d", 0, n)
	}
}
