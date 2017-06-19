package domain

import (
	"fmt"
	"time"

	"github.com/harrowio/harrow/uuidhelper"
)

type Project struct {
	defaultSubject
	Uuid             string     `json:"uuid"`
	OrganizationUuid string     `json:"organizationUuid" db:"organization_uuid"`
	Name             string     `json:"name"`
	Public           bool       `json:"public"`
	CreatedAt        time.Time  `json:"createdAt"        db:"created_at"`
	ArchivedAt       *time.Time `json:"archivedAt"       db:"archived_at"`
}

func ValidateProject(u *Project) error {

	if len(u.Name) == 0 {
		return NewValidationError("name", "required")
	}

	if len(u.OrganizationUuid) == 0 {
		return NewValidationError("organizationUuid", "required")
	}

	return nil
}

func (self *Project) OwnUrl(requestScheme, requestBaseUri string) string {
	return fmt.Sprintf("%s://%s/projects/%s", requestScheme, requestBaseUri, self.Uuid)
}

func (self *Project) Links(response map[string]map[string]string, requestScheme, requestBaseUri string) map[string]map[string]string {
	response["self"] = map[string]string{"href": self.OwnUrl(requestScheme, requestBaseUri)}
	response["environments"] = map[string]string{"href": fmt.Sprintf("%s://%s/projects/%s/environments", requestScheme, requestBaseUri, self.Uuid)}
	response["invitations"] = map[string]string{"href": fmt.Sprintf("%s://%s/invitations", requestScheme, requestBaseUri)}
	response["jobs"] = map[string]string{"href": fmt.Sprintf("%s://%s/projects/%s/jobs", requestScheme, requestBaseUri, self.Uuid)}
	response["schedules"] = map[string]string{"href": fmt.Sprintf("%s://%s/projects/%s/schedules", requestScheme, requestBaseUri, self.Uuid)}
	response["scheduled-executions"] = map[string]string{"href": fmt.Sprintf("%s://%s/projects/%s/scheduled-executions", requestScheme, requestBaseUri, self.Uuid)}
	response["operations"] = map[string]string{"href": fmt.Sprintf("%s://%s/projects/%s/operations", requestScheme, requestBaseUri, self.Uuid)}
	response["organization"] = map[string]string{"href": fmt.Sprintf("%s://%s/organizations/%s", requestScheme, requestBaseUri, self.OrganizationUuid)}
	response["repositories"] = map[string]string{"href": fmt.Sprintf("%s://%s/projects/%s/repositories", requestScheme, requestBaseUri, self.Uuid)}
	response["project-members"] = map[string]string{"href": fmt.Sprintf("%s://%s/projects/%s/members", requestScheme, requestBaseUri, self.Uuid)}
	response["leave"] = map[string]string{"href": fmt.Sprintf("%s://%s/projects/%s/members", requestScheme, requestBaseUri, self.Uuid)}
	response["webhooks"] = map[string]string{"href": fmt.Sprintf("%s://%s/projects/%s/webhooks", requestScheme, requestBaseUri, self.Uuid)}
	response["project-card"] = map[string]string{"href": fmt.Sprintf("%s://%s/projects/%s/card", requestScheme, requestBaseUri, self.Uuid)}
	response["git-triggers"] = map[string]string{"href": fmt.Sprintf("%s://%s/projects/%s/git-triggers", requestScheme, requestBaseUri, self.Uuid)}
	response["slack-notifiers"] = map[string]string{"href": fmt.Sprintf("%s://%s/projects/%s/slack-notifiers", requestScheme, requestBaseUri, self.Uuid)}
	response["email-notifiers"] = map[string]string{"href": fmt.Sprintf("%s://%s/projects/%s/email-notifiers", requestScheme, requestBaseUri, self.Uuid)}
	response["job-notifiers"] = map[string]string{"href": fmt.Sprintf("%s://%s/projects/%s/job-notifiers", requestScheme, requestBaseUri, self.Uuid)}
	response["tasks"] = map[string]string{"href": fmt.Sprintf("%s://%s/projects/%s/tasks", requestScheme, requestBaseUri, self.Uuid)}
	response["scripts"] = map[string]string{"href": fmt.Sprintf("%s://%s/projects/%s/scripts", requestScheme, requestBaseUri, self.Uuid)}
	response["notification-rules"] = map[string]string{"href": fmt.Sprintf("%s://%s/projects/%s/notification-rules", requestScheme, requestBaseUri, self.Uuid)}

	return response
}

// FindOrganization satisfies authz.BelongsToOrganization in order to determine
// authorization.
func (self *Project) FindOrganization(store OrganizationStore) (*Organization, error) {
	return store.FindByUuid(self.OrganizationUuid)
}

// FindProject satisfies authz.BelongsToProject in order to determine
// authorization.
func (self *Project) FindProject(store ProjectStore) (*Project, error) {
	if self.Uuid == "" {
		return nil, nil
	}

	return self, nil
}

// NewInvitationForUser returns an invitation object suitable for inviting the
// given Harrow user to the given project.  Membership indicates the
// desired project membership level, i.e. whether the invited user should be
// considered a "guest" or "member", etc of the project, should she accept.
// Valid values for membership are defined by the various MembershipType*
// constants.
func (self *Project) NewInvitationToUser(message string, membership string, recipient *User) *Invitation {
	return &Invitation{
		InviteeUuid:      recipient.Uuid,
		ProjectUuid:      self.Uuid,
		OrganizationUuid: self.OrganizationUuid,
		Message:          message,
		RecipientName:    recipient.Name,
		Email:            recipient.Email,
		MembershipType:   membership,
	}
}

// NewInvitationToHarrow returns an invitation object suitable for
// inviting someone to join Harrow and become part of this project.
// See NewInvitationForUser for the meaning of the membership parameter.
func (self *Project) NewInvitationToHarrow(name, email, message string, membership string) *Invitation {
	return &Invitation{
		InviteeUuid:      uuidhelper.MustNewV4(),
		ProjectUuid:      self.Uuid,
		OrganizationUuid: self.OrganizationUuid,
		Message:          message,
		RecipientName:    name,
		Email:            email,
		MembershipType:   membership,
	}
}

// NewEnvironment constructs a new environment object for this project.
func (self *Project) NewEnvironment(name string) *Environment {
	return &Environment{
		Name:        name,
		ProjectUuid: self.Uuid,
		Variables: EnvironmentVariables{
			M: map[string]string{},
		},
	}
}

func (self *Project) NewDefaultEnvironment() *Environment {
	env := self.NewEnvironment("Default")
	env.IsDefault = true
	return env
}

// NewTask constructs a new task object for this project.
func (self *Project) NewTask(name, body string) *Task {
	return &Task{
		Name:        name,
		Body:        body,
		ProjectUuid: self.Uuid,
		Type:        "script",
	}
}

// NewMembership constructs a new project membership for this project
// and the given user.
func (self *Project) NewMembership(member *User, membershipType string) *ProjectMembership {
	return &ProjectMembership{
		ProjectUuid:    self.Uuid,
		UserUuid:       member.Uuid,
		MembershipType: membershipType,
	}
}

// NewRepository constructs a new repository for this project.
func (self *Project) NewRepository(name, url string) *Repository {
	return &Repository{
		ProjectUuid: self.Uuid,
		Name:        name,
		Url:         url,
	}
}

// NewJob constructs a new job for the given task and environment in
// this project.
func (self *Project) NewJob(name, taskUuid, environmentUuid string) *Job {
	return &Job{
		Name:            name,
		ProjectUuid:     self.Uuid,
		TaskUuid:        taskUuid,
		EnvironmentUuid: environmentUuid,
	}
}

// NewEmailNotifier returns a new email notifier for the given
// recipient address and url host in this project.
func (self *Project) NewEmailNotifier(recipient, urlHost string) *EmailNotifier {
	return &EmailNotifier{
		ProjectUuid: &self.Uuid,
		Recipient:   recipient,
		UrlHost:     urlHost,
	}
}

// NewGitTrigger returns a new git trigger in this project which fires
// on changes to master.
func (self *Project) NewGitTrigger(name, creatorUuid, jobUuid string) *GitTrigger {
	trigger := NewGitTrigger(name, creatorUuid)
	trigger.JobUuid = jobUuid
	trigger.ProjectUuid = self.Uuid
	trigger.ChangeType = "change"
	trigger.MatchRef = ".*"
	return trigger
}

// NewNotificationRule ...
func (self *Project) NewNotificationRule(notifierType, notifierUuid, jobUuid, matchActivities string) *NotificationRule {
	return &NotificationRule{
		ProjectUuid:   self.Uuid,
		NotifierType:  notifierType,
		NotifierUuid:  notifierUuid,
		JobUuid:       &jobUuid,
		MatchActivity: matchActivities,
	}
}

func (self *Project) AuthorizationName() string {
	if self.Public {
		return "public"
	} else {
		return "project"
	}
}

func (self *Project) CreationDate() time.Time {
	return self.CreatedAt
}

func (self *Project) DeletionDate() time.Time {
	if self.ArchivedAt == nil {
		return time.Time{}
	}

	return *self.ArchivedAt
}

func (self *Project) Id() string {
	return self.Uuid
}
