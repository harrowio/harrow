package projector

import (
	"fmt"

	"github.com/harrowio/harrow/domain"
)

type Operation struct {
	Uuid    string
	JobUuid string
}

func (self *Operation) ProjectUuid(index IndexTransaction) string {
	job := &Job{}
	if err := index.Get(self.JobUuid, job); err != nil {
		panic(fmt.Sprintf("Operation(%s).ProjectUuid(): %s", self.Uuid, err))
	}

	task := &Task{}
	if err := index.Get(job.TaskUuid, task); err != nil {
		panic(fmt.Sprintf("Operation(%s).ProjectUuid(): %s", self.Uuid, err))
	}

	return task.ProjectUuid
}

type Operations struct {
	ByUuid map[string]*Operation
}

func NewOperations() *Operations {
	return &Operations{}
}

func (self *Operations) SubscribedTo() []string {
	return []string{"operation.started"}
}

func (self *Operations) HandleActivity(tx IndexTransaction, activity *domain.Activity) error {
	switch activity.Name {
	case "operation.started":
		operation, ok := activity.Payload.(*domain.Operation)
		if !ok {
			return NewTypeError(activity, operation)
		}

		if operation.JobUuid == nil {
			return nil
		}

		projectedOperation := &Operation{
			Uuid:    operation.Uuid,
			JobUuid: *operation.JobUuid,
		}

		return tx.Put(operation.Uuid, projectedOperation)
	}

	return nil
}
