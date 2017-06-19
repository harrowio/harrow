package http

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

type promptStatus struct {
	Key       string `json:"key"`
	Dismissed bool   `json:"dismissed"`
}

func MountPromptHandler(r *mux.Router, ctxt ServerContext) {
	h := promptHandler{}

	// Collection
	root := r.PathPrefix("/prompts").Subrouter()

	// Item
	item := root.PathPrefix("/{key}").Subrouter()
	item.Methods("GET").Handler(HandlerFunc(ctxt, h.Show)).
		Name("prompt-show")
	item.Methods("POST").Handler(HandlerFunc(ctxt, h.Post)).
		Name("prompt-create")
	item.Methods("DELETE").Handler(HandlerFunc(ctxt, h.Delete)).
		Name("prompt-delete")

	root.Methods("GET").Handler(HandlerFunc(ctxt, h.Index)).
		Name("prompt-index")
}

type promptHandler struct {
}

func (self promptHandler) Index(ctxt RequestContext) error {

	currentUser := ctxt.User()

	members, err := ctxt.KeyValueStore().SMembers(currentUser.Uuid)
	if err != nil {
		return err
	} else {
		ctxt.W().Header().Set("Content-Type", "application/json")
		var memberStatuses []promptStatus
		for _, member := range members {
			memberStatuses = append(memberStatuses, promptStatus{
				Key:       member,
				Dismissed: true,
			})
		}

		jsonResponse, _ := json.Marshal(memberStatuses)

		ctxt.W().Write(jsonResponse)

		return nil
	}
}

func (self promptHandler) Show(ctxt RequestContext) (err error) {

	promptKey := ctxt.PathParameter("key")
	currentUser := ctxt.User()

	dismissed, err := ctxt.KeyValueStore().SIsMember(currentUser.Uuid, promptKey)
	if err != nil {
		return err
	} else {
		ctxt.W().Header().Set("Content-Type", "application/json")
		jsonResponse, _ := json.Marshal(promptStatus{
			Key:       promptKey,
			Dismissed: dismissed,
		})
		ctxt.W().Write(jsonResponse)

		return nil
	}
}

func (self promptHandler) Post(ctxt RequestContext) error {

	promptKey := ctxt.PathParameter("key")
	currentUser := ctxt.User()

	_, err := ctxt.KeyValueStore().SRem(currentUser.Uuid, promptKey)
	if err != nil {
		return err
	} else {
		ctxt.W().WriteHeader(http.StatusOK)
		return nil
	}

}

func (self promptHandler) Delete(ctxt RequestContext) error {

	promptKey := ctxt.PathParameter("key")
	currentUser := ctxt.User()

	if promptKey == "all" {
		err := ctxt.KeyValueStore().Del(currentUser.Uuid)
		if err != nil {
			return err
		} else {
			ctxt.W().WriteHeader(http.StatusOK)
			return nil
		}
	}

	_, err := ctxt.KeyValueStore().SAdd(currentUser.Uuid, promptKey)
	if err != nil {
		return err
	} else {
		ctxt.W().WriteHeader(http.StatusOK)
		return nil
	}

}
