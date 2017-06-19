package projector

import (
	"fmt"

	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/logger"
)

type Job struct {
	Uuid            string
	EnvironmentUuid string
	EnvironmentName string
	TaskUuid        string
	TaskName        string
}

func (self *Job) Name() string {
	return fmt.Sprintf("%s - %s", self.EnvironmentName, self.TaskName)
}

func (self *Job) UpdateEnvironment(uuid string, index IndexTransaction) error {
	environment := &Environment{}
	if err := index.Get(uuid, environment); err != nil {
		return err
	}
	self.EnvironmentUuid = uuid
	self.EnvironmentName = environment.Name
	return nil
}

func (self *Job) UpdateTask(uuid string, index IndexTransaction) error {
	task := &Task{}
	if err := index.Get(uuid, task); err != nil {
		return err
	}
	self.TaskUuid = uuid
	self.TaskName = task.Name
	return nil
}

type Jobs struct {
	log logger.Logger
}

func NewJobs(log logger.Logger) *Jobs {
	self := &Jobs{
		log: log,
	}

	return self
}

func (self *Jobs) SubscribedTo() []string {
	return []string{"job.added", "job.edited"}
}

func (self *Jobs) HandleActivity(tx IndexTransaction, activity *domain.Activity) error {
	switch activity.Name {
	case "job.added":
		job, ok := activity.Payload.(*domain.Job)
		if !ok {
			return NewTypeError(activity, job)
		}

		projectedJob := &Job{
			Uuid:            job.Uuid,
			EnvironmentUuid: job.EnvironmentUuid,
			TaskUuid:        job.TaskUuid,
		}

		projectedJob.UpdateEnvironment(job.EnvironmentUuid, tx)
		projectedJob.UpdateTask(job.TaskUuid, tx)

		return tx.Put(job.Uuid, projectedJob)
	case "job.edited":
		job, ok := activity.Payload.(*domain.Job)
		if !ok {
			return NewTypeError(activity, job)
		}

		existingJob := &Job{}
		if err := tx.Get(job.Uuid, existingJob); err != nil {
			return err
		}

		if job.Uuid != existingJob.Uuid {
			return nil
		}

		if err := existingJob.UpdateEnvironment(job.EnvironmentUuid, tx); err != nil {
			return err
		}

		if err := existingJob.UpdateTask(job.TaskUuid, tx); err != nil {
			return err
		}

		return tx.Put(existingJob.Uuid, existingJob)
	}

	return nil
}
