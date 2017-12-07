package scheduler

import (
	"github.com/harrowio/harrow/bus/activity"
	"github.com/harrowio/harrow/bus/broadcast"
	"github.com/harrowio/harrow/config"
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/stores"
	"github.com/rs/zerolog/log"

	"github.com/jmoiron/sqlx"
)

func schedulableFromUuid(db *sqlx.DB, uuid string) (domain.Schedulable, error) {
	tx, _ := db.Beginx()
	defer tx.Rollback()
	store := stores.NewDbScheduleStore(tx)
	log.Debug().Msgf("schedulableFromUuid: loading schedule %q", uuid)
	sched, err := store.FindInclArchivedByUuid(uuid)
	if err != nil {
		return nil, err
	}

	if sched.Timespec != nil {
		log.Debug().Msgf("timespec: %q", *sched.Timespec)
	}
	if sched.Cronspec != nil {
		log.Debug().Msgf("cronspec: %q", *sched.Cronspec)
	}

	return domain.NewSchedulable(sched)
}

func runOrPool(db *sqlx.DB, pool *Pool, s domain.Schedulable, activityBus activity.Sink) {

	if s.IsRecurring() {
		pool.Add(db, s, activityBus)
	} else {
		liveSchedule := &LiveSchedule{Schedule: s, ActivityBus: activityBus}
		if wait := liveSchedule.nextCycle(liveSchedule.WaitDuration()); wait != nil {
			<-wait
		}

		// The schedule might have been disabled in the meantime
		if reloadedSchedulable, err := schedulableFromUuid(db, s.Id()); err != nil {
			log.Error().Msgf("Failed to reload schedule %s, operation won't run: %s", s.Id(), err)
			return
		} else {
			liveSchedule.Schedule = reloadedSchedulable
		}
		err := liveSchedule.CreateOperation(db)
		if err == nil {

			tx, err := db.Beginx()
			if err != nil {
				log.Fatal().Msgf("runOrPool Schedule:%s: db.Beginx: %s\n", s.Id(), err)
			}
			defer tx.Rollback()
			scheduleStore := stores.NewDbScheduleStore(tx)
			err = scheduleStore.DisableSchedule(s.Id(), domain.ScheduleDisabledRanOnce, nil)
			if err != nil {
				log.Error().Msgf("runOrPool Schedule:%s: unable to DisableSchedule: %s\n", s.Id(), err)
			} else {
				tx.Commit()
			}
		}
	}
}

func handleCreates(db *sqlx.DB, pool *Pool, messages <-chan broadcast.Message, activityBus activity.Sink) {
	for msg := range messages {
		if msg.Table() != "schedules" {
			msg.RejectForever()
			continue
		}

		s, err := schedulableFromUuid(db, msg.UUID())
		if err != nil {
			log.Error().Msgf("handleCreates: Error creating schedulable: %s (%s)\n", msg.UUID(), err)
			msg.RejectForever()
			continue
		}

		msg.Acknowledge()

		go runOrPool(db, pool, s, activityBus)
	}
}

func handleChanges(db *sqlx.DB, pool *Pool, messages <-chan broadcast.Message, activityBus activity.Sink) {
	for msg := range messages {
		if msg.Table() != "schedules" {
			msg.RejectForever()
			continue
		}

		pool.Remove(msg.UUID())

		s, err := schedulableFromUuid(db, msg.UUID())
		if err != nil && err.Error() == domain.ScheduleDisabledRanOnce {
			log.Info().Msgf("handleChanges: schedule %q ran already", msg.UUID())
			msg.Acknowledge()
			continue
		}
		if err != nil {
			log.Error().Msgf("handleChanges: Error creating schedulable: %s (%s)\n", msg.UUID(), err)
			msg.RejectForever()
			continue
		}

		msg.Acknowledge()

		if !s.IsDisabled() {
			runOrPool(db, pool, s, activityBus)
		}
	}
}

const ProgramName = "scheduler"

func Main() {
	c := config.GetConfig()

	//
	// Regular database connection used for the world snapshot
	//
	db, err := c.DB()
	if err != nil {
		log.Fatal().Err(err)
	}

	activityBus := activity.NewAMQPTransport(c.AmqpConnectionString(), "scheduler")
	defer activityBus.Close()

	liveSchedules := &Pool{DB: db, Members: make(map[string]LiveSchedule)}

	broadcastBus := broadcast.NewAMQPTransport(c.AmqpConnectionString(), "scheduler")
	defer broadcastBus.Close()

	creates, err := broadcastBus.Consume(broadcast.Create)
	if err != nil {
		log.Fatal().Msgf("broadcastBus.Consume(broadcast.Create): %s", err)
	}

	changes, err := broadcastBus.Consume(broadcast.Change)
	if err != nil {
		log.Fatal().Msgf("broadcastBus.Consume(broadcast.Change): %s", err)
	}

	go handleCreates(db, liveSchedules, creates, activityBus)
	go handleChanges(db, liveSchedules, changes, activityBus)

	q := `SELECT * FROM schedules WHERE disabled IS NULL AND archived_at IS NULL;`
	schedules := []*domain.Schedule{}

	err = db.Select(&schedules, q)
	if err != nil {
		panic(err)
	}
	for _, schedule := range schedules {
		s, err := domain.NewSchedulable(schedule)
		if err != nil {
			log.Error().Msgf("Error adding schedule: %s\n", err)
			continue
		}
		runOrPool(db, liveSchedules, s, activityBus)
	}
	select {}
}
