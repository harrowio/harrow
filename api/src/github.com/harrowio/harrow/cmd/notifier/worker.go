package notifier

import (
	"fmt"

	"github.com/harrowio/harrow/domain"
)

type NotificationRulesStore interface {
	FindByProjectUuid(projectUuid string) ([]*domain.NotificationRule, error)
}

type Scheduler interface {
	ScheduleNotification(rule *domain.NotificationRule, activity *domain.Activity) error
}

type Worker struct {
	notificationRules NotificationRulesStore
	scheduler         Scheduler
}

func NewWorker(rules NotificationRulesStore, scheduler Scheduler) *Worker {
	return &Worker{
		notificationRules: rules,
		scheduler:         scheduler,
	}
}

func (self *Worker) HandleActivity(activity *domain.Activity) error {
	projectUuid := activity.ProjectUuid()
	if projectUuid == "" {
		return nil
	}

	possibleRules, err := self.notificationRules.FindByProjectUuid(projectUuid)
	if err != nil {
		return err
	}

	for _, rule := range possibleRules {
		if rule.Matches(activity) {
			self.Notify(rule, activity)
		}
	}
	return nil
}

func (self *Worker) Notify(rule *domain.NotificationRule, activity *domain.Activity) {
	fmt.Printf("notify: rule=%s activity=%s@%d\n", rule.Uuid, activity.Name, activity.Id)
	if err := self.scheduler.ScheduleNotification(rule, activity); err != nil {
		fmt.Printf("notify: %s\n", err)
	}
}
