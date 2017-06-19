package projector

import "github.com/harrowio/harrow/domain"

type Environment struct {
	Uuid string
	Name string
}

type Environments struct {
}

func NewEnvironments() *Environments {
	return &Environments{}
}

func (self *Environments) SubscribedTo() []string {
	return []string{"environment.added", "environmend.edited"}
}

func (self *Environments) HandleActivity(tx IndexTransaction, activity *domain.Activity) error {
	switch activity.Name {
	case "environment.added", "environment.edited":
		environment, ok := activity.Payload.(*domain.Environment)
		if !ok {
			return NewTypeError(activity, environment)
		}
		projectedEnvironment := &Environment{
			Uuid: environment.Uuid,
			Name: environment.Name,
		}
		return tx.Put(environment.Uuid, projectedEnvironment)
	}

	return nil
}
