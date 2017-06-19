package test_helpers

import (
	"net/http"
	"os"

	"github.com/harrowio/harrow/config"
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/stores"

	"testing"

	"github.com/jmoiron/sqlx"
)

var db *sqlx.DB

func GetConfig(t *testing.T) *config.Config {
	return config.GetConfig()
}

func GetDbConnection(t *testing.T) *sqlx.DB {
	var err error
	if db == nil {
		c := GetConfig(t)
		if db, err = c.DB(); err != nil {
			t.Fatal("Error Opening Database Handle:", err)
		}
	}
	return db
}

func GetDbTx(t *testing.T) *sqlx.Tx {
	tx, err := GetDbConnection(t).Beginx()
	if err != nil {
		t.Fatal(err)
	}
	return tx
}

func MustCreateOrganization(t *testing.T, tx *sqlx.Tx, o *domain.Organization) *domain.Organization {
	var err error
	var store *stores.DbOrganizationStore = stores.NewDbOrganizationStore(tx)
	if o.Uuid, err = store.Create(o); err != nil {
		t.Fatal(err)
	}
	o, err = store.FindByUuid(o.Uuid)
	if err != nil {
		t.Fatal(err)
	}
	return o
}

func MustCreateOrganizationMembership(t *testing.T, tx *sqlx.Tx, om *domain.OrganizationMembership) *domain.OrganizationMembership {
	var err error
	var store *stores.DbOrganizationMembershipStore = stores.NewDbOrganizationMembershipStore(tx)
	if err = store.Create(om); err != nil {
		t.Fatal(err)
	}
	om, err = store.FindByOrganizationAndUserUuids(om.OrganizationUuid, om.UserUuid)
	if err != nil {
		t.Fatal(err)
	}
	return om
}

func MustCreateEnvironment(t *testing.T, tx *sqlx.Tx, env *domain.Environment) *domain.Environment {
	store := stores.NewDbEnvironmentStore(tx)
	uuid, err := store.Create(env)
	if err != nil {
		t.Fatalf("MustCreateEnvironment: store.Create: %s\n", err)
	}

	retrievedEnvironment, err := store.FindByUuid(uuid)
	if err != nil {
		t.Fatalf("MustCreateEnvironment: store.FindByUuid: %s\n", err)
	}

	return retrievedEnvironment
}

func MustCreateSecret(t *testing.T, tx *sqlx.Tx, ss stores.SecretKeyValueStore, secret *domain.Secret) *domain.Secret {
	store := stores.NewSecretStore(ss, tx)
	uuid, err := store.Create(secret)
	if err != nil {
		t.Fatalf("MustCreateSecret: store.Create: %s\n", err)
	}

	retrievedSecret, err := store.FindByUuid(uuid)
	if err != nil {
		t.Fatalf("MustCreateSecret: store.FindByUuid: %s\n", err)
	}

	return retrievedSecret
}

func MustCreateJob(t *testing.T, tx *sqlx.Tx, job *domain.Job) *domain.Job {
	store := stores.NewDbJobStore(tx)
	uuid, err := store.Create(job)

	if err != nil {
		t.Fatalf("MustCreateJob: store.Create: %s\n", err)
	}

	retrievedJob, err := store.FindByUuid(uuid)
	if err != nil {
		t.Fatalf("MustCreateJob: store.FindByUuid: %s\n", err)
	}

	return retrievedJob
}

func MustCreateSchedule(t *testing.T, tx *sqlx.Tx, schedule *domain.Schedule) *domain.Schedule {
	store := stores.NewDbScheduleStore(tx)
	uuid, err := store.Create(schedule)
	if err != nil {
		t.Fatalf("MustCreateSchedule: store.Create: %s\n", err)
	}

	retrievedSchedule, err := store.FindByUuid(uuid)
	if err != nil {
		t.Fatalf("MustCreateSchedule: store.FindByUuid: %s\n", err)
	}

	return retrievedSchedule
}

func MustCreateTask(t *testing.T, tx *sqlx.Tx, task *domain.Task) *domain.Task {
	store := stores.NewDbTaskStore(tx)
	uuid, err := store.Create(task)
	if err != nil {
		t.Fatalf("MustCreateTask: store.Create: %s\n", err)
	}

	retrievedTask, err := store.FindByUuid(uuid)
	if err != nil {
		t.Fatalf("MustCreateTask: store.FindByUuid: %s\n", err)
	}

	return retrievedTask
}

func MustCreateProject(t *testing.T, tx *sqlx.Tx, p *domain.Project) *domain.Project {
	var err error
	var store *stores.DbProjectStore = stores.NewDbProjectStore(tx)
	if p.Uuid, err = store.Create(p); err != nil {
		t.Fatal(err)
	}
	p, err = store.FindByUuid(p.Uuid)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func MustCreateRepository(t *testing.T, tx *sqlx.Tx, r *domain.Repository) *domain.Repository {
	var err error
	var store *stores.DbRepositoryStore = stores.NewDbRepositoryStore(tx)
	if r.Uuid, err = store.Create(r); err != nil {
		t.Fatal(err)
	}
	r, err = store.FindByUuid(r.Uuid)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func MustCreateOperation(t *testing.T, tx *sqlx.Tx, o *domain.Operation) *domain.Operation {
	var err error
	var store *stores.DbOperationStore = stores.NewDbOperationStore(tx)
	if o.Uuid, err = store.Create(o); err != nil {
		t.Fatal(err)
	}
	o, err = store.FindByUuid(o.Uuid)
	if err != nil {
		t.Fatal(err)
	}
	return o
}

func MustCreateOAuthToken(t *testing.T, tx *sqlx.Tx, token *domain.OAuthToken) *domain.OAuthToken {
	var err error
	var store *stores.DbOAuthTokenStore = stores.NewDbOAuthTokenStore(tx)
	if token.Uuid, err = store.Create(token); err != nil {
		t.Fatal(err)
	}
	token, err = store.FindByUuid(token.Uuid)
	if err != nil {
		t.Fatal(err)
	}
	return token
}

func MustCreateUser(t *testing.T, tx *sqlx.Tx, u *domain.User) *domain.User {
	var err error
	var store *stores.DbUserStore = stores.NewDbUserStore(tx, GetConfig(t))

	if u.Uuid, err = store.Create(u); err != nil {
		t.Fatal(err)
	}
	u, err = store.FindByUuid(u.Uuid)
	if err != nil {
		t.Fatal(err)
	}
	return u
}

func MustCreateSession(t *testing.T, tx *sqlx.Tx, s *domain.Session) *domain.Session {
	var err error
	var store *stores.DbSessionStore = stores.NewDbSessionStore(tx)
	if s.Uuid, err = store.Create(s); err != nil {
		t.Fatal(err)
	}
	s, err = store.FindByUuid(s.Uuid)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func MustCreateInvitation(t *testing.T, tx *sqlx.Tx, inv *domain.Invitation) *domain.Invitation {
	var (
		err  error
		uuid string
	)
	store := stores.NewDbInvitationStore(tx)
	if uuid, err = store.Create(inv); err != nil {
		t.Fatal(err)
	}

	invitation, err := store.FindByUuid(uuid)
	if err != nil {
		t.Fatal(err)
	}
	return invitation
}

func MustCreateProjectMembership(t *testing.T, tx *sqlx.Tx, membership *domain.ProjectMembership) *domain.ProjectMembership {
	var (
		err  error
		uuid string
	)
	store := stores.NewDbProjectMembershipStore(tx)
	if uuid, err = store.Create(membership); err != nil {
		t.Fatal(err)
	}

	membership, err = store.FindByUuid(uuid)
	if err != nil {
		t.Fatal(err)
	}
	return membership
}

func MustCreateWorkspaceBaseImage(t *testing.T, tx *sqlx.Tx, wsbi *domain.WorkspaceBaseImage) *domain.WorkspaceBaseImage {
	store := stores.NewDbWorkspaceBaseImageStore(tx)
	uuid, err := store.Create(wsbi)
	if err != nil {
		t.Fatal(err)
	}

	wsbi, err = store.FindByUuid(uuid)
	if err != nil {
		t.Fatal(err)
	}

	return wsbi
}

func MustCreateRepositoryCredential(t *testing.T, tx *sqlx.Tx, ss stores.SecretKeyValueStore, rc *domain.RepositoryCredential) *domain.RepositoryCredential {
	store := stores.NewRepositoryCredentialStore(ss, tx)
	uuid, err := store.Create(rc)
	if err != nil {
		t.Fatal(err)
	}

	rc, err = store.FindByUuid(uuid)
	if err != nil {
		t.Fatal(err)
	}

	return rc
}

func MustSelectBillingPlan(t *testing.T, tx *sqlx.Tx, organization *domain.Organization, plan *domain.BillingPlan) {
	eventData := &domain.BillingPlanSelected{}
	eventData.FillFromPlan(plan)
	eventData.SubscriptionId = "subscription-" + organization.Uuid
	event := organization.NewBillingEvent(eventData)
	_, err := stores.NewDbBillingEventStore(tx).Create(event)
	if err != nil {
		t.Fatal(err)
	}
}

func MustCreateGitTrigger(t *testing.T, tx *sqlx.Tx, trigger *domain.GitTrigger) *domain.GitTrigger {
	store := stores.NewDbGitTriggerStore(tx)
	uuid, err := store.Create(trigger)
	if err != nil {
		t.Fatal(err)
	}

	gitTrigger, err := store.FindByUuid(uuid)
	if err != nil {
		t.Fatal(err)
	}

	return gitTrigger
}

func InStrings(needle string, haystack []string) bool {
	for _, str := range haystack {
		if str == needle {
			return true
		}
	}

	return false
}

func IsBraintreeProxyAvailable() bool {
	res, err := http.Get("http://localhost:10002/")
	if res != nil {
		defer res.Body.Close()
	}

	if err != nil || res.StatusCode >= 400 {
		return false
	}

	return true
}

func Flaky(t *testing.T) {
	if os.Getenv("HARROW_SKIP_FLAKY_TESTS") != "" {
		t.Skip("flaky test skipped")
	}
}
