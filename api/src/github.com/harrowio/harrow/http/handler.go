package http

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/harrowio/harrow/activities"
	"github.com/harrowio/harrow/stores"
	"github.com/lib/pq"

	"github.com/harrowio/harrow/authz"
	"github.com/harrowio/harrow/domain"
)

var (
	nullSink                  = NewNullSink()
	activitySink ActivitySink = nullSink
)

func withActivitySink(sink ActivitySink, do func()) {
	original := activitySink
	defer func() {
		err := recover()
		activitySink = original
		if err != nil {
			panic(err)
		}
	}()
	activitySink = sink
	do()
}

const (
	StatusUnprocessableEntity = 422 // from WebDAV
)

func handleHttpError(w http.ResponseWriter, err *Error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(err.status)
	w.Write(err.AsJSON())
}

func handleError(w http.ResponseWriter, err error) {
	if err == io.ErrUnexpectedEOF {
		handleError(w, ErrUnexpectedEOF)
		return
	}

	switch e := err.(type) {
	case *domain.NotFoundError:
		handleHttpError(w, ErrNotFound)
		return
	case *authz.Error:
		if e.Reason() == "capability_missing" {
			handleHttpError(w, NewCapabilityMissingError(e.HaveCapabilities, e.MissingCapability))
		} else {
			handleHttpError(w, NewError(403, e.Reason(), e.Error()))
		}
		return
	case *json.SyntaxError:
		handleHttpError(w, NewError(400, "syntax_error", err.Error()))
		return
	case *json.UnmarshalTypeError:
		handleHttpError(w, NewError(400, "unmarshal_type_error", err.Error()))
		return
	case *domain.ValidationError:
		handleHttpError(w, NewValidationError(e))
		return
	case *pq.Error:
		handleHttpError(w, NewInternalError(err))
		return
	case *Error:
		handleHttpError(w, e)
		return
	default:
		// 418 in the client indicates that this method
		// is not clever enough to handle all errors
		handleHttpError(w, NewError(418, "not_clever_enough", err.Error()))
	}
}

func markUserActivity(ctxt RequestContext) error {
	if ctxt.User() == nil {
		return nil
	}

	currentUser := ctxt.User()
	activityStore := stores.NewDbActivityStore(ctxt.Tx())

	userHistory := domain.NewUserHistory(currentUser.Uuid, time.Now())
	if err := activityStore.AllByUser(currentUser.Uuid, userHistory.HandleActivity); err != nil {
		return err
	}

	previousSegment := userHistory.PreviousSegment()
	if newSegment := userHistory.Segment(); newSegment != previousSegment {
		if previousSegment != "" {
			ctxt.EnqueueActivity(activities.UserLeftSegment(previousSegment), &currentUser.Uuid)
		}

		ctxt.EnqueueActivity(activities.UserEnteredSegment(newSegment), &currentUser.Uuid)
	}

	if !userHistory.IsActive() {
		ctxt.EnqueueActivity(activities.UserReportedAsActive(currentUser), &currentUser.Uuid)
	}

	return nil
}

func HandlerFunc(ctxt ServerContext, h Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		requestContext, err := ctxt.RequestContext(w, r)
		if err != nil {
			handleError(w, err)
			return
		}

		// If no errors occured during request processing, we will commit the tx
		// (see below) before this runs making it a no-op.
		defer requestContext.RollbackTx()

		defer func(w http.ResponseWriter) {
			if r := recover(); r != nil {
				ctxt.Log().Info().Msgf("recovered from panic: %v\n", r)
				handleError(w, ErrRecoveredFromPanic)
			}
		}(w)

		if handlerErr := h(requestContext); handlerErr != nil {
			handleError(w, handlerErr)
			return
		}

		if err := markUserActivity(requestContext); err != nil {
			handleError(w, err)
			return
		}

		if err := requestContext.CommitTx(); err != nil {
			handleError(w, NewError(500, "transaction_commit_failed", err.Error()))
			return
		}

		for _, activity := range requestContext.Activities() {
			activitySink.EmitActivity(activity)
		}
	}

	// From here on (owing to above) panics will be caught, and
	// errors correctly handled

	return http.HandlerFunc(fn)
}
