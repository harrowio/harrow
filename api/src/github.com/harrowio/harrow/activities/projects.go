package activities

import "github.com/harrowio/harrow/domain"

func init() {
	registerPayload(ProjectCreated(&domain.Project{}))
	registerPayload(ProjectDeleted(&domain.Project{}))
}

func ProjectCreated(project *domain.Project) *domain.Activity {
	return &domain.Activity{
		Name:       "project.created",
		Extra:      map[string]interface{}{},
		OccurredOn: Clock.Now(),
		Payload:    project,
	}
}

func ProjectDeleted(project *domain.Project) *domain.Activity {
	return &domain.Activity{
		Name:       "project.deleted",
		Extra:      map[string]interface{}{},
		OccurredOn: Clock.Now(),
		Payload:    project,
	}
}
