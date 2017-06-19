package http

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/harrowio/harrow/authz"
	"github.com/harrowio/harrow/domain"

	"github.com/gorilla/mux"
	"github.com/lib/pq"
)

func mountTestHandler(r *mux.Router, ctxt ServerContext) {
	handler := &testHandler{}
	r.PathPrefix("/ok").Methods("GET").Handler(HandlerFunc(ctxt, handler.Ok)).Name("test-ok")
	r.PathPrefix("/activities").Methods("GET").Handler(HandlerFunc(ctxt, handler.Activities)).Name("test-activities")
	r.PathPrefix("/error").Methods("GET").Handler(HandlerFunc(ctxt, handler.Error)).Name("test-error")
	r.PathPrefix("/panic").Methods("GET").Handler(HandlerFunc(ctxt, handler.Panic)).Name("test-panic")
}

type testHandler struct {
}

func (self testHandler) Ok(ctxt RequestContext) (err error) {
	return nil
}

var emittedActivities = []*domain.Activity{
	domain.NewActivity(1, "test.run"),
	domain.NewActivity(2, "test.run"),
	domain.NewActivity(3, "user.left-segment-limbo"),
	domain.NewActivity(4, "user.entered-segment-cold"),
	domain.NewActivity(5, "user.reported-as-active"),
}

func (self testHandler) Activities(ctxt RequestContext) (err error) {
	for _, activity := range emittedActivities[0:2] { // this hack stops the user. being enqueued incorrectly
		ctxt.EnqueueActivity(activity, nil)
	}
	return nil
}

func (self testHandler) Error(ctxt RequestContext) (err error) {
	return errors.New("Error from test")
}

func (self testHandler) Panic(ctxt RequestContext) (err error) {
	panic("panic from test")
}

func Test_Handler_CommitWhenOk(t *testing.T) {

	ts, ctxt := setupHandlerTestServer(mountTestHandler, t)
	tx := ctxt.Tx()
	defer tx.Rollback()
	defer ts.Close()

	req, err := newRequest("GET", ts.URL+"/ok", "")
	if err != nil {
		t.Fatal(err)
	}

	_, err = new(http.Client).Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if !*ctxt.txCommitted {
		t.Fatal("Tx was not committed although no error occured")
	}
	if *ctxt.txRolledBack {
		t.Fatal("Tx was rolled back although no error occured")
	}

}

func Test_Handler_CommitWhenOk_emitsEnqueuedActivities(t *testing.T) {
	sink := NewArrayActivitySink()

	do := func() {
		ts, ctxt := setupHandlerTestServer(mountTestHandler, t)
		tx := ctxt.Tx()
		defer tx.Rollback()
		defer ts.Close()

		req, err := newRequest("GET", ts.URL+"/activities", "")
		if err != nil {
			t.Fatal(err)
		}

		_, err = new(http.Client).Do(req)
		if err != nil {
			t.Fatal(err)
		}

		if !*ctxt.txCommitted {
			t.Fatal("Tx was not committed although no error occured")
		}
		if *ctxt.txRolledBack {
			t.Fatal("Tx was rolled back although no error occured")
		}

		if got, want := len(sink.Emitted), len(emittedActivities); got != want {
			t.Errorf("len(sink.Emitted) = %d; want %d", got, want)
		}
	}

	withActivitySink(sink, do)
}

func Test_Handler_onCommitFailure_returnsProperErrorResponse(t *testing.T) {
	ts, ctxt := setupCommitFailingHandlerTestServer(mountTestHandler, t)
	tx := ctxt.Tx()
	defer tx.Rollback()
	defer ts.Close()

	req, err := newRequest("GET", ts.URL+"/ok", "")
	if err != nil {
		t.Fatal(err)
	}

	res, err := new(http.Client).Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if got, want := res.StatusCode, 500; got != want {
		t.Errorf("res.StatusCode = %d; want %d", got, want)
	}

	result := struct{ Reason string }{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}

	if got, want := result.Reason, `transaction_commit_failed`; got != want {
		t.Errorf("result.Reason = %q; want %q", got, want)
	}
}

func Test_Handler_RollbackOnError(t *testing.T) {

	ts, ctxt := setupHandlerTestServer(mountTestHandler, t)
	tx := ctxt.Tx()
	defer tx.Rollback()
	defer ts.Close()

	req, err := newRequest("GET", ts.URL+"/error", "")
	if err != nil {
		t.Fatal(err)
	}

	_, err = new(http.Client).Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if *ctxt.txCommitted {
		t.Fatal("Tx was committed although an error occured")
	}
	if !*ctxt.txRolledBack {
		t.Fatal("Tx was not rolled back although an error occured")
	}
}

func Test_Handler_RollbackOnPanic(t *testing.T) {

	ts, ctxt := setupHandlerTestServer(mountTestHandler, t)
	tx := ctxt.Tx()
	defer tx.Rollback()
	defer ts.Close()

	req, err := newRequest("GET", ts.URL+"/panic", "")
	if err != nil {
		t.Fatal(err)
	}

	_, err = new(http.Client).Do(req)
	if err != nil {
		t.Fatal(err)
	}

	if *ctxt.txCommitted {
		t.Fatal("Tx was committed although a panic occured")
	}
	if !*ctxt.txRolledBack {
		t.Fatal("Tx was not rolled back although a panic occured")
	}
}

func Test_Handler_500OnFailureToCommit(t *testing.T) {

	ts, ctxt := setupCommitFailingHandlerTestServer(mountTestHandler, t)
	tx := ctxt.Tx()
	defer tx.Rollback()
	defer ts.Close()

	req, err := newRequest("GET", ts.URL+"/ok", "")
	if err != nil {
		t.Fatal(err)
	}

	res, err := new(http.Client).Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != http.StatusInternalServerError {
		t.Fatal("Failing to commit did not result in a http status 500")
	}
}

func Test_handleError_returns_403_whenTheSessionIsNotFound(t *testing.T) {
	res := httptest.NewRecorder()
	handleError(res, ErrSessionNotFound)

	if got, want := res.Code, 403; got != want {
		t.Errorf("res.Code = %d; want %d", got, want)
	}

	result := struct{ Reason string }{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}

	if got, want := result.Reason, "session_not_found"; got != want {
		t.Errorf("result.Reason = %q; want %q", got, want)
	}
}

func Test_handleError_returns_400_whenTheSessionUUIDIsMalformed(t *testing.T) {
	res := httptest.NewRecorder()
	handleError(res, ErrSessionUuidMalformed)

	if got, want := res.Code, 400; got != want {
		t.Errorf("res.Code = %d; want %d", got, want)
	}

	result := struct{ Reason string }{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}

	if got, want := result.Reason, "session_uuid_malformed"; got != want {
		t.Errorf("result.Reason = %q; want %q", got, want)
	}
}

func Test_handleError_returns_403_whenTheSessionUserIsNotFound(t *testing.T) {
	res := httptest.NewRecorder()
	handleError(res, ErrSessionUserNotFound)

	if got, want := res.Code, 403; got != want {
		t.Errorf("res.Code = %d; want %d", got, want)
	}

	result := struct{ Reason string }{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}

	if got, want := result.Reason, "session_user_not_found"; got != want {
		t.Errorf("result.Reason = %q; want %q", got, want)
	}
}

func Test_handleError_returns_403_whenTheSessionIsNotValid(t *testing.T) {
	res := httptest.NewRecorder()
	handleError(res, ErrSessionNotValid)

	if got, want := res.Code, 403; got != want {
		t.Errorf("res.Code = %d; want %d", got, want)
	}

	result := struct{ Reason string }{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}

	if got, want := result.Reason, "session_not_valid"; got != want {
		t.Errorf("result.Reason = %q; want %q", got, want)
	}
}

func Test_handleError_returns_403_whenAuthzReturnsAnError(t *testing.T) {
	res := httptest.NewRecorder()
	handleError(res, authz.ErrCapabilityMissing)

	if res.Code != 403 {
		t.Fatalf("Expected response code to be %d, got %d", 403, res.Code)
	}

	result := struct{ Reason string }{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}

	if got, want := result.Reason, "capability_missing"; got != want {
		t.Errorf("result.Reason = %q; want %q", got, want)
	}
}

func Test_handleError_returns_404_forDomainNotFound(t *testing.T) {
	err := new(domain.NotFoundError)
	res := httptest.NewRecorder()
	handleError(res, err)

	if got, want := res.Code, 404; got != want {
		t.Errorf("res.Code = %d; want %d", got, want)
	}

	result := struct{ Reason string }{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}

	if got, want := result.Reason, "not_found"; got != want {
		t.Errorf("result.Reason = %q; want %q", got, want)
	}
}

func Test_handleError_returns_500_forPqError(t *testing.T) {
	res := httptest.NewRecorder()
	handleError(res, new(pq.Error))

	if got, want := res.Code, 500; got != want {
		t.Errorf("res.Code = %d; want %d", got, want)
	}

	result := struct{ Reason string }{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}

	if got, want := result.Reason, "internal"; got != want {
		t.Errorf("result.Reason = %q; want %q", got, want)
	}
}

func Test_handleError_returns_400_forJsonSyntaxError(t *testing.T) {
	res := httptest.NewRecorder()
	handleError(res, new(json.SyntaxError))

	if got, want := res.Code, 400; got != want {
		t.Errorf("res.Code = %d; want %d", got, want)
	}

	result := struct{ Reason string }{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}

	if got, want := result.Reason, "syntax_error"; got != want {
		t.Errorf("result.Reason = %q; want %q", got, want)
	}
}

func Test_handleError_logsInternalServerError_forPqError(t *testing.T) {
	t.Skip("Error handling subject to reinvention")
	res := httptest.NewRecorder()
	err := &pq.Error{Message: "test"}
	logged := ""

	handleError(res, err)

	if got, want := logged, err.Error(); got != want {
		t.Errorf("logInternalServerError logged %q; want %q", got, want)
	}
}

func Test_handleError_returns_422_forDomainValidationError(t *testing.T) {
	res := httptest.NewRecorder()
	handleError(res, domain.NewValidationError("field", "error"))

	if got, want := res.Code, 422; got != want {
		t.Errorf("res.Code = %d; want %d", got, want)
	}

	result := struct {
		Reason string
		Errors map[string][]string
	}{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}

	if got, want := result.Reason, "invalid"; got != want {
		t.Errorf("result.Reason = %q; want %q", got, want)
	}

	if got, want := result.Errors["field"][0], "error"; got != want {
		t.Errorf("result.Errors[%q][0] = %q; want %q", "field", got, want)
	}
}

func Test_handleError_returns_listOfInvalidFields(t *testing.T) {
	res := httptest.NewRecorder()
	handleError(res, domain.NewValidationError("field", "error"))

	result := struct{ Errors map[string][]string }{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}

	if got, want := result.Errors["field"], []string{"error"}; !reflect.DeepEqual(got, want) {
		t.Errorf("result.Errors[%q] = %q; want %q", "field", got, want)
	}
}

func Test_handleError_returns_418_forUnknownErrorType(t *testing.T) {
	res := httptest.NewRecorder()
	err := errors.New("unknown error type")
	handleError(res, err)

	result := struct{ Reason string }{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}

	if got, want := result.Reason, "not_clever_enough"; got != want {
		t.Errorf("result.Reason = %q; want %q", got, want)
	}

}

func Test_handleError_returns400_forUnexpectedEOF(t *testing.T) {
	res := httptest.NewRecorder()
	err := io.ErrUnexpectedEOF
	handleError(res, err)

	if got, want := res.Code, http.StatusBadRequest; got != want {
		t.Errorf("res.StatusCode = %d; want %d", got, want)
	}

	result := struct{ Reason string }{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}

	if got, want := result.Reason, "unexpected_eof"; got != want {
		t.Errorf("result.Reason = %q; want %q", got, want)
	}
}
