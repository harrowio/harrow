package activities

import "github.com/harrowio/harrow/domain"

func init() {
	registerPayload(InvitationCreated(&domain.Invitation{}))
	registerPayload(InvitationAccepted(&domain.Invitation{}))
	registerPayload(InvitationRefused(&domain.Invitation{}))
}

func InvitationCreated(invitation *domain.Invitation) *domain.Activity {
	return &domain.Activity{
		Name:       "invitation.created",
		OccurredOn: Clock.Now(),
		Payload:    invitation,
		Extra:      map[string]interface{}{},
	}
}

func InvitationAccepted(invitation *domain.Invitation) *domain.Activity {
	return &domain.Activity{
		Name:       "invitation.accepted",
		OccurredOn: Clock.Now(),
		Payload:    invitation,
		Extra:      map[string]interface{}{},
	}
}

func InvitationRefused(invitation *domain.Invitation) *domain.Activity {
	return &domain.Activity{
		Name:       "invitation.refused",
		OccurredOn: Clock.Now(),
		Payload:    invitation,
		Extra:      map[string]interface{}{},
	}
}
