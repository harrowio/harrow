package harrowArchivist

import (
	"fmt"
	"time"

	"github.com/harrowio/harrow/activities"
	"github.com/harrowio/harrow/domain"
)

type CreationSubject interface {
	Id() string
	CreationDate() time.Time
	AuthorizationName() string
}

type DocumentCreation struct {
	subject    CreationSubject
	activities Activities
}

func NewDocumentCreation(subject CreationSubject, activities Activities) *DocumentCreation {
	return &DocumentCreation{
		subject:    subject,
		activities: activities,
	}
}

// Activity returns the activity to create for the organization.
func (self *DocumentCreation) Activity(prototype *domain.Activity) (*domain.Activity, error) {
	var (
		activity *domain.Activity
		err      error
	)

	if existing, err := self.activities.FindActivityByNameAndPayloadUuid(prototype.Name, self.subject.Id()); err != nil && !domain.IsNotFound(err) {
		return nil, err
	} else if existing != nil {
		return nil, nil
	}

	if prototype.Name == "organization.created" {
		activity, err = self.organizationCreated()
	} else {
		activity, err = &domain.Activity{
			Name:    prototype.Name,
			Payload: self.subject,
			Extra:   map[string]interface{}{},
		}, nil
	}

	if err != nil {
		return nil, err
	}
	activity.OccurredOn = self.subject.CreationDate()
	return activity, nil
}

func (self *DocumentCreation) organizationCreated() (*domain.Activity, error) {
	organization, ok := self.subject.(*domain.Organization)
	if !ok {
		return nil, fmt.Errorf("expected subject to be %T, but is %T", organization, self.subject)
	}

	return activities.OrganizationCreated(organization, domain.FreePlan), nil
}
