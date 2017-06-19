package http

import (
	"testing"

	"github.com/gorilla/mux"
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/stores"
	"github.com/harrowio/harrow/test_helpers"
)

func Test_ScheduleHandler_Routing(t *testing.T) {
	r := mux.NewRouter()
	MountScheduleHandler(r, nil)

	spec := routingSpec{
		{"POST", "/schedules", "schedule-create"},
		{"PUT", "/schedules", "schedule-edit"},
		{"DELETE", "/schedules/:uuid", "schedule-delete"},
		{"GET", "/schedules/:uuid", "schedule-show"},
	}

	spec.run(r, t)
}

func Test_ScheduleHandler_Create_emitsJobScheduledActivity(t *testing.T) {
	h := NewHandlerTest(MountScheduleHandler, t)
	defer h.Cleanup()

	h.LoginAs("default")
	job := h.World().Job("default")
	timespec := "now"

	h.Do("POST", h.Url("/schedules"), &schedParamsWrapper{
		Subject: schedParams{
			JobUuid:     job.Uuid,
			Timespec:    &timespec,
			Description: "Do it!",
		},
	})

	t.Logf("Response body:\n%s\n", h.ResponseBody())

	for _, activity := range h.Activities() {
		t.Logf("Activity: %s\n", activity.Name)
		if activity.Name == "job.scheduled" {
			return
		}
	}

	t.Fatalf("Activity %q not found", "job.scheduled")
}

func Test_ScheduleHandler_Delete_emitsScheduleDeletedActivity(t *testing.T) {
	h := NewHandlerTest(MountScheduleHandler, t)
	defer h.Cleanup()

	h.LoginAs("default")
	job := h.World().Job("default")
	timespec := "now"
	schedule := test_helpers.MustCreateSchedule(t, h.Tx(), &domain.Schedule{
		Uuid:        "535f8a26-06b6-420c-8c04-f45f39fb485c",
		Timespec:    &timespec,
		JobUuid:     job.Uuid,
		Parameters:  domain.NewOperationParameters(),
		UserUuid:    h.World().User("default").Uuid,
		Description: "Do it!",
	})

	h.Subject(schedule)
	h.Do("DELETE", h.UrlFor("self"), nil)
	for _, activity := range h.Activities() {
		t.Logf("Activity: %s\n", activity.Name)
		if activity.Name == "schedule.deleted" {
			return
		}
	}

	t.Logf("Response body:\n%s\n", h.ResponseBody())

	t.Fatalf("Activity %q not found", "schedule.deleted")
}

func Test_ScheduleHandler_Delete_requires_user_to_be_logged_in(t *testing.T) {
	h := NewHandlerTest(MountScheduleHandler, t)
	defer h.Cleanup()

	job := h.World().Job("default")
	timespec := "now"
	schedule := test_helpers.MustCreateSchedule(t, h.Tx(), &domain.Schedule{
		Uuid:        "535f8a26-06b6-420c-8c04-f45f39fb485c",
		Timespec:    &timespec,
		JobUuid:     job.Uuid,
		Parameters:  domain.NewOperationParameters(),
		UserUuid:    h.World().User("default").Uuid,
		Description: "Do it!",
	})

	apiError := &ErrorJSON{}

	h.LogOut()
	h.Subject(schedule)
	h.ResultTo(apiError)
	h.Do("DELETE", h.UrlFor("self"), nil)
	t.Logf("Response body:\n%s\n", h.ResponseBody())
	if got, want := apiError.Reason, "login_required"; got != want {
		t.Errorf(`apiError.Reason = %v; want %v`, got, want)
	}
}

func Test_ScheduleHandler_Delete_deletes_schedule(t *testing.T) {
	h := NewHandlerTest(MountScheduleHandler, t)
	defer h.Cleanup()

	h.LoginAs("default")
	job := h.World().Job("default")
	timespec := "now"
	schedule := test_helpers.MustCreateSchedule(t, h.Tx(), &domain.Schedule{
		Uuid:        "535f8a26-06b6-420c-8c04-f45f39fb485c",
		Timespec:    &timespec,
		JobUuid:     job.Uuid,
		Parameters:  domain.NewOperationParameters(),
		UserUuid:    h.World().User("default").Uuid,
		Description: "Do it!",
	})

	h.Subject(schedule)
	h.Do("DELETE", h.UrlFor("self"), nil)
	t.Logf("Response body:\n%s\n", h.ResponseBody())

	_, err := stores.NewDbScheduleStore(h.Tx()).FindByUuid(schedule.Uuid)
	if !domain.IsNotFound(err) {
		t.Fatal("Expected schedule to be deleted")
	}
}

func Test_ScheduleHandler_Delete_returns_404_if_schedule_is_not_found(t *testing.T) {
	h := NewHandlerTest(MountScheduleHandler, t)
	defer h.Cleanup()

	h.LoginAs("default")
	job := h.World().Job("default")
	timespec := "now"
	schedule := test_helpers.MustCreateSchedule(t, h.Tx(), &domain.Schedule{
		Uuid:        "535f8a26-06b6-420c-8c04-f45f39fb485c",
		Timespec:    &timespec,
		JobUuid:     job.Uuid,
		Parameters:  domain.NewOperationParameters(),
		UserUuid:    h.World().User("default").Uuid,
		Description: "Do it!",
	})

	h.Subject(schedule)
	h.Do("DELETE", h.UrlFor("self"), nil)
	t.Logf("Response body:\n%s\n", h.ResponseBody())

	apiError := &ErrorJSON{}
	h.ResultTo(apiError)
	h.Do("DELETE", h.UrlFor("self"), nil)

	if got, want := h.Response().StatusCode, 404; got != want {
		t.Errorf(`h.Response().StatusCode = %v; want %v`, got, want)
	}

	if got, want := apiError.Reason, "not_found"; got != want {
		t.Errorf(`apiError.Reason = %v; want %v`, got, want)
	}
}

func Test_ScheduleHandler_Edit_returns_404_if_schedule_does_not_exist(t *testing.T) {
	h := NewHandlerTest(MountScheduleHandler, t)
	defer h.Cleanup()

	h.LoginAs("default")
	job := h.World().Job("default")
	timespec := "now"

	apiError := &ErrorJSON{}
	h.ResultTo(apiError)
	h.Do("PUT", h.Url("/schedules"), &schedParamsWrapper{
		Subject: schedParams{
			Uuid:     "ea61abb1-8b28-4ff6-b5b3-eeb239edd458",
			JobUuid:  job.Uuid,
			Timespec: &timespec,
		},
	})
	t.Logf("Response body:\n%s\n", h.ResponseBody())

	if got, want := h.Response().StatusCode, 404; got != want {
		t.Errorf(`h.Response().StatusCode = %v; want %v`, got, want)
	}

	if got, want := apiError.Reason, "not_found"; got != want {
		t.Errorf(`apiError.Reason = %v; want %v`, got, want)
	}

}

func Test_ScheduleHandler_Edit_returns_error_if_user_is_logged_out(t *testing.T) {
	h := NewHandlerTest(MountScheduleHandler, t)
	defer h.Cleanup()

	h.LogOut()
	job := h.World().Job("default")
	timespec := "now"

	apiError := &ErrorJSON{}
	h.ResultTo(apiError)
	h.Do("PUT", h.Url("/schedules"), &schedParamsWrapper{
		Subject: schedParams{
			Uuid:     "ea61abb1-8b28-4ff6-b5b3-eeb239edd458",
			JobUuid:  job.Uuid,
			Timespec: &timespec,
		},
	})
	t.Logf("Response body:\n%s\n", h.ResponseBody())

	if got, want := h.Response().StatusCode, 403; got != want {
		t.Errorf(`h.Response().StatusCode = %v; want %v`, got, want)
	}

	if got, want := apiError.Reason, "login_required"; got != want {
		t.Errorf(`apiError.Reason = %v; want %v`, got, want)
	}

}

func Test_ScheduleHandler_Edit_emits_schedule_edited_activity(t *testing.T) {
	h := NewHandlerTest(MountScheduleHandler, t)
	defer h.Cleanup()

	h.LoginAs("default")
	job := h.World().Job("default")
	timespec := "now"
	in5Minutes := "now + 5 minutes"
	schedule := test_helpers.MustCreateSchedule(t, h.Tx(), &domain.Schedule{
		Uuid:        "535f8a26-06b6-420c-8c04-f45f39fb485c",
		Timespec:    &timespec,
		JobUuid:     job.Uuid,
		Parameters:  domain.NewOperationParameters(),
		UserUuid:    h.World().User("default").Uuid,
		Description: "Do it!",
	})

	h.Subject(schedule)
	h.Do("PUT", h.Url("/schedules"), &schedParamsWrapper{
		Subject: schedParams{
			Uuid:     schedule.Uuid,
			Timespec: &in5Minutes,
		},
	})
	for _, activity := range h.Activities() {
		t.Logf("Activity: %s\n", activity.Name)
		if activity.Name == "schedule.edited" {
			return
		}
	}

	t.Logf("Response body:\n%s\n", h.ResponseBody())

	t.Fatalf("Activity %q not found", "schedule.edited")
}

func Test_ScheduleHandler_Edit_updates_schedule(t *testing.T) {
	h := NewHandlerTest(MountScheduleHandler, t)
	defer h.Cleanup()

	h.LoginAs("default")
	job := h.World().Job("default")
	timespec := "now"
	in5Minutes := "now + 5 minutes"
	newDescription := "Deploy later"

	schedule := test_helpers.MustCreateSchedule(t, h.Tx(), &domain.Schedule{
		Uuid:        "535f8a26-06b6-420c-8c04-f45f39fb485c",
		Timespec:    &timespec,
		JobUuid:     job.Uuid,
		Parameters:  domain.NewOperationParameters(),
		UserUuid:    h.World().User("default").Uuid,
		Description: "Do it!",
	})

	h.Subject(schedule)
	h.Do("PUT", h.Url("/schedules"), &schedParamsWrapper{
		Subject: schedParams{
			Uuid:        schedule.Uuid,
			Timespec:    &in5Minutes,
			Description: newDescription,
		},
	})
	t.Logf("Response body:\n%s\n", h.ResponseBody())

	updatedSchedule, err := stores.NewDbScheduleStore(h.Tx()).FindByUuid(schedule.Uuid)
	if err != nil {
		t.Fatal(err)
	}

	if got := updatedSchedule.Timespec; got == nil {
		t.Fatalf(`updatedSchedule.Timespec is nil`)
	}

	if got, want := *updatedSchedule.Timespec, in5Minutes; got != want {
		t.Errorf(`*updatedSchedule.Timespec = %v; want %v`, got, want)
	}

	if got, want := updatedSchedule.Description, newDescription; got != want {
		t.Errorf(`updatedSchedule.Description = %v; want %v`, got, want)
	}
}
