package gitTriggerWorker

import (
	"strings"

	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/logger"
)

type TriggerIndex interface {
	FindTriggersForActivity(*domain.Activity) ([]*domain.GitTrigger, error)
}

type Scheduler interface {
	ScheduleJob(forTrigger *domain.GitTrigger, params *domain.OperationParameters) error
}

type GitTriggerWorker struct {
	triggers  TriggerIndex
	scheduler Scheduler
	log       logger.Logger
}

func NewGitTriggerWorker(triggerIndex TriggerIndex, scheduler Scheduler) *GitTriggerWorker {
	return &GitTriggerWorker{
		triggers:  triggerIndex,
		scheduler: scheduler,
	}
}

func (self *GitTriggerWorker) Log() logger.Logger {
	if self.log == nil {
		self.log = logger.Discard
	}
	return self.log
}

func (self *GitTriggerWorker) SetLogger(l logger.Logger) {
	self.log = l
}

// HandleActivity schedules a job for activity if any matching git
// trigger is found in the git trigger index.
func (self *GitTriggerWorker) HandleActivity(activity *domain.Activity) error {
	self.Log().Info().Msgf("Handling %s@%d\n", activity.Name, activity.Id)
	triggers, err := self.triggers.FindTriggersForActivity(activity)
	if _, ok := err.(*domain.NotFoundError); ok {
		return nil
	}

	if err != nil {
		return err
	}
	self.log.Debug().Msgf("Checking against %d triggers", len(triggers))
	for _, trigger := range triggers {
		if trigger.Match(activity) {
			self.log.Info().Msgf("match trigger=%q activity=%d\n", trigger.Uuid, activity.Id)
			err := self.scheduler.ScheduleJob(trigger, OperationParametersForActivity(activity))
			if err != nil {
				self.log.Error().Msgf("scheduler.ScheduleJob: %s\n", err)
			}
		}
	}

	return nil
}

func OperationParametersForActivity(activity *domain.Activity) *domain.OperationParameters {
	symbolicRef := ""
	repositoryUuid := ""
	switch payload := activity.Payload.(type) {
	case *domain.RepositoryRef:
		symbolicRef = payload.Symbolic
		repositoryUuid = payload.RepositoryUuid
	case *domain.ChangedRepositoryRef:
		symbolicRef = payload.Symbolic
		repositoryUuid = payload.RepositoryUuid
	default:
		return domain.NewOperationParameters()
	}

	params := domain.NewOperationParameters()
	params.Checkout[repositoryUuid] = strings.Replace(symbolicRef, "refs/heads/", "", 1)

	return params
}
