package http

import (
	"encoding/json"
	"errors"

	"github.com/gorilla/mux"

	"github.com/harrowio/harrow/activities"
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/stores"
	"github.com/harrowio/harrow/uuidhelper"
)

type projectMemberHandler struct {
	ctxt    RequestContext
	current *domain.ProjectMember
	param   *domain.ProjectMember
	stores  struct {
		organizations           *stores.DbOrganizationStore
		projects                *stores.DbProjectStore
		projectMemberships      *stores.DbProjectMembershipStore
		organizationMemberships *stores.DbOrganizationMembershipStore
		users                   *stores.DbUserStore
	}
}

func MountProjectMemberHandler(r *mux.Router, ctxt ServerContext) {

	ph := &projectMemberHandler{}

	// Collection
	root := r.PathPrefix("/project-members").Subrouter()
	item := root.PathPrefix("/{uuid}").Subrouter()

	item.Methods("DELETE").Handler(HandlerFunc(ctxt, ph.Remove)).
		Name("project-member-remove")

	root.Methods("PUT").Handler(HandlerFunc(ctxt, ph.Update)).
		Name("project-member-update")

}

func (self *projectMemberHandler) init(ctxt RequestContext) (*projectMemberHandler, error) {
	h := &projectMemberHandler{
		ctxt: ctxt,
	}

	tx := ctxt.Tx()
	c := ctxt.Config()
	h.stores.organizationMemberships = stores.NewDbOrganizationMembershipStore(tx)
	h.stores.projects = stores.NewDbProjectStore(tx)
	h.stores.projectMemberships = stores.NewDbProjectMembershipStore(tx)
	h.stores.users = stores.NewDbUserStore(tx, &c)

	return h, nil
}

func (self *projectMemberHandler) loadMember(projectUuid, userUuid string) (*domain.ProjectMember, error) {
	user, err := self.stores.users.FindByUuid(userUuid)
	if err != nil {
		return nil, err
	}

	project, err := self.stores.projects.FindByUuid(projectUuid)
	if err != nil {
		return nil, err
	}

	projectMembership, err := self.stores.projectMemberships.FindByUserAndProjectUuid(userUuid, projectUuid)
	if err != nil {
		if _, ok := err.(*domain.NotFoundError); !ok {
			return nil, err
		}
	}

	orgMembership, err := self.stores.organizationMemberships.FindByOrganizationAndUserUuids(project.OrganizationUuid, userUuid)
	if err != nil {
		if _, ok := err.(*domain.NotFoundError); !ok {
			return nil, err
		}
	}

	member := domain.NewProjectMember(user, project, projectMembership, orgMembership)
	return member, nil
}

func (self *projectMemberHandler) Update(ctxt RequestContext) error {
	h, err := self.init(ctxt)
	if err != nil {
		return err
	}

	params := &domain.ProjectMember{}
	wrapper := &halWrapper{Subject: params}
	if err := json.NewDecoder(ctxt.R().Body).Decode(&wrapper); err != nil {
		return err
	}

	promoter, err := h.loadMember(params.ProjectUuid, ctxt.User().Uuid)
	if err != nil {
		return err
	}

	toPromote, err := h.loadMember(params.ProjectUuid, params.User.Uuid)
	if err != nil {
		return err
	}

	if err := promoter.Promote(toPromote); err != nil {
		return err
	}

	newMembership := toPromote.ToMembership()
	if len(newMembership.Uuid) > 0 {
		if err := h.stores.projectMemberships.Update(newMembership); err != nil {
			return err
		}
	} else {
		if _, err := h.stores.projectMemberships.Create(newMembership); err != nil {
			return err
		}
	}

	writeAsJson(ctxt, toPromote)
	return nil

}

func (self *projectMemberHandler) Remove(ctxt RequestContext) error {
	h, err := self.init(ctxt)
	if err != nil {
		return err
	}

	params := struct {
		ProjectUuid string `json:"projectUuid"`
	}{
		ProjectUuid: ctxt.R().FormValue("projectUuid"),
	}

	if !uuidhelper.IsValid(params.ProjectUuid) {
		return NewMalformedParameters("projectUuid", errors.New("invalid_uuid"))
	}

	remover, err := h.loadMember(params.ProjectUuid, ctxt.User().Uuid)
	if err != nil {
		return err
	}

	toBeRemoved, err := h.loadMember(params.ProjectUuid, ctxt.PathParameter("uuid"))
	if err != nil {
		return err
	}

	if err := remover.Remove(toBeRemoved, h.stores.projectMemberships); err != nil {
		return err
	}

	project, err := h.stores.projects.FindByUuid(params.ProjectUuid)
	if err != nil {
		return err
	}

	if remover.User.Uuid == toBeRemoved.User.Uuid {
		ctxt.EnqueueActivity(activities.UserLeftProject(toBeRemoved.User, project), nil)
	} else {
		ctxt.EnqueueActivity(activities.UserRemovedFromProject(toBeRemoved.User, project), nil)
	}

	return nil
}
