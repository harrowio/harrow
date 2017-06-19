package http

import (
	"reflect"
	"strings"
	"testing"

	"github.com/harrowio/harrow/config"
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/test_helpers"
)

type ArrayActivitySink struct {
	Emitted []*domain.Activity
}

func NewArrayActivitySink() *ArrayActivitySink {
	return &ArrayActivitySink{
		Emitted: []*domain.Activity{},
	}
}

func (self *ArrayActivitySink) EmitActivity(activity *domain.Activity) {
	self.Emitted = append(self.Emitted, activity)
}

func Test_StandardContext_newTx_setsContextUser(t *testing.T) {
	db := test_helpers.GetDbConnection(t)
	tx := db.MustBegin()
	defer tx.Rollback()
	ctxt := NewStandardContextTx(db, tx, config.GetConfig(), test_helpers.NewMockKeyValueStore(), test_helpers.NewMockSecretKeyValueStore())
	world := test_helpers.MustNewWorld(tx, t)
	user := world.User("default")
	session := setupTestLoginSession(t, tx, user)

	req, err := newAuthenticatedRequest(session.Uuid, "GET", "/does-not-exist", "")
	if err != nil {
		t.Fatal(err)
	}

	reqCtxt, err := ctxt.RequestContext(nil, req)
	if err != nil {
		t.Fatal(err)
	}
	contextUser := struct {
		Uuid string `db:"uuid"`
	}{}

	if u := reqCtxt.User(); user.Uuid != u.Uuid {
		t.Fatalf("Wrong uuid for reqCtxt.User: %q, expected: %q", u.Uuid, user.Uuid)
	}

	if err := reqCtxt.Tx().Get(&contextUser, `SELECT current_setting('harrow.context_user_uuid') as uuid`); err != nil {
		t.Fatal(err)
	}

	if contextUser.Uuid != user.Uuid {
		t.Fatalf("Wrong context user uuid: %q, expected: %q", contextUser.Uuid, user.Uuid)
	}
}

func Test_StandardContext_newTx_does_not_setContextUser_ifUserIsLoggedOut(t *testing.T) {
	db := test_helpers.GetDbConnection(t)
	tx := db.MustBegin()
	defer tx.Rollback()
	ctxt := NewStandardContextTx(db, tx, config.GetConfig(), test_helpers.NewMockKeyValueStore(), test_helpers.NewMockSecretKeyValueStore())

	req, err := newRequest("GET", "/does-not-exist", "")
	if err != nil {
		t.Fatal(err)
	}

	reqCtxt, err := ctxt.RequestContext(nil, req)
	if err != nil {
		t.Fatal(err)
	}
	contextUser := struct {
		Uuid string `db:"uuid"`
	}{}

	if u := reqCtxt.User(); u != (nil) {
		t.Fatalf("Did not expect a user on the request, got: %#v", u)
	}

	if err := reqCtxt.Tx().Get(&contextUser, `SELECT current_setting('harrow.context_user_uuid') as uuid`); err == nil {
		if contextUser.Uuid != "" {
			t.Fatalf("Expected contextUser.Uuid to be empty, got: %q", contextUser.Uuid)
		}
	} else {
		msg := err.Error()
		if !strings.Contains(msg, "harrow.context_user_uuid") {
			t.Fatalf("Unexpected error: %s", msg)
		}
	}
}

func Test_StandardContext_RequestContext_returnsErrSesssionExpired_ifSessionIsExpired(t *testing.T) {
	db := test_helpers.GetDbConnection(t)
	tx := db.MustBegin()
	defer tx.Rollback()
	ctxt := NewStandardContextTx(db, tx, config.GetConfig(), test_helpers.NewMockKeyValueStore(), test_helpers.NewMockSecretKeyValueStore())
	world := test_helpers.MustNewWorld(tx, t)
	user := world.User("default")
	session := setupTestLoginSession(t, tx, user)

	tx.MustExec(`UPDATE sessions SET expires_at = '2010-01-01'::timestamp WHERE uuid = $1`, session.Uuid)

	req, err := newAuthenticatedRequest(session.Uuid, "GET", "/does-not-exist", "")
	if err != nil {
		t.Fatal(err)
	}

	_, err = ctxt.RequestContext(nil, req)
	if err == nil {
		t.Fatal("expected an error")
	}

	if got, want := err, ErrSessionExpired; !reflect.DeepEqual(got, want) {
		t.Fatalf("err = %#v; want %#v", got, want)
	}

}

func Test_StandardContext_RequestContext_EnqueueActivity_recordsCurrentUserUuidOnActivity(t *testing.T) {
	db := test_helpers.GetDbConnection(t)
	tx := db.MustBegin()
	defer tx.Rollback()
	ctxt := NewStandardContextTx(db, tx, config.GetConfig(), test_helpers.NewMockKeyValueStore(), test_helpers.NewMockSecretKeyValueStore())
	world := test_helpers.MustNewWorld(tx, t)
	user := world.User("default")
	session := setupTestLoginSession(t, tx, user)

	req, err := newAuthenticatedRequest(session.Uuid, "GET", "/does-not-exist", "")
	if err != nil {
		t.Fatal(err)
	}
	reqCtxt, err := ctxt.RequestContext(nil, req)
	if err != nil {
		t.Fatal(err)
	}

	activity := domain.NewActivity(1, "test.run")
	reqCtxt.EnqueueActivity(activity, nil)
	got := ""
	if activity.ContextUserUuid != nil {
		got = *activity.ContextUserUuid
	}
	if want := reqCtxt.User().Uuid; got != want {
		t.Errorf("*activity.ContextUserUuid = %s; want %s", got, want)
	}
}

func Test_StandardContext_RequestContext_EnqueueActivity_recordsProvidedUserUuidOnActivity(t *testing.T) {
	db := test_helpers.GetDbConnection(t)
	tx := db.MustBegin()
	defer tx.Rollback()
	ctxt := NewStandardContextTx(db, tx, config.GetConfig(), test_helpers.NewMockKeyValueStore(), test_helpers.NewMockSecretKeyValueStore())
	world := test_helpers.MustNewWorld(tx, t)
	user := world.User("default")
	session := setupTestLoginSession(t, tx, user)

	req, err := newAuthenticatedRequest(session.Uuid, "GET", "/does-not-exist", "")
	if err != nil {
		t.Fatal(err)
	}
	reqCtxt, err := ctxt.RequestContext(nil, req)
	if err != nil {
		t.Fatal(err)
	}

	activity := domain.NewActivity(1, "test.run")
	reqCtxt.EnqueueActivity(activity, &user.Uuid)
	got := ""
	if activity.ContextUserUuid != nil {
		got = *activity.ContextUserUuid
	}
	if want := reqCtxt.User().Uuid; got != want {
		t.Errorf("*activity.ContextUserUuid = %s; want %s", got, want)
	}
}
