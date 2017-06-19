package projector

import (
	"sync"
	"time"

	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/logger"
)

type ActivityStore interface {
	AllByNameSince(activityNames []string, since time.Time, handler func(*domain.Activity) error) error
}

type Projector struct {
	update  *sync.Mutex
	index   Index
	log     logger.Logger
	handler ActivityHandler
}

func NewProjector(index Index, log logger.Logger) *Projector {
	self := &Projector{
		update: new(sync.Mutex),
		index:  index,
		log:    log,
	}
	self.handler = NewBroadcastHandler(log).
		Add(NewProjects()).
		Add(NewTasks()).
		Add(NewEnvironments()).
		Add(NewJobs(log)).
		Add(NewOperations()).
		Add(NewProjectCards())

	return self
}

func (self *Projector) SubscribedTo() []string {
	return self.handler.SubscribedTo()
}

func (self *Projector) HandleActivity(tx IndexTransaction, activity *domain.Activity) error {
	if activity.OccurredOn.After(self.Version(tx)) {
		if err := tx.Put("version", activity.OccurredOn); err != nil {
			self.log.Error().Msgf("msg=version_update_failed err=%s", err)
		}
	}

	return self.handler.HandleActivity(tx, activity)
}

func (self *Projector) Update(activityStore ActivityStore) error {
	self.update.Lock()
	defer self.update.Unlock()

	version := time.Time{}
	return self.index.Update(func(tx IndexTransaction) error {
		version = self.Version(tx)
		self.log.Info().Msgf("fetch=%s", version.Format(time.RFC3339))
		return activityStore.AllByNameSince(
			self.SubscribedTo(),
			version,
			func(activity *domain.Activity) error {
				return self.HandleActivity(tx, activity)
			},
		)
	})
}

func (self *Projector) Version(tx IndexTransaction) time.Time {
	version := time.Time{}
	if err := tx.Get("version", &version); err != nil {
		self.log.Error().Msgf("failed to retrieve version")
		self.log.Error().Msgf("err=%s", err)
		return time.Time{}
	}

	return version
}
