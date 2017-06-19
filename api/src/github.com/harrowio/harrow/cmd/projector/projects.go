package projector

import "github.com/harrowio/harrow/domain"

type Project struct {
	Uuid string
	Name string
}

type Projects struct {
}

func NewProjects() *Projects {
	return &Projects{}
}

func (self *Projects) SubscribedTo() []string {
	return []string{"project.created"}
}

func (self *Projects) HandleActivity(tx IndexTransaction, activity *domain.Activity) error {
	switch activity.Name {
	case "project.created":
		project, ok := activity.Payload.(*domain.Project)
		if !ok {
			return NewTypeError(activity, project)
		}
		projectedProject := Project{
			Uuid: project.Uuid,
			Name: project.Name,
		}
		return tx.Put(project.Uuid, projectedProject)
	}
	return nil
}
