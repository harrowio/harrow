package stores_test

import (
	"testing"

	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/stores"
	helpers "github.com/harrowio/harrow/test_helpers"
)

func TestScheduleStore_DisableSchedule(t *testing.T) {
	tx := helpers.GetDbTx(t)
	world := helpers.MustNewWorld(tx, t)
	defer tx.Rollback()

	nowString := "now"
	schedule := helpers.MustCreateSchedule(t, tx, &domain.Schedule{
		UserUuid:        world.User("default").Uuid,
		JobUuid:         world.Job("default").Uuid,
		Description:     "testing",
		Timespec:        &nowString,
		Disabled:        nil,
		DisabledBecause: nil,
	})

	scheduleStore := stores.NewDbScheduleStore(tx)
	errorString := "error"
	scheduleStore.DisableSchedule(schedule.Uuid, domain.ScheduleDisabledInternalError, &errorString)

	schedule, err := scheduleStore.FindByUuid(schedule.Uuid)
	if err != nil {
		t.Fatal("Could not load Schedule:", err)
	}
	if *schedule.Disabled != domain.ScheduleDisabledInternalError || *schedule.DisabledBecause != errorString {
		t.Fatalf("Schedule %#v was not disabled due to the internal error '%s'\n", schedule, errorString)
	}
}
