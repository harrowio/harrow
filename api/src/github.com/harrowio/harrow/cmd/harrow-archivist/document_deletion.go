package harrowArchivist

import (
	"fmt"
	"time"

	"github.com/harrowio/harrow/activities"
	"github.com/harrowio/harrow/domain"
)

type DeletionSubject interface {
	Id() string
	DeletionDate() time.Time
	AuthorizationName() string
}

type DocumentDeletion struct {
	subject    DeletionSubject
	activities Activities
}

func NewDocumentDeletion(subject DeletionSubject, activities Activities) *DocumentDeletion {
	return &DocumentDeletion{
		subject:    subject,
		activities: activities,
	}
}

// Activity returns the activity to create for the organization.
func (self *DocumentDeletion) Activity(prototype *domain.Activity) (*domain.Activity, error) {
	var (
		activity *domain.Activity
		err      error
	)

	if existing, err := self.activities.FindActivityByNameAndPayloadUuid(prototype.Name, self.subject.Id()); err != nil && !domain.IsNotFound(err) {
		return nil, err
	} else if existing != nil {
		return nil, nil
	}

	if prototype.Name == "organization.archived" {
		activity, err = self.organizationArchived()
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
	activity.OccurredOn = self.subject.DeletionDate()
	return activity, nil
}

func (self *DocumentDeletion) organizationArchived() (*domain.Activity, error) {
	organization, ok := self.subject.(*domain.Organization)
	if !ok {
		return nil, fmt.Errorf("expected subject to be %T, but is %T", organization, self.subject)
	}

	return activities.OrganizationArchived(organization), nil
}
