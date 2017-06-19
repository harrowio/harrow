package projector

import "github.com/harrowio/harrow/domain"

type Task struct {
	Uuid        string
	ProjectUuid string
	Name        string
}

type Tasks struct {
}

func NewTasks() *Tasks {
	return &Tasks{}
}

func (self *Tasks) SubscribedTo() []string {
	return []string{"task.added", "task.edited"}
}

func (self *Tasks) HandleActivity(tx IndexTransaction, activity *domain.Activity) error {
	switch activity.Name {
	case "task.added", "task.edited":
		task, ok := activity.Payload.(*domain.Task)
		if !ok {
			return NewTypeError(activity, task)
		}
		projectedTask := &Task{
			Uuid:        task.Uuid,
			Name:        task.Name,
			ProjectUuid: task.ProjectUuid,
		}
		return tx.Put(task.Uuid, projectedTask)
	}

	return nil
}
