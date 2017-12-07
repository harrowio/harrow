package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/harrowio/harrow/activities"
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/stores"
	"github.com/harrowio/harrow/uuidhelper"
)

type OrgParamsWrapper struct {
	Subject OrgParams
}

type OrgParams struct {
	Uuid        string `json:"uuid"`
	Name        string `json:"name"`
	Public      bool   `json:"public"`
	GithubLogin string `json:"github_login"`
	PlanUuid    string `json:"planUuid"`
}

func ReadOrgParams(r io.Reader) (*OrgParamsWrapper, error) {
	dec := json.NewDecoder(r)
	var params OrgParamsWrapper
	if err := dec.Decode(&params); err != nil {
		return nil, err
	}
	return &params, nil
}

func copyOrgParams(params *OrgParams, model *domain.Organization) {
	model.Uuid = params.Uuid
	model.Name = params.Name
	model.Public = params.Public
	model.GithubLogin = params.GithubLogin
}

type orgHandler struct {
}

func MountOrganizationHandler(r *mux.Router, ctxt ServerContext) {
	h := &orgHandler{}

	// Collection
	root := r.PathPrefix("/organizations").Subrouter()
	root.Methods("POST").Handler(HandlerFunc(ctxt, h.CreateOrUpdate)).
		Name("organization-create")
	root.Methods("PUT").Handler(HandlerFunc(ctxt, h.CreateOrUpdate)).
		Name("organization-update")

	// Relationships
	related := root.PathPrefix("/{uuid}/").Subrouter()
	related.Methods("GET").Path("/projects").Handler(HandlerFunc(ctxt, h.Projects)).
		Name("organization-projects")
	related.Methods("GET").Path("/memberships").Handler(HandlerFunc(ctxt, h.Memberships)).
		Name("organization-memberships")
	related.Methods("GET").Path("/members").Handler(HandlerFunc(ctxt, h.Members)).
		Name("organization-members")
	related.Methods("GET").Path("/limits").Handler(HandlerFunc(ctxt, h.Limits)).
		Name("organization-limits")
	related.Methods("GET").Path("/project-cards").Handler(HandlerFunc(ctxt, h.ProjectCards)).
		Name("organization-project-cards")

	// Item
	item := root.Path("/{uuid}").Subrouter()
	item.Methods("GET").Handler(HandlerFunc(ctxt, h.Show)).
		Name("organization-show")
	item.Methods("DELETE").Handler(HandlerFunc(ctxt, h.Delete)).
		Name("organization-archive")

}

func (self orgHandler) Show(ctxt RequestContext) (err error) {

	var organizationUuid string = ctxt.PathParameter("uuid")

	tx := ctxt.Tx()
	orgStore := stores.NewDbOrganizationStore(tx)

	org, err := orgStore.FindByUuid(organizationUuid)
	if err != nil {
		return err
	}

	if allowed, err := ctxt.Auth().CanRead(org); !allowed {
		return err
	}

	billingHistory, err := stores.NewDbBillingHistoryStore(
		ctxt.Tx(),
		ctxt.KeyValueStore(),
	).Load()
	if err != nil {
		return err
	}

	limits := NewLimitsFromContext(ctxt)
	reported, err := limits.ForOrganizationUuid(org.Uuid)
	if err != nil {
		return err
	}

	org.Embed("limits", reported)

	result := struct {
		*domain.Organization
		BillingPlanUuid      string                 `json:"billingPlanUuid"`
		CreditCards          []*domain.CreditCard   `json:"creditCards"`
		BillingExtrasGranted []*domain.BillingEvent `json:"extrasGranted"`
	}{
		Organization:         org,
		BillingPlanUuid:      billingHistory.PlanUuidFor(org.Uuid),
		CreditCards:          billingHistory.CreditCardsFor(org.Uuid),
		BillingExtrasGranted: billingHistory.ExtrasGrantedTo(org.Uuid),
	}

	writeAsJson(ctxt, &result)

	return nil
}

func (self orgHandler) CreateOrUpdate(ctxt RequestContext) (err error) {

	tx := ctxt.Tx()
	currentUser := ctxt.User()

	params, err := ReadOrgParams(ctxt.R().Body)
	if err != nil {
		return err
	}

	store := stores.NewDbOrganizationStore(tx)
	orgMembershipStore := stores.NewDbOrganizationMembershipStore(tx)

	isNew := true
	var model *domain.Organization
	// If the UUID is valid, try to load the existing model
	if uuidhelper.IsValid(params.Subject.Uuid) {
		model, err = store.FindByUuid(params.Subject.Uuid)
		_, notFound := err.(*domain.NotFoundError)
		// handle errors that are not an instance of *domain.NotFoundError
		if err != nil && !notFound {
			return err
		}
		isNew = model == nil
	}

	if isNew {
		model = &domain.Organization{
			Uuid: uuidhelper.MustNewV4(), // required for auth
		}
		if allowed, err := ctxt.Auth().CanCreate(model); !allowed {
			return err
		}
	} else {
		if allowed, err := ctxt.Auth().CanUpdate(model); !allowed {
			return err
		}
	}
	copyOrgParams(&params.Subject, model)

	if model.Public == false && !uuidhelper.IsValid(params.Subject.PlanUuid) {
		return domain.NewValidationError("planUuid", "empty")
	}

	if err := domain.ValidateOrganization(model); err != nil {
		return err
	}

	var uuid string

	if isNew {
		uuid, err = store.Create(model)

		if err == nil {
			orgMembership := &domain.OrganizationMembership{
				OrganizationUuid: uuid,
				UserUuid:         currentUser.Uuid,
				Type:             domain.MembershipTypeOwner,
			}
			err = orgMembershipStore.Create(orgMembership)
			if model.Public == false {
				if err != nil {
					ctxt.Log().Warn().Msgf("error retrieving plan: %s", err)
					return err
				}
				subscriptionId := fmt.Sprintf("free:%s", uuid)
				events := stores.NewDbBillingEventStore(ctxt.Tx())
				eventData := &domain.BillingPlanSelected{
					UserUuid:       ctxt.User().Uuid,
					PlanUuid:       domain.FreePlanUuid,
					SubscriptionId: subscriptionId,
				}
				eventData.FillFromPlan(domain.FreePlan)
				planSelected := model.NewBillingEvent(eventData)
				if _, err := events.Create(planSelected); err != nil {
					ctxt.Log().Warn().Msgf("error saving billing event (%s), event data: %#v", err, planSelected)
				}

				ctxt.EnqueueActivity(activities.OrganizationCreated(model, domain.FreePlan), nil)
			} else {
				ctxt.EnqueueActivity(activities.OrganizationCreated(model, nil), nil)
			}
		}
	} else {
		err = store.Update(model)
		uuid = model.Uuid
	}
	if err != nil {
		return err
	}

	model, err = store.FindByUuid(uuid)
	if err != nil {
		return err
	}

	if isNew {
		ctxt.W().Header().Add("Location", urlForSubject(ctxt.R(), model))
		ctxt.W().WriteHeader(http.StatusCreated)
	}

	writeAsJson(ctxt, model)

	return nil
}

func (self orgHandler) Delete(ctxt RequestContext) (err error) {

	tx := ctxt.Tx()
	orgUuid := ctxt.PathParameter("uuid")

	orgStore := stores.NewDbOrganizationStore(tx)
	org, err := orgStore.FindByUuid(orgUuid)
	if err != nil {
		return err
	}

	if allowed, err := ctxt.Auth().CanArchive(org); !allowed {
		return err
	}

	if err := orgStore.ArchiveByUuid(orgUuid); err != nil {
		return err
	}

	ctxt.EnqueueActivity(activities.OrganizationArchived(org), nil)

	ctxt.W().WriteHeader(http.StatusNoContent)

	return nil
}

func (self orgHandler) Projects(ctxt RequestContext) (err error) {

	orgUuid := ctxt.PathParameter("uuid")

	tx := ctxt.Tx()
	projectStore := stores.NewDbProjectStore(tx)
	orgStore := stores.NewDbOrganizationStore(tx)

	org, err := orgStore.FindByUuid(orgUuid)
	if err != nil {
		return err
	}

	if allowed, err := ctxt.Auth().CanRead(org); !allowed {
		return err
	}

	orgProjects, err := projectStore.FindAllByOrganizationUuid(orgUuid)
	if err != nil {
		return err
	}

	var interfaceProjects []interface{} = make([]interface{}, 0, len(orgProjects))
	for _, p := range orgProjects {
		if allowed, _ := ctxt.Auth().CanRead(p); allowed {
			interfaceProjects = append(interfaceProjects, p)
		}
	}

	writeCollectionPageAsJson(ctxt, &CollectionPage{
		Total:      len(interfaceProjects),
		Count:      len(interfaceProjects),
		Collection: interfaceProjects,
	})

	return nil
}

func (self orgHandler) Memberships(ctxt RequestContext) (err error) {

	orgUuid := ctxt.PathParameter("uuid")

	tx := ctxt.Tx()
	orgStore := stores.NewDbOrganizationStore(tx)
	orgMembershipsStore := stores.NewDbOrganizationMembershipStore(tx)
	org, err := orgStore.FindByUuid(orgUuid)
	if err != nil {
		return err
	}

	if allowed, err := ctxt.Auth().CanRead(org); !allowed {
		return err
	}

	var memberships []*domain.OrganizationMembership
	orgMemberships, err := orgMembershipsStore.FindAllByOrganizationUuid(orgUuid)
	if err != nil {
		return err
	}

	for _, membership := range orgMemberships {
		memberships = append(memberships, membership)
	}

	var interfaceMemberships []interface{} = make([]interface{}, 0, len(memberships))
	for _, m := range memberships {
		if allowed, _ := ctxt.Auth().CanRead(m); allowed {
			interfaceMemberships = append(interfaceMemberships, m)
		}
	}

	writeCollectionPageAsJson(ctxt, &CollectionPage{
		Total:      len(interfaceMemberships),
		Count:      len(interfaceMemberships),
		Collection: interfaceMemberships,
	})

	return nil
}

func (self orgHandler) Members(ctxt RequestContext) (err error) {

	orgUuid := ctxt.PathParameter("uuid")

	tx := ctxt.Tx()
	orgStore := stores.NewDbOrganizationStore(tx)
	orgMemberStore := stores.NewDbOrganizationMemberStore(tx)
	org, err := orgStore.FindByUuid(orgUuid)
	if err != nil {
		return err
	}

	if allowed, err := ctxt.Auth().CanRead(org); !allowed {
		return err
	}

	members, err := orgMemberStore.FindAllByOrganizationUuid(orgUuid)
	if err != nil {
		return err
	}

	interfaceMembers := make([]interface{}, 0, len(members))
	for _, m := range members {
		if allowed, _ := ctxt.Auth().CanRead(m); allowed {
			interfaceMembers = append(interfaceMembers, m)
		}
	}

	writeCollectionPageAsJson(ctxt, &CollectionPage{
		Total:      len(interfaceMembers),
		Count:      len(interfaceMembers),
		Collection: interfaceMembers,
	})

	return nil
}

func (self orgHandler) Limits(ctxt RequestContext) error {

	if ctxt.User() == nil {
		return ErrLoginRequired
	}

	organizationUuid := ctxt.PathParameter("uuid")

	limits := NewLimitsFromContext(ctxt)
	reported, err := limits.ForOrganizationUuid(organizationUuid)
	if err != nil {
		return err
	}

	writeAsJson(ctxt, reported)

	return nil
}

func (self orgHandler) ProjectCards(ctxt RequestContext) error {

	organizationUuid := ctxt.PathParameter("uuid")
	organization, err := stores.NewDbOrganizationStore(ctxt.Tx()).FindByUuid(organizationUuid)
	if err != nil {
		return err
	}

	if allowed, err := ctxt.Auth().CanRead(organization); !allowed {
		return err
	}

	cards := []*domain.ProjectCard{}
	//
	// Projector currently can't start fast enough to be useful, so comment it out
	//
	// if response, err := http.Get(os.Getenv("HARROW_PROJECTOR_URL") + "/organizations/" + organizationUuid); err == nil {
	// 	cardsByProjectUuid := CardsByProjectUuid{}
	// 	body := &projector.Response{
	// 		Subject: &cardsByProjectUuid,
	// 	}
	//
	// 	if err := json.NewDecoder(response.Body).Decode(body); err != nil {
	// 		ctxt.Log().Error().Msgf("/organizations/%s/project-cards", err)
	// 	}
	//
	// 	if body.Error != "" {
	// 		ctxt.Log().Error().Msgf("/organizations/%s/project-cards", err)
	// 	}
	//
	// 	cards = cardsByProjectUuid.ProjectCards()
	// }
	if len(cards) == 0 {
		cardsFromDB, err := stores.NewDbProjectCardStore(ctxt.Tx()).FindAllByOrganizationUuid(organizationUuid)
		if err != nil {
			return err
		}
		cards = cardsFromDB
	}

	result := []interface{}{}
	for _, card := range cards {
		if allowed, _ := ctxt.Auth().CanRead(card); allowed {
			result = append(result, card)
		}
	}

	writeCollectionPageAsJson(ctxt, &CollectionPage{
		Total:      len(result),
		Count:      len(result),
		Collection: result,
	})

	return nil
}
