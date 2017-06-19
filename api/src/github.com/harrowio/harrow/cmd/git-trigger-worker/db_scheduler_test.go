package gitTriggerWorker

import (
	"testing"

	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/stores"
	"github.com/harrowio/harrow/test_helpers"
	"github.com/jmoiron/sqlx"
)

type SchedulerTestData struct {
	t              *testing.T
	trigger        *domain.GitTrigger
	finishTxCalled bool
}

func NewSchedulerTestData(t *testing.T, trigger *domain.GitTrigger) *SchedulerTestData {
	return &SchedulerTestData{
		t:       t,
		trigger: trigger,
	}
}

func (self *SchedulerTestData) InitTx(tx *sqlx.Tx) error {
	world := test_helpers.MustNewWorld(tx, self.t)
	project := world.Project("public")
	task := world.Task("default")
	environment := world.Environment("astley")
	job := project.NewJob("test-job", task.Uuid, environment.Uuid)
	job.Uuid = self.trigger.JobUuid
	user := domain.NewUser("test user", "test@localhost", "long-enough-password")
	user.Uuid = self.trigger.CreatorUuid
	test_helpers.MustCreateUser(self.t, tx, user)
	test_helpers.MustCreateJob(self.t, tx, job)
	return nil
}

func (self *SchedulerTestData) FinishTx(tx *sqlx.Tx) error {
	defer tx.Rollback()
	schedules, err := stores.NewDbScheduleStore(tx).FindAllByJobUuid(self.trigger.JobUuid)
	if err != nil {
		return err
	}

	if got, want := len(schedules), 1; got != want {
		self.t.Fatalf(`len(schedules) = %v; want %v`, got, want)
	}

	if got := schedules[0].Timespec; got == nil {
		self.t.Fatalf(`schedules[0].Timespec is nil`)
	}

	if got, want := *schedules[0].Timespec, "now"; got != want {
		self.t.Errorf(`*schedules[0].Timespec = %v; want %v`, got, want)
	}

	if got, want := schedules[0].UserUuid, self.trigger.CreatorUuid; got != want {
		self.t.Errorf(`schedules[0].CreatorUuid = %v; want %v`, got, want)
	}

	if got, want := schedules[0].Parameters.TriggeredByGitTrigger, self.trigger.Uuid; got != want {
		self.t.Errorf(`schedules[0].Parameters.TriggeredByGitTrigger = %v; want %v`, got, want)
	}

	if got, want := schedules[0].Parameters.Reason, domain.OperationTriggeredByGitTrigger; got != want {
		self.t.Errorf(`schedules[0].Parameters.Reason = %v; want %v`, got, want)
	}

	if got, want := schedules[0].Parameters.GitTriggerName, self.trigger.Name; got != want {
		self.t.Errorf(`schedules[0].Parameters.GitTriggerName = %v; want %v`, got, want)
	}

	self.finishTxCalled = true

	return nil
}

func TestDbScheduler_ScheduleJob_createsAOneTimeScheduleForNowToRunTheSpecifiedJob(t *testing.T) {
	jobUuid := "0381257f-2d3c-4503-b5f8-c4e4940dde87"
	creatorUuid := "04fb1e8a-d966-4c64-ac28-b95418d3a1cf"
	triggerUuid := "28d4e271-3374-419e-9bfc-38afda36952f"
	trigger := &domain.GitTrigger{
		Uuid:        triggerUuid,
		CreatorUuid: creatorUuid,
		JobUuid:     jobUuid,
		Name:        "A git trigger",
	}
	schedulerTestData := NewSchedulerTestData(t, trigger)
	scheduler := NewDbScheduler(db).
		InitTxWith(schedulerTestData.InitTx).
		FinishTxWith(schedulerTestData.FinishTx)

	scheduler.ScheduleJob(trigger, domain.NewOperationParameters())

	if got, want := schedulerTestData.finishTxCalled, true; got != want {
		t.Errorf(`schedulerTestData.finishTxCalled = %v; want %v`, got, want)
	}
}
