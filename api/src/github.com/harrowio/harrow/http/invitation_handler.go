package http

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/harrowio/harrow/activities"
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/stores"
	"github.com/harrowio/harrow/uuidhelper"
)

type CreateInvitationParams struct {
	RecipientName  string `json:"recipientName"`
	Email          string `json:"email"`
	InviteeUuid    string `json:"inviteeUuid"`
	ProjectUuid    string `json:"projectUuid"`
	MembershipType string `json:"membershipType"`
	Message        string `json:"message"`
}

type PatchInvitationParams struct {
	Accept string `json:"accept"`
}

func (p *PatchInvitationParams) AcceptsInvitation() bool {
	return p.Accept == "accept"
}

func (p *PatchInvitationParams) RefusesInvitation() bool {
	return p.Accept == "refuse"
}

type invitationHandler struct{}

func MountInvitationHandler(r *mux.Router, ctxt ServerContext) {
	h := invitationHandler{}

	// Collection
	root := r.PathPrefix("/invitations").Subrouter()
	root.Methods("POST").Handler(HandlerFunc(ctxt, h.Create)).
		Name("invitation-create")

	// Item
	item := root.PathPrefix("/{uuid}").Subrouter()
	item.Methods("GET").Handler(HandlerFunc(ctxt, h.Show)).
		Name("invitation-show")
	item.Methods("PATCH").Handler(HandlerFunc(ctxt, h.Accept)).
		Name("invitation-accept")
}

func (h *invitationHandler) Create(ctxt RequestContext) error {
	store := stores.NewDbInvitationStore(ctxt.Tx())
	c := ctxt.Config()
	projectStore := stores.NewDbProjectStore(ctxt.Tx())
	userStore := stores.NewDbUserStore(ctxt.Tx(), &c)

	params := &CreateInvitationParams{}

	if err := decodeHALParams(ctxt.R().Body, params); err != nil {
		return err
	}

	if !uuidhelper.IsValid(params.ProjectUuid) {
		return domain.NewValidationError("projectUuid", "malformed")
	}
	project, err := projectStore.FindByUuid(params.ProjectUuid)
	if err != nil {
		switch err.(type) {
		case *domain.NotFoundError:
			return domain.NewValidationError("projectUuid", "not found")
		default:
			return err
		}
	}

	inviteeUuid := params.InviteeUuid
	var inv *domain.Invitation

	if inviteeUuid == "" {
		inv = project.NewInvitationToHarrow(
			params.RecipientName,
			params.Email,
			params.Message,
			params.MembershipType,
		)
	} else {
		invitee, err := userStore.FindByUuid(params.InviteeUuid)
		if err != nil {
			switch err.(type) {
			case *domain.NotFoundError:
				return domain.NewValidationError("inviteeUuid", "not found")
			default:
				return err
			}
		}
		inv = project.NewInvitationToUser(params.Message, params.MembershipType, invitee)
	}
	inv.CreatorUuid = ctxt.User().Uuid

	if err := inv.Validate(); err != nil {
		return err
	}

	if allowed, err := ctxt.Auth().CanCreate(inv); !allowed {
		return err
	}

	uuid, err := store.Create(inv)
	if err != nil {
		return err
	}

	ctxt.EnqueueActivity(activities.InvitationCreated(inv), nil)

	inv, err = store.FindByUuid(uuid)
	if err != nil {
		return err
	}

	ctxt.W().WriteHeader(http.StatusCreated)
	writeAsJson(ctxt, inv)
	return nil
}

func (h *invitationHandler) Show(ctxt RequestContext) error {
	invitationUuid := ctxt.PathParameter("uuid")
	invitationStore := stores.NewDbInvitationStore(ctxt.Tx())
	projectStore := stores.NewDbProjectStore(ctxt.Tx())
	organizationStore := stores.NewDbOrganizationStore(ctxt.Tx())
	c := ctxt.Config()
	userStore := stores.NewDbUserStore(ctxt.Tx(), &c)
	invitation, err := invitationStore.FindByUuid(invitationUuid)
	if err != nil {
		return err
	}
	project, err := projectStore.FindByUuid(invitation.ProjectUuid)
	if err != nil {
		return err
	}

	organization, err := organizationStore.FindByUuid(invitation.OrganizationUuid)
	if err != nil {
		return err
	}

	creator, err := userStore.FindByUuid(invitation.CreatorUuid)
	if err != nil {
		return err
	}

	response := struct {
		*domain.Invitation
		CreatorName      string `json:"creatorName"`
		ProjectName      string `json:"projectName"`
		OrganizationName string `json:"organizationName"`
	}{invitation, creator.Name, project.Name, organization.Name}

	writeAsJson(ctxt, &response)
	return nil
}

func (h *invitationHandler) Accept(ctxt RequestContext) error {
	invitationUuid := ctxt.PathParameter("uuid")
	params := &PatchInvitationParams{}
	dec := json.NewDecoder(ctxt.R().Body)
	if err := dec.Decode(params); err != nil {
		return err
	}

	invitationStore := stores.NewDbInvitationStore(ctxt.Tx())
	projectMembershipStore := stores.NewDbProjectMembershipStore(ctxt.Tx())
	invitation, err := invitationStore.FindByUuid(invitationUuid)
	if err != nil {
		return err
	}

	project, err := stores.NewDbProjectStore(ctxt.Tx()).FindByUuid(invitation.ProjectUuid)
	if err != nil {
		return err
	}

	if !invitation.IsOpen() {
		return domain.NewValidationError("invitation", "not open")
	}

	if params.AcceptsInvitation() {
		invitation.Accept(ctxt.User())

		if _, err := projectMembershipStore.Create(invitation.NewProjectMembership()); err != nil {
			return err
		}
		ctxt.EnqueueActivity(activities.InvitationAccepted(invitation), nil)
		ctxt.EnqueueActivity(activities.UserJoinedProject(ctxt.User(), project), nil)
	} else if params.RefusesInvitation() {
		invitation.Refuse()
		ctxt.EnqueueActivity(activities.InvitationRefused(invitation), nil)
	} else {
		return domain.NewValidationError("accept", "invalid")
	}

	if err := invitationStore.Update(invitation); err != nil {
		return err
	}

	ctxt.W().WriteHeader(http.StatusOK)
	writeAsJson(ctxt, invitation)
	return nil
}
