package gitTriggerWorker

import (
	"testing"

	"github.com/harrowio/harrow/activities"
	"github.com/harrowio/harrow/test_helpers"
	"github.com/jmoiron/sqlx"

	"github.com/harrowio/harrow/config"
	"github.com/harrowio/harrow/domain"
)

var (
	c  = config.GetConfig()
	db *sqlx.DB
)

func init() {
	err := (error)(nil)
	db, err = c.DB()
	if err != nil {
		panic(err)
	}
}

type testData struct {
	t        *testing.T
	project  *domain.Project
	activity *domain.Activity
}

func (self *testData) initTx(tx *sqlx.Tx) error {
	world := test_helpers.MustNewWorld(tx, self.t)
	self.project = world.Project("public")
	self.activity.SetProjectUuid(self.project.Uuid)
	otherProject := world.Project("private")
	job := world.Job("default")
	user := world.User("default")
	publicTrigger := domain.NewGitTrigger("public", user.Uuid).
		MatchingRef(".").
		ForChangeType("add").
		ForJob(job.Uuid).
		InProject(self.project.Uuid)

	privateTrigger := domain.NewGitTrigger("private", user.Uuid).
		MatchingRef(".").
		ForChangeType("add").
		ForJob(job.Uuid).
		InProject(otherProject.Uuid)

	test_helpers.MustCreateGitTrigger(self.t, tx, publicTrigger)
	test_helpers.MustCreateGitTrigger(self.t, tx, privateTrigger)

	return nil
}

func bootstrapTestData(t *testing.T, activity *domain.Activity) *testData {
	return &testData{
		activity: activity,
		t:        t,
	}
}

func TestDbTriggerIndex_FindTriggersForActivity_returnsNoErrorIfActivityIsNotAssociatedWithAProject(t *testing.T) {
	activity := domain.NewActivity(1, "repository-metadata.ref-changed")
	if got, want := activity.ProjectUuid(), ""; got != want {
		t.Errorf(`activity.ProjectUuid() = %v; want %v`, got, want)
	}

	index := NewDbTriggerIndex(db)
	_, err := index.FindTriggersForActivity(activity)
	if got, want := err, (error)(nil); got != want {
		t.Errorf(`err = %v; want %v`, got, want)
	}
}

func TestDbTriggerIndex_FindTriggersForActivity_returnsOnlyTriggersFromProjectAssociatedWithActivity(t *testing.T) {
	activity := activities.RepositoryMetaDataRefAdded(&domain.RepositoryRef{})
	testData := bootstrapTestData(t, activity)
	index := NewDbTriggerIndex(db).
		InitTxWith(testData.initTx)

	triggers, err := index.FindTriggersForActivity(activity)
	if err != nil {
		t.Fatal(err)
	}

	if got, want := len(triggers), 1; got != want {
		t.Fatalf(`len(triggers) = %v; want %v`, got, want)
	}

	if got, want := triggers[0].ProjectUuid, testData.project.Uuid; got != want {
		t.Errorf(`triggers[0].ProjectUuid = %v; want %v`, got, want)
	}
}
