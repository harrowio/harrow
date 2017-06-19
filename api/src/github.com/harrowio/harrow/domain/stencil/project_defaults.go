package stencil

type ProjectDefaults struct {
	conf *Configuration
}

func NewProjectDefaults(configuration *Configuration) *ProjectDefaults {
	return &ProjectDefaults{
		conf: configuration,
	}
}

// Create creates the default environment in the project given by the
// configured ProjectUuid.
func (self *ProjectDefaults) Create() error {
	errors := NewError()
	project, err := self.conf.Projects.FindProject(self.conf.ProjectUuid)
	if err != nil {
		errors.Add("FindProject", self.conf.ProjectUuid, err)
		return errors.ToError()
	}

	environment := project.NewDefaultEnvironment()
	if err := self.conf.Environments.CreateEnvironment(environment); err != nil {
		errors.Add("CreateEnvironment", environment, err)
	}

	return errors.ToError()
}
