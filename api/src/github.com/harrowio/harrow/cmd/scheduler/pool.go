package scheduler

import (
	"fmt"
	"sync"

	"github.com/harrowio/harrow/bus/activity"
	"github.com/harrowio/harrow/domain"
	"github.com/rs/zerolog/log"

	"github.com/jmoiron/sqlx"
)

type Pool struct {
	sync.Mutex
	DB      *sqlx.DB
	Members map[string]LiveSchedule
}

func (self *Pool) Add(db *sqlx.DB, s domain.Schedulable, activityBus activity.Sink) {

	liveSchedule := LiveSchedule{Schedule: s, ActivityBus: activityBus}

	nxtOp, err := liveSchedule.Schedule.NextOperation(nil)
	if nxtOp.IsZero() || err != nil {
		log.Error().Msgf("%s: %s", s.Id(), err)
		return
	}

	//
	// Monitor returns (nothing) when the schedule becomes invalid
	// so we can use this chance to strip it of it's
	//
	go func(uuid string, db *sqlx.DB) {
		liveSchedule.Monitor(db)
		log.Debug().Msg("Monitor stopped monitoring, removing schedule")
		log.Debug().Msg("pool.Add obtaining lock")
		self.Remove(uuid)
		log.Debug().Msg("pool.Add[d] releasing lock")
	}(s.Id(), db)

	log.Debug().Msg("pool.Add obtaining lock")
	self.Lock()
	self.Members[s.Id()] = liveSchedule
	log.Debug().Msg("pool.Add[d] releasing lock")
	log.Info().Msgf("pool.Add %q", s.Id())
	self.Unlock()

}

func (self *Pool) Remove(uuid string) {
	log.Debug().Msg("pool.Remove obtaining lock")
	self.Lock()
	defer func(s *Pool) {
		log.Debug().Msg("pool.Remove[d] releasing lock")
		s.Unlock()
	}(self)
	log.Info().Msgf("pool.Remove %q", uuid)
	log.Debug().Msg("pool.Remove deleting internal struct member")
	delete(self.Members, uuid)
}

func (self *Pool) ToString() string {
	return fmt.Sprintf("Pool Size: %d", len(self.Members))
}
