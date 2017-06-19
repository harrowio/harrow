package stencil

import (
	"fmt"

	"github.com/harrowio/harrow/domain"
)

// ScriptEditorDefaults sets up defaults for supporting operation of
// the script editor.  This is necessary for scheduling operations in
// order to test any scripts, when the user doesn't have any tasks
// yet.
type ScriptEditorDefaults struct {
	conf        *Configuration
	task        *domain.Task
	environment *domain.Environment
	job         *domain.Job
}

func NewScriptEditorDefaults(configuration *Configuration) *ScriptEditorDefaults {
	return &ScriptEditorDefaults{
		conf: configuration,
	}
}

// Create creates the following objects for this stencil:
//
// Environments: default
//
// Tasks: default
func (self *ScriptEditorDefaults) Create() error {
	errors := NewError()

	project, err := self.conf.Projects.FindProject(self.conf.ProjectUuid)
	if err != nil {
		return errors.Add("FindProject", project, err)
	}

	environmentUrn := fmt.Sprintf("urn:harrow:default-environment:%s", self.conf.ProjectUuid)
	defaultEnvironment, err := self.conf.Environments.FindEnvironmentByName(environmentUrn)
	if err != nil && domain.IsNotFound(err) {
		defaultEnvironment = project.NewEnvironment(environmentUrn)
		if err := self.conf.Environments.CreateEnvironment(defaultEnvironment); err != nil {
			errors.Add("CreateEnvironment", defaultEnvironment, err)
		}
	} else if err != nil {
		errors.Add("FindEnvironmentByName", environmentUrn, err)
	}
	self.environment = defaultEnvironment

	taskUrn := fmt.Sprintf("urn:harrow:default-task:%s", self.conf.ProjectUuid)
	defaultTask, err := self.conf.Tasks.FindTaskByName(taskUrn)
	if err != nil && domain.IsNotFound(err) {
		defaultTask = project.NewTask(taskUrn, self.DefaultTaskBody())
		if err := self.conf.Tasks.CreateTask(defaultTask); err != nil {
			errors.Add("CreateTask", defaultTask, err)
		}
	} else if err != nil {
		errors.Add("FindTaskByName", taskUrn, err)
	}
	self.task = defaultTask

	if self.environment != nil && self.task != nil {
		jobUrn := fmt.Sprintf("urn:harrow:default-job:%s", self.conf.ProjectUuid)
		defaultJob, err := self.conf.Jobs.FindJobByName(jobUrn)
		if err != nil && domain.IsNotFound(err) {
			defaultJob = project.NewJob(jobUrn, defaultTask.Uuid, defaultEnvironment.Uuid)
			if err := self.conf.Jobs.CreateJob(defaultJob); err != nil {
				errors.Add("CreateJob", defaultJob, err)
			}
		} else if err != nil {
			errors.Add("FindJobByName", jobUrn, err)
		}
		self.job = defaultJob
	}

	return errors.ToError()
}

// Job returns the default job as created by this stencil.
func (self *ScriptEditorDefaults) Job() *domain.Job {
	return self.job
}

// Environment returns the default environment as created by this
// stencil.
func (self *ScriptEditorDefaults) Environment() *domain.Environment {
	return self.environment
}

// Task returns the default task as created by this stencil.
func (self *ScriptEditorDefaults) Task() *domain.Task {
	return self.task
}

func (self *ScriptEditorDefaults) DefaultTaskBody() string {
	return `#!/bin/bash -e

`
}
