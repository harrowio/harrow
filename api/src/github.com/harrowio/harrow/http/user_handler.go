package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"github.com/harrowio/harrow/activities"
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/stores"
	"github.com/harrowio/harrow/uuidhelper"
)

func ReadPatchUserParams(r io.Reader) (*patchUserParams, error) {
	decoder := json.NewDecoder(r)
	var p patchUserParams
	err := decoder.Decode(&p)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func ReadCreateUserParams(r io.Reader) (*createUserParams, error) {
	decoder := json.NewDecoder(r)
	var w createUserParamsWrapper
	err := decoder.Decode(&w)
	if err != nil {
		return nil, err
	}
	w.Subject.Email = strings.TrimSpace(w.Subject.Email)

	return &w.Subject, nil
}

type createUserParamsWrapper struct {
	Subject createUserParams
}

type createUserParams struct {
	Email          string            `json:"email"`
	Name           string            `json:"name"`
	Password       string            `json:"password"`
	InvitationUuid string            `json:"invitationUuid"`
	Parameters     domain.Dictionary `json:"signupParameters"`
}

type updateUserParams struct {
	Uuid        string `json:"uuid"`
	Email       string `json:"email"`
	Name        string `json:"name"`
	Password    string `json:"password"`
	NewPassword string `json:"newPassword"`
}

func (self *updateUserParams) Validate() error {
	self.Email = strings.TrimSpace(self.Email)
	if n := len(self.NewPassword); n > 0 && n < 10 {
		return domain.NewValidationError("password", "too_short")
	}

	return nil
}

type patchUserParams struct {
	TwoFactorAuthEnabled bool   `json:"twoFactorAuthEnabled"`
	TotpGenerateSecret   *bool  `json:"totpGenerateSecret"`
	TotpToken            *int32 `json:"totpToken"`
}

func (self *patchUserParams) IsEnableTotp() bool {
	return self.TwoFactorAuthEnabled && self.TotpToken != nil
}

func (self *patchUserParams) IsDisableTotp() bool {
	return !self.TwoFactorAuthEnabled && self.TotpToken != nil
}

func (self *patchUserParams) IsGenerateTotpSecret() bool {
	return self.TotpGenerateSecret != nil && *self.TotpGenerateSecret
}

func MountUserHandler(r *mux.Router, ctxt ServerContext) {

	uh := &userHandler{}

	// Collection
	root := r.PathPrefix("/users").Subrouter()
	root.Methods("PUT").Handler(HandlerFunc(ctxt, uh.Update)).
		Name("user-update")

	// Relationships
	related := root.PathPrefix("/{uuid}/").Subrouter()
	related.Methods("GET").Path("/oauth-tokens").Handler(HandlerFunc(ctxt, uh.oAuthTokens)).
		Name("user-oauth-tokens")
	related.Methods("GET").Path("/organizations").Handler(HandlerFunc(ctxt, uh.Organizations)).
		Name("user-organizations")
	related.Methods("GET").Path("/activities").Handler(HandlerFunc(ctxt, uh.Activities))
	related.Methods("GET").Path("/blocks").Handler(HandlerFunc(ctxt, uh.Blocks)).
		Name("user-blocks")
	related.Methods("GET").Path("/sessions").Handler(HandlerFunc(ctxt, uh.Sessions)).
		Name("user-sessions")
	related.Methods("GET").Path("/projects").Handler(HandlerFunc(ctxt, uh.Projects)).
		Name("user-projects")
	related.Methods("GET").Path("/jobs").Handler(HandlerFunc(ctxt, uh.Jobs)).
		Name("user-jobs")
	related.Methods("POST").Path("/verify-email").Handler(HandlerFunc(ctxt, uh.VerifyEmail)).
		Name("user-verify-email")
	related.Methods("PATCH").Path("/mfa").Handler(HandlerFunc(ctxt, uh.ChangeMFA)).
		Name("user-change-mfa")

	// Item
	item := root.Path("/{uuid}").Subrouter()
	item.Methods("GET").Handler(HandlerFunc(ctxt, uh.Show)).
		Name("user-show")

	root.Methods("POST").Handler(HandlerFunc(ctxt, uh.Create)).
		Name("user-create")
}

type userHandler struct {
}

func (self userHandler) Create(ctxt RequestContext) (err error) {

	c := ctxt.Config()
	userStore := stores.NewDbUserStore(ctxt.Tx(), &c)

	params, err := ReadCreateUserParams(ctxt.R().Body)
	if err != nil {
		return err
	}

	if len(params.Parameters) >= 10 {
		return domain.NewValidationError("parameters", "too_many_values")
	}

	userId := ""
	if uuidhelper.IsValid(params.InvitationUuid) {
		invitationStore := stores.NewDbInvitationStore(ctxt.Tx())
		if invitation, err := invitationStore.FindByUuid(params.InvitationUuid); err == nil {
			userId = invitation.InviteeUuid
		} else {
			ctxt.Log().Info().Msgf("sign-up user for invitation: %s", err)
		}
	}

	user := &domain.User{
		Uuid:             userId,
		Name:             params.Name,
		Email:            params.Email,
		Password:         params.Password,
		UrlHost:          ctxt.R().Host,
		SignupParameters: params.Parameters,
	}

	if allowed, err := ctxt.Auth().Can(domain.CapabilitySignUp, user); !allowed {
		return err
	}

	userUuid, err := userStore.Create(user)
	if err != nil {
		return err
	}

	blockValidFrom := domain.Clock.Now().Add(12 * time.Hour)
	block, err := user.NewBlock("email_unverified")
	if err != nil {
		return NewInternalError(err)
	}

	block.BlockForever(blockValidFrom)
	if err := stores.NewDbUserBlockStore(ctxt.Tx()).Create(block); err != nil {
		return err
	}

	ctxt.EnqueueActivity(activities.UserSignedUp(user), &userUuid)

	for parameter, value := range params.Parameters {
		ctxt.EnqueueActivity(activities.UserSignupParameterSet(user, parameter, value), &user.Uuid)
	}

	user, err = userStore.FindByUuid(userUuid)
	if err != nil {
		return err
	}

	ctxt.W().Header().Add("Location", urlForSubject(ctxt.R(), user))
	ctxt.W().WriteHeader(http.StatusCreated)

	writeAsJson(ctxt, user)

	return err

}

func (self userHandler) Update(ctxt RequestContext) error {

	params := &updateUserParams{}
	dec := json.NewDecoder(ctxt.R().Body)
	if err := dec.Decode(&halWrapper{Subject: params}); err != nil {
		return err
	}

	if err := params.Validate(); err != nil {
		return err
	}

	cfg := ctxt.Config()
	userStore := stores.NewDbUserStore(ctxt.Tx(), &cfg)
	user, err := userStore.FindByUuid(params.Uuid)
	if err != nil {
		return err
	}

	// Don't check for password when the user has no password (oauth account)
	if _, err = userStore.FindUuidByEmailAddressAndPassword(user.Email, params.Password); !user.WithoutPassword && err != nil {
		if _, ok := err.(*domain.NotFoundError); ok {
			return domain.NewValidationError("password", "invalid")
		}
		return err
	}

	if params.NewPassword != "" {
		user.Password = params.NewPassword
	}
	user.Email = params.Email
	user.Name = params.Name
	if allowed, err := ctxt.Auth().CanUpdate(user); !allowed {
		return err
	}

	if err := userStore.Update(user); err != nil {
		return err
	}

	writeAsJson(ctxt, user)
	return nil
}

func (self userHandler) Show(ctxt RequestContext) (err error) {

	userUuid := ctxt.PathParameter("uuid")

	c := ctxt.Config()
	userStore := stores.NewDbUserStore(ctxt.Tx(), &c)

	user, err := userStore.FindByUuid(userUuid)
	if err != nil {
		return err
	}

	if allowed, err := ctxt.Auth().CanRead(user); !allowed {
		return err
	}

	writeAsJson(ctxt, user)

	return err

}

func (self userHandler) ChangeMFA(ctxt RequestContext) (err error) {

	c := ctxt.Config()
	userUuid := ctxt.PathParameter("uuid")
	userStore := stores.NewDbUserStore(ctxt.Tx(), &c)

	user, err := userStore.FindByUuid(userUuid)
	if err != nil {
		return err
	}

	if allowed, err := ctxt.Auth().CanUpdate(user); !allowed {
		return err
	}

	params, err := ReadPatchUserParams(ctxt.R().Body)
	if err != nil {
		return err
	}

	if params.IsEnableTotp() {
		totp := *params.TotpToken
		if len(fmt.Sprintf("%0d", totp)) != 6 {
			return domain.NewValidationError("totp", "invalid_totp_token")
		}

		err = user.EnableTotp(totp)
		if err != nil {
			return err
		}
		err = userStore.Update(user)
		if err != nil {
			return err
		}

	} else if params.IsGenerateTotpSecret() {
		user.GenerateTotpSecret()
		err = userStore.Update(user)
		if err != nil {
			return err
		}

		user, err = userStore.FindByUuid(user.Uuid)
		if err != nil {
			return err
		}
	} else if params.IsDisableTotp() {
		err = user.DisableTotp(*params.TotpToken)
		if err != nil {
			return err
		}

		err = userStore.Update(user)
		if err != nil {
			return err
		}
	} else {
		return domain.NewValidationError("totpToken", "required")
	}

	view := struct {
		*domain.User
		TotpSecret string `json:"totpSecret,omitempty"`
	}{user, user.TotpSecret}

	writeAsJson(ctxt, view)

	return err
}

func (self userHandler) Organizations(ctxt RequestContext) (err error) {

	userUuid := ctxt.PathParameter("uuid")

	orgStore := stores.NewDbOrganizationStore(ctxt.Tx())

	c := ctxt.Config()
	userStore := stores.NewDbUserStore(ctxt.Tx(), &c)

	user, err := userStore.FindByUuid(userUuid)
	if err != nil {
		return err
	}

	if allowed, err := ctxt.Auth().CanRead(user); !allowed {
		return err
	}

	orgs, err := orgStore.FindAllByUserUuidThroughMemberships(userUuid)
	if err != nil {
		return err
	}
	billingHistory, err := stores.NewDbBillingHistoryStore(
		ctxt.Tx(),
		ctxt.KeyValueStore(),
	).Load()
	if err != nil {
		return err
	}

	limitsStore := stores.NewDbLimitsStore(ctxt.Tx())
	limitsStore.SetLogger(ctxt.Log())

	embedLimits := func(o *domain.Organization) {
		limits, err := limitsStore.FindByOrganizationUuid(o.Uuid)
		if err != nil {
			ctxt.Log().Warn().Msgf("user.Organizations.FindLimits: %q: %s", o.Uuid, err)
			return
		}

		planUuid := billingHistory.PlanUuidFor(o.Uuid)
		plan, err := stores.NewDbBillingPlanStore(ctxt.Tx(), stores.NewBraintreeProxy()).FindByUuid(planUuid)
		if err != nil {
			ctxt.Log().Warn().Msgf("error loading plan for organization %s: %s", o.Uuid, err)
			return
		}

		reported := limits.Report(
			plan,
			billingHistory.ExtraUsersFor(o.Uuid),
			billingHistory.ExtraProjectsFor(o.Uuid),
		)
		ctxt.Log().Debug().Msgf("embedding limits for %q: %#v", o.Uuid, reported)
		o.Embed("limits", reported)
	}

	var interfaceOrgs []interface{} = make([]interface{}, 0, len(orgs))
	for _, o := range orgs {
		if allowed, _ := ctxt.Auth().CanRead(o); allowed {
			embedLimits(o)
			interfaceOrgs = append(interfaceOrgs, o)
		}
	}

	writeCollectionPageAsJson(ctxt, &CollectionPage{
		Total:      len(interfaceOrgs),
		Count:      len(interfaceOrgs),
		Collection: interfaceOrgs,
	})

	return err

}

func (self userHandler) Sessions(ctxt RequestContext) (err error) {

	userUuid := ctxt.PathParameter("uuid")

	c := ctxt.Config()
	userStore := stores.NewDbUserStore(ctxt.Tx(), &c)

	user, err := userStore.FindByUuid(userUuid)
	if err != nil {
		return err
	}

	sessionStore := stores.NewDbSessionStore(ctxt.Tx())

	if allowed, err := ctxt.Auth().CanRead(user); !allowed {
		return err
	}

	sessions, err := sessionStore.FindAllByUserUuid(userUuid)
	if err != nil {
		return err
	}

	var interfaceSessions []interface{} = make([]interface{}, 0, len(sessions))
	for _, s := range sessions {
		if allowed, _ := ctxt.Auth().CanRead(s); allowed {
			interfaceSessions = append(interfaceSessions, s)
		}
	}

	writeCollectionPageAsJson(ctxt, &CollectionPage{
		Total:      len(interfaceSessions),
		Count:      len(interfaceSessions),
		Collection: interfaceSessions,
	})

	return err

}

func (self userHandler) oAuthTokens(ctxt RequestContext) (err error) {

	userUuid := ctxt.PathParameter("uuid")

	c := ctxt.Config()
	userStore := stores.NewDbUserStore(ctxt.Tx(), &c)
	tokenStore := stores.NewDbOAuthTokenStore(ctxt.Tx())

	user, err := userStore.FindByUuid(userUuid)
	if err != nil {
		return err
	}

	if allowed, err := ctxt.Auth().CanRead(user); !allowed {
		return err
	}

	tokens, err := tokenStore.FindAllByUserUuid(userUuid)
	if err != nil {
		return err
	}

	var interfaceTokens []interface{} = make([]interface{}, 0, len(tokens))
	for _, t := range tokens {
		if allowed, _ := ctxt.Auth().CanRead(t); allowed {
			interfaceTokens = append(interfaceTokens, t)
		}
	}

	writeCollectionPageAsJson(ctxt, &CollectionPage{
		Total:      len(interfaceTokens),
		Count:      len(interfaceTokens),
		Collection: interfaceTokens,
	})

	return err
}

func (self userHandler) Projects(ctxt RequestContext) (err error) {

	userUuid := ctxt.PathParameter("uuid")

	c := ctxt.Config()
	userStore := stores.NewDbUserStore(ctxt.Tx(), &c)
	projectStore := stores.NewDbProjectStore(ctxt.Tx())

	user, err := userStore.FindByUuid(userUuid)
	if err != nil {
		return err
	}

	if allowed, err := ctxt.Auth().CanRead(user); !allowed {
		return err
	}

	membershipOnly := ctxt.R().FormValue("membershipOnly")
	projects := []*domain.Project{}
	if membershipOnly != "" {
		projects, err = projectStore.FindAllByUserUuidOnlyThroughProjectMemberships(userUuid)
	} else {
		projects, err = projectStore.FindAllByUserUuid(userUuid)
	}
	if err != nil {
		return err
	}
	orgStore := stores.NewDbOrganizationStore(ctxt.Tx())

	var interfaceProjects []interface{} = make([]interface{}, 0, len(projects))
	for _, p := range projects {
		if allowed, _ := ctxt.Auth().CanRead(p); allowed {
			org, err := orgStore.FindByUuid(p.OrganizationUuid)
			if err == nil {
				p.Embed("organizations", org)
			}
			interfaceProjects = append(interfaceProjects, p)
		}
	}

	writeCollectionPageAsJson(ctxt, &CollectionPage{
		Total:      len(interfaceProjects),
		Count:      len(interfaceProjects),
		Collection: interfaceProjects,
	})

	return err
}

func (self userHandler) Jobs(ctxt RequestContext) (err error) {

	userUuid := ctxt.PathParameter("uuid")

	c := ctxt.Config()
	userStore := stores.NewDbUserStore(ctxt.Tx(), &c)
	projectStore := stores.NewDbProjectStore(ctxt.Tx())
	operationStore := stores.NewDbOperationStore(ctxt.Tx())
	jobStore := stores.NewDbJobStore(ctxt.Tx())

	user, err := userStore.FindByUuid(userUuid)
	if err != nil {
		return err
	}

	if allowed, err := ctxt.Auth().CanRead(user); !allowed {
		return err
	}

	projects, err := projectStore.FindAllByUserUuid(userUuid)
	if err != nil {
		return err
	}

	uuids := []string{}
	for _, project := range projects {
		uuids = append(uuids, project.Uuid)
	}

	jobs, err := jobStore.FindAllByProjectUuids(uuids)
	if err != nil {
		return err
	}

	var interfaceJobs []interface{} = make([]interface{}, 0, len(jobs))
	for _, job := range jobs {
		if strings.HasPrefix(job.Name, "urn:harrow:default-job") {
			continue
		}

		if allowed, _ := ctxt.Auth().CanRead(job); allowed {
			if err := job.FindRecentOperations(operationStore); err != nil {
				return err
			}

			interfaceJobs = append(interfaceJobs, job)
		}
	}

	writeCollectionPageAsJson(ctxt, &CollectionPage{
		Total:      len(interfaceJobs),
		Count:      len(interfaceJobs),
		Collection: interfaceJobs,
	})

	return nil
}

func (self userHandler) Blocks(ctxt RequestContext) error {

	userUuid := ctxt.PathParameter("uuid")

	userBlockStore := stores.NewDbUserBlockStore(ctxt.Tx())
	blocks, err := userBlockStore.FindAllByUserUuid(userUuid)
	if err != nil {
		return err
	}

	interfaceBlocks := []interface{}{}
	for _, block := range blocks {
		interfaceBlocks = append(interfaceBlocks, block)
	}

	writeCollectionPageAsJson(ctxt, &CollectionPage{
		Total:      len(interfaceBlocks),
		Count:      len(interfaceBlocks),
		Collection: interfaceBlocks,
	})

	return nil
}

func (self userHandler) Activities(ctxt RequestContext) error {

	userUuid := ctxt.PathParameter("uuid")

	streams := stores.NewKVActivityStreamStore(
		ctxt.KeyValueStore(),
	)
	streams.SetLogger(ctxt.Log())

	if allowed, err := ctxt.Auth().CanRead(ctxt.User()); !allowed {
		return err
	}

	activities, err := streams.FindStreamForUser(userUuid)
	if err != nil {
		return err
	}

	interfaceActivities := make([]interface{}, len(activities))
	for i, activity := range activities {
		interfaceActivities[i] = activity
	}

	writeCollectionPageAsJson(ctxt, &CollectionPage{
		Count:      len(interfaceActivities),
		Total:      len(interfaceActivities),
		Collection: interfaceActivities,
	})

	return nil
}

func (self userHandler) VerifyEmail(ctxt RequestContext) error {

	userUuid := ctxt.PathParameter("uuid")
	params := struct {
		Token string `json:"token"`
	}{}
	if err := json.NewDecoder(ctxt.R().Body).Decode(&params); err != nil {
		return err
	}

	c := ctxt.Config()
	user, err := stores.NewDbUserStore(ctxt.Tx(), &c).FindByUuid(userUuid)
	if err != nil {
		return err
	}

	if user.Token != params.Token {
		return domain.NewValidationError("token", "invalid")
	}

	blocks := stores.NewDbUserBlockStore(ctxt.Tx())
	if err := blocks.DeleteByUserUuidAndReason(userUuid, "email_unverified"); err != nil {
		return err
	}

	ctxt.EnqueueActivity(activities.UserEmailVerified(user), &userUuid)

	return nil
}
