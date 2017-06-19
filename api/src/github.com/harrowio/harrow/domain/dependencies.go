package domain

// ProjectStore defines all operations that are necessary for the domain
// objects to work properly.
type ProjectStore interface {
	FindByUuid(uuid string) (*Project, error)
	FindByMemberUuid(uuid string) (*Project, error)
	FindByOrganizationUuid(uuid string) (*Project, error)
	FindByNotifierUuid(uuid, notifierType string) (*Project, error)
	FindByJobUuid(uuid string) (*Project, error)
	FindByTaskUuid(uuid string) (*Project, error)
	FindByRepositoryUuid(uuid string) (*Project, error)
	FindByEnvironmentUuid(uuid string) (*Project, error)
	FindByWebhookUuid(uuid string) (*Project, error)
	FindByNotificationRule(notifierType string, notifierUuid string) (*Project, error)
}

type JobStore interface {
	FindByUuid(uuid string) (*Job, error)
}

// UserStore defines all the operations that are necessary for the domain
// to fetch associated user objects.
type UserStore interface {
	FindByUuid(uuid string) (*User, error)
	FindAllSubscribers(watchableId, event string) ([]*User, error)
}

// OrganizationStore defines all the operations that are necessary for
// the domain to fetch associated organization objects.
type OrganizationStore interface {
	FindByUuid(uuid string) (*Organization, error)
	FindByProjectUuid(uuid string) (*Organization, error)
}

// EnvironmentStore defines all the operations that are necessary for domain
// objects to fetch associated environment objects.
type EnvironmentStore interface {
	FindByJobUuid(uuid string) (*Environment, error)
}

// SecretStore defines all the operations that are necessary for domain
// objects to fetch associated secret objects
type SecretStore interface {
	FindAllByEnvironmentUuid(environmentUuid string) ([]*Secret, error)
}

// WorkspaceBaseImageStore defines all the operations that are necessary
// for domain objects to fetch associated workspace base images.
type WorkspaceBaseImageStore interface {
	FindByUuid(uuid string) (*WorkspaceBaseImage, error)
}

// RepositoryStore defines all the operations that are necessary for
// domain objects to fetch associated repository objects.
type RepositoryStore interface {
	FindByUuid(repositoryUuid string) (*Repository, error)
	FindAllByJobUuid(jobUuid string) ([]*Repository, error)
	MarkAsAccessible(repositoryUuid string, accessible bool) error
}

// RepositoryCredentialStore defines all the operations that are necessary for
// domain objects to fetch associated repository credential objects.
type RepositoryCredentialStore interface {
	FindByRepositoryUuid(repositoryUuid string) (*RepositoryCredential, error)
	FindByRepositoryUuidAndType(repositoryUuid string, credentialType RepositoryCredentialType) (*RepositoryCredential, error)
}

// TaskStore defines all the operations that are necessary for domain
// objects to fetch associated task objects.
type TaskStore interface {
	FindByJobUuid(jobUuid string) (*Task, error)
}

type OperationStore interface {
	MarkExitStatus(operationUuid string, exitStatus int) error
	FindPreviousOperation(currentOperationUuid string) (*Operation, error)
}

type SubscriptionStore interface {
	Create(subscription *Subscription) (string, error)
	Find(watchableId, event, userUuid string) (*Subscription, error)
	FindEventsForUser(watchableId, userUuid string) ([]string, error)
	Delete(subscriptionUuid string) error
}

type InvitationStore interface {
	FindByUserAndProjectUuid(userId, projectId string) (*Invitation, error)
}

// TotpToken defines the methods necessary for verifying time-based
// one-time password tokens.
type TotpToken interface {
	FromNow(period int64) int32
	Now() int32
}

type Archiver interface {
	ArchiveByUuid(uuid string) error
}

type RecentOperations interface {
	FindRecentByJobUuid(n int, jobUuid string) ([]*Operation, error)
}

type EventPayload interface {
	Get(key string) string
}

type RepositoriesByName interface {
	FindAllByProjectUuidAndRepositoryName(projectUuid, name string) ([]*Repository, error)
}

type DeliveryStore interface {
	FindByUuid(uuid string) (*Delivery, error)
}
