package activities

import "github.com/harrowio/harrow/domain"

func init() {
	registerPayload(DeployKeyGenerated(&domain.RepositoryCredential{}))
}

func DeployKeyGenerated(rc *domain.RepositoryCredential) *domain.Activity {
	return &domain.Activity{
		Name:       "deploy-key.generated",
		OccurredOn: Clock.Now(),
		Extra:      map[string]interface{}{},
		Payload:    rc,
	}
}
