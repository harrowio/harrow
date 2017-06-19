package scheduler

import (
	"time"

	"github.com/harrowio/harrow/activities"
	"github.com/harrowio/harrow/bus/activity"
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/stores"
	"github.com/rs/zerolog/log"

	"github.com/jmoiron/sqlx"
)

const (
	harrowBaseImageUuid = "31b0127a-6d63-4d22-b32b-e1cfc04f4007"
)

type cycleChan <-chan time.Time

type LiveSchedule struct {
	Schedule    domain.Schedulable
	ActivityBus activity.Sink
}

func (self *LiveSchedule) WaitDuration() time.Duration {
	nxt, err := self.Schedule.NextOperation(nil)
	if nxt.IsZero() || err != nil {
		return 0
	}
	waitDur := nxt.Sub(time.Now().UTC())
	return waitDur
}

func (self *LiveSchedule) nextCycle(waitDur time.Duration) cycleChan {
	if waitDur < 0 {
		return nil
	}
	return time.After(waitDur)
}

func (self *LiveSchedule) Monitor(db *sqlx.DB) {

	nxtChn := make(chan cycleChan, 1)
	waitDur := self.WaitDuration()
	curChn := self.nextCycle(waitDur)

	for {
		select {
		case c := <-nxtChn:
			if c == nil {
				log.Info().Msgf("Schedule:%s: no more scheduled operations.\n", self.Schedule.Id())
				return // Bail out completely, Pool.Prune() will clean us up out of the pool map
			}
			curChn = c
		case <-curChn:
			err := self.CreateOperation(db)
			if err != nil {
				log.Info().Msgf("Schedule:%s: could not create operation: %s.\n", self.Schedule.Id(), err)
				return
			}
			nxtChn <- self.nextCycle(self.WaitDuration())
		}
	}

}

func (self *LiveSchedule) CreateOperation(db *sqlx.DB) error {
	jobId := self.Schedule.JobId()
	op := &domain.Operation{
		JobUuid:                &jobId,
		WorkspaceBaseImageUuid: harrowBaseImageUuid,
		Type:       domain.OperationTypeJobScheduled,
		TimeLimit:  900,
		Parameters: self.Schedule.OperationParameters(),
	}
	tx, err := db.Beginx()
	if err != nil {
		log.Error().Msgf("Schedule:%s: db.Beginx: %s\n", self.Schedule.Id(), err)
	}
	defer tx.Rollback()
	store := stores.NewDbOperationStore(tx)
	if op.Uuid, err = store.Create(op); err != nil {

		tx, err := db.Beginx()
		if err != nil {
			log.Error().Msgf("CreateOperation Schedule:%s: db.Beginx: %s\n", self.Schedule.Id(), err)
		}
		defer tx.Rollback()
		scheduleStore := stores.NewDbScheduleStore(tx)
		errString := err.Error()
		err = scheduleStore.DisableSchedule(self.Schedule.Id(), domain.ScheduleDisabledInternalError, &errString)
		if err != nil {
			log.Error().Msgf("CreateOperation Schedule:%s: unable to DisableSchedule: %s\n", self.Schedule.Id(), err)
		} else {
			if err := self.ActivityBus.Publish(activities.OperationScheduled(op)); err != nil {
				log.Error().Msgf("CreateOperation Schedule:%s: activityBus.Publish: %s\n", self.Schedule.Id(), err)
			}
			tx.Commit()
		}

		return err
	} else {
		if err := self.ActivityBus.Publish(activities.OperationScheduled(op)); err != nil {
			log.Error().Msgf("CreateOperation Schedule:%s: activityBus.Publish: %s\n", self.Schedule.Id(), err)
		}
		tx.Commit()
		return nil
	}
}

func (self *LiveSchedule) Terminate(c chan bool) {
	log.Info().Msgf("Schedule:%s: terminating\n", self.Schedule.Id())
	c <- true
}
