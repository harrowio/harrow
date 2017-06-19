package http

import (
	"testing"

	"github.com/gorilla/mux"
)

func Test_TaskHandler_Routing(t *testing.T) {
	r := mux.NewRouter()
	MountTaskHandler(r, nil)

	spec := routingSpec{
		{"PUT", "/tasks", "task-update"},
		{"POST", "/tasks", "task-create"},
		{"GET", "/tasks/:uuid/jobs", "task-jobs"},
		{"GET", "/tasks/:uuid", "task-show"},
		{"DELETE", "/tasks/:uuid", "task-archive"},
	}

	spec.run(r, t)
}

func Test_TaskHandler_CreateUpdate_emitsTaskCreateActivity_whenCreatingATask(t *testing.T) {
	h := NewHandlerTest(MountTaskHandler, t)
	defer h.Cleanup()

	h.LoginAs("default")

	project := h.World().Project("public")

	h.Do("POST", h.Url("/tasks"), &taskParamsWrapper{
		Subject: taskParams{
			Name:        "created",
			Body:        "#!/bin/bash\ndate\n",
			ProjectUuid: project.Uuid,
			Type:        "script",
		},
	})

	for _, activity := range h.Activities() {
		t.Logf("Activity: %s\n", activity.Name)
		if activity.Name == "task.added" {
			return
		}
	}

	t.Fatalf("Activity %q not found", "task.added")
}

func Test_TaskHandler_CreateUpdate_emitsTaskEditedActivity_whenUpdatingATask(t *testing.T) {
	h := NewHandlerTest(MountTaskHandler, t)
	defer h.Cleanup()

	h.LoginAs("default")

	task := h.World().Task("default")

	h.Do("PUT", h.Url("/tasks"), &taskParamsWrapper{
		Subject: taskParams{
			Uuid:        task.Uuid,
			Name:        task.Name,
			Body:        "#!/bin/bash\ndate\n",
			ProjectUuid: task.ProjectUuid,
			Type:        "script",
		},
	})

	for _, activity := range h.Activities() {
		t.Logf("Activity: %s\n", activity.Name)
		if activity.Name == "task.edited" {
			return
		}
	}

	t.Fatalf("Activity %q not found", "task.edited")
}
