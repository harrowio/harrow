package activities

import (
	"fmt"

	"github.com/harrowio/harrow/domain"
)

func init() {
	registerPayload(UserSignedUp(&domain.User{}))
	registerPayload(UserSignedUpViaGithub(&domain.User{}))
	registerPayload(UserSignedUpViaCapistrano(&domain.User{}))
	registerPayload(UserLoggedIn(&domain.User{}))
	registerPayload(UserLoggedInViaGithub(&domain.User{}))
	registerPayload(UserConnectedGithub(&domain.User{}))
	registerPayload(UserJoinedProject(nil, nil))
	registerPayload(UserLeftProject(nil, nil))
	registerPayload(UserRemovedFromProject(nil, nil))
	registerPayload(UserRequestedPasswordReset(&domain.User{}))
	registerPayload(UserEmailVerified(&domain.User{}))
	registerPayload(UserRequestedVerificationEmail(&domain.User{}))
	registerPayload(UserSignupParameterSet(&domain.User{}, "", ""))
	registerPayload(UserResetPassword(&domain.User{}))
	registerPayload(UserAddedToProject(&domain.User{}, &domain.Project{}))
	registerPayload(UserAccountUsedInTooManyPlaces(&domain.User{}))
	registerPayload(UserReportedAsActive(&domain.User{}))
	registerPayload(UserEnteredSegment(""))
	registerPayload(UserLeftSegment(""))

	for _, segment := range domain.UserSegments {
		registerPayload(UserEnteredSegment(segment))
		registerPayload(UserLeftSegment(segment))
	}
}

func UserSignedUp(user *domain.User) *domain.Activity {
	return &domain.Activity{
		Name:       "user.signed-up",
		OccurredOn: Clock.Now(),
		Extra:      map[string]interface{}{},
		Payload:    user.Scrub(),
	}
}

func UserSignedUpViaGithub(user *domain.User) *domain.Activity {
	return &domain.Activity{
		Name:       "user.signed-up-via-github",
		OccurredOn: Clock.Now(),
		Extra:      map[string]interface{}{},
		Payload:    user.Scrub(),
	}
}

func UserLoggedIn(user *domain.User) *domain.Activity {
	return &domain.Activity{
		Name:       "user.logged-in",
		OccurredOn: Clock.Now(),
		Extra:      map[string]interface{}{},
		Payload:    user.Scrub(),
	}
}

func UserLoggedInViaGithub(user *domain.User) *domain.Activity {
	return &domain.Activity{
		Name:       "user.logged-in-via-github",
		OccurredOn: Clock.Now(),
		Extra:      map[string]interface{}{},
		Payload:    user.Scrub(),
	}
}

func UserConnectedGithub(user *domain.User) *domain.Activity {
	return &domain.Activity{
		Name:       "user.user-connected-github",
		OccurredOn: Clock.Now(),
		Extra:      map[string]interface{}{},
		Payload:    user.Scrub(),
	}
}

type UserProjectPayload struct {
	User    *domain.User
	Project *domain.Project
}

func (self *UserProjectPayload) FindProject(projects domain.ProjectStore) (*domain.Project, error) {
	return self.Project.FindProject(projects)
}

func UserAddedToProject(user *domain.User, project *domain.Project) *domain.Activity {
	payload := &UserProjectPayload{
		User:    user.Scrub(),
		Project: project,
	}

	return &domain.Activity{
		Name:       "user.added-to-project",
		OccurredOn: Clock.Now(),
		Extra:      map[string]interface{}{},
		Payload:    payload,
	}
}

func UserJoinedProject(user *domain.User, project *domain.Project) *domain.Activity {
	payload := &UserProjectPayload{
		User:    user.Scrub(),
		Project: project,
	}

	return &domain.Activity{
		Name:       "user.joined-project",
		OccurredOn: Clock.Now(),
		Extra:      map[string]interface{}{},
		Payload:    payload,
	}
}

func UserLeftProject(user *domain.User, project *domain.Project) *domain.Activity {
	payload := &UserProjectPayload{
		User:    user.Scrub(),
		Project: project,
	}

	return &domain.Activity{
		Name:       "user.left-project",
		OccurredOn: Clock.Now(),
		Extra:      map[string]interface{}{},
		Payload:    payload,
	}
}

func UserRemovedFromProject(user *domain.User, project *domain.Project) *domain.Activity {
	payload := &UserProjectPayload{
		User:    user.Scrub(),
		Project: project,
	}

	return &domain.Activity{
		Name:       "user.removed-from-project",
		OccurredOn: Clock.Now(),
		Extra:      map[string]interface{}{},
		Payload:    payload,
	}
}

func UserRequestedPasswordReset(user *domain.User) *domain.Activity {
	return &domain.Activity{
		Name:       "user.requested-password-reset",
		OccurredOn: Clock.Now(),
		Extra:      map[string]interface{}{},
		Payload:    user.Scrub(),
	}
}

func UserRequestedVerificationEmail(user *domain.User) *domain.Activity {
	return &domain.Activity{
		Name:       "user.requested-verification-email",
		OccurredOn: Clock.Now(),
		Extra:      map[string]interface{}{},
		Payload:    user.Scrub(),
	}
}

func UserEmailVerified(user *domain.User) *domain.Activity {
	return &domain.Activity{
		Name:       "user.email-verified",
		OccurredOn: Clock.Now(),
		Extra:      map[string]interface{}{},
		Payload:    user.Scrub(),
	}
}

func UserSignedUpViaCapistrano(user *domain.User) *domain.Activity {
	return &domain.Activity{
		Name:       "user.signed-up-via-capistrano",
		OccurredOn: Clock.Now(),
		Extra:      map[string]interface{}{},
		Payload:    user.Scrub(),
	}
}

type UserSignupParameterSetPayload struct {
	UserUuid  string
	Parameter string
	Value     interface{}
}

func UserSignupParameterSet(user *domain.User, parameter string, value interface{}) *domain.Activity {
	return &domain.Activity{
		Name:       "user.signup-parameter-set",
		OccurredOn: Clock.Now(),
		Extra:      map[string]interface{}{},
		Payload: &UserSignupParameterSetPayload{
			UserUuid:  user.Uuid,
			Parameter: parameter,
			Value:     value,
		},
	}
}

func UserResetPassword(user *domain.User) *domain.Activity {
	return &domain.Activity{
		Name:       "user.reset-password",
		OccurredOn: Clock.Now(),
		Extra:      map[string]interface{}{},
		Payload:    user.Scrub(),
	}
}

func UserReportedAsActive(user *domain.User) *domain.Activity {
	return &domain.Activity{
		Name:       "user.reported-as-active",
		OccurredOn: Clock.Now(),
		Extra:      map[string]interface{}{},
		Payload:    user.Scrub(),
	}
}

func UserAccountUsedInTooManyPlaces(user *domain.User) *domain.Activity {
	return &domain.Activity{
		Name:       "user.account-used-in-too-many-places",
		OccurredOn: Clock.Now(),
		Extra:      map[string]interface{}{},
		Payload:    user.Scrub(),
	}
}

type SegmentPayload struct {
	Name string `json:"name"`
}

func UserEnteredSegment(segmentName string) *domain.Activity {
	return &domain.Activity{
		Name:       fmt.Sprintf("user.entered-segment-%s", segmentName),
		OccurredOn: Clock.Now(),
		Extra:      map[string]interface{}{},
		Payload: &SegmentPayload{
			Name: segmentName,
		},
	}
}

func UserLeftSegment(segmentName string) *domain.Activity {
	return &domain.Activity{
		Name:       fmt.Sprintf("user.left-segment-%s", segmentName),
		OccurredOn: Clock.Now(),
		Extra:      map[string]interface{}{},
		Payload: &SegmentPayload{
			Name: segmentName,
		},
	}
}
