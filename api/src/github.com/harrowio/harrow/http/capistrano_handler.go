package http

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"regexp"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"github.com/harrowio/harrow/activities"
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/domain/stencil"
	"github.com/harrowio/harrow/git"
	"github.com/harrowio/harrow/stores"
	"github.com/harrowio/harrow/uuidhelper"
)

type CapistranoSignUpParameters struct {
	Name          string `json:"name"`
	Password      string `json:"password"`
	Email         string `json:"email"`
	RepositoryURL string `json:"repository_url"`
}

func (params *CapistranoSignUpParameters) OrganizationName() string {
	if params.RepositoryURL == "" {
		return strings.Replace(
			strings.ToLower(params.Name),
			" ",
			"-",
			-1,
		)
	}

	return regexp.MustCompile(".*[:/]+(?P<organization>[^/]+)/([^/]+)$").
		ReplaceAllString(params.RepositoryURL, "${organization}")
}

func (params *CapistranoSignUpParameters) ProjectName() string {
	if params.RepositoryURL == "" {
		return strings.Replace(
			strings.ToLower(params.Name),
			" ",
			"-",
			-1,
		)
	}

	return regexp.MustCompile(".*[:/](?P<organization>[^/]+)/(?P<project>[^/]+)$").
		ReplaceAllString(params.RepositoryURL, "${project}")
}

type CapistranoSignUpResponse struct {
	OrganizationUuid string `json:"organization_uuid"`
	ProjectUuid      string `json:"project_uuid"`
	SessionUuid      string `json:"session_uuid"`
	OrganizationName string `json:"organization_name"`
	ProjectName      string `json:"project_name"`
}

type capistranoHandler struct {
	users *stores.DbUserStore

	user         *domain.User
	organization *domain.Organization
	project      *domain.Project
	sessionUuid  string
}

func (h *capistranoHandler) init(ctxt RequestContext) (*capistranoHandler, error) {
	handler := &capistranoHandler{}
	handler.users = stores.NewDbUserStore(ctxt.Tx(), c)
	return handler, nil
}

func (h *capistranoHandler) subjectFrom(src io.Reader) (*CapistranoSignUpParameters, error) {
	dec := json.NewDecoder(src)
	params := &CapistranoSignUpParameters{}

	if err := dec.Decode(params); err != nil {
		return nil, err
	}

	return params, nil
}

func MountCapistranoHandler(r *mux.Router, ctxt ServerContext) {
	h := &capistranoHandler{}

	r.Methods("POST").Path("/capistrano/sign-up").Handler(HandlerFunc(ctxt, h.SignUp)).
		Name("capistrano-sign-up")

}

func (self *capistranoHandler) SignUp(ctxt RequestContext) error {
	h, err := self.init(ctxt)
	if err != nil {
		return err
	}

	params, err := h.subjectFrom(ctxt.R().Body)
	if err != nil {
		return err
	}

	if err := h.signUpUser(params, ctxt); err != nil {
		return err
	}

	if err := h.createOrganization(params, ctxt); err != nil {
		return err
	}

	if err := h.createProject(params, ctxt); err != nil {
		return err
	}

	if err := h.createRepository(params, ctxt); err != nil {
		return err
	}

	ctxt.EnqueueActivity(activities.UserSignedUpViaCapistrano(h.user), &h.user.Uuid)

	response := &CapistranoSignUpResponse{
		OrganizationUuid: h.organization.Uuid,
		OrganizationName: h.organization.Name,
		SessionUuid:      h.sessionUuid,
		ProjectUuid:      h.project.Uuid,
		ProjectName:      h.project.Name,
	}

	ctxt.W().Header().Set("Content-Type", "application/json")

	return json.NewEncoder(ctxt.W()).Encode(response)
}

func (self *capistranoHandler) createRepository(params *CapistranoSignUpParameters, ctxt RequestContext) error {
	if params.RepositoryURL == "" {
		return nil
	}

	repository := &domain.Repository{
		Url:         params.RepositoryURL,
		Name:        params.RepositoryURL,
		ProjectUuid: self.project.Uuid,
	}
	store := stores.NewDbRepositoryStore(ctxt.Tx())
	gitRepo, err := git.NewRepository(repository.Url)
	if err != nil {
		return domain.NewValidationError("url", err.Error())
	}
	userInfo := gitRepo.URL.User
	if gitRepo.UsesHTTP() {
		gitRepo.URL.User = nil
		repository.Url = gitRepo.URL.String()
	}

	repositoryUuid, err := store.Create(repository)
	if err != nil {
		return err
	}
	ctxt.EnqueueActivity(activities.RepositoryAdded(repository), &self.user.Uuid)

	if _, err := makeSshRepositoryCredential(ctxt, repositoryUuid); err != nil {
		return err
	}

	if userInfo != nil && gitRepo.UsesHTTP() {
		toSave := (*domain.BasicRepositoryCredential)(nil)
		repositoryCredentials := stores.NewRepositoryCredentialStore(ctxt.SecretKeyValueStore(), ctxt.Tx())
		existingCredential, err := repositoryCredentials.FindByRepositoryUuidAndType(repository.Uuid, domain.RepositoryCredentialBasic)
		if err != nil && !domain.IsNotFound(err) {
			return err
		}

		password, _ := userInfo.Password()
		if existingCredential != nil {
			toSave = &domain.BasicRepositoryCredential{
				RepositoryCredential: existingCredential,
				Username:             userInfo.Username(),
				Password:             password,
			}
		} else {
			toSave = &domain.BasicRepositoryCredential{
				RepositoryCredential: &domain.RepositoryCredential{
					Uuid:           uuidhelper.MustNewV4(),
					Name:           "HTTP Access",
					RepositoryUuid: repository.Uuid,
					Type:           domain.RepositoryCredentialBasic,
					Status:         domain.RepositoryCredentialPresent,
				},
				Username: userInfo.Username(),
				Password: password,
			}
		}

		credential, err := toSave.AsRepositoryCredential()
		if err != nil {
			return err
		}
		if existingCredential == nil {
			if _, err := repositoryCredentials.Create(credential); err != nil {
				return err
			}
		} else {
			if err := repositoryCredentials.Update(credential); err != nil {
				return err
			}
		}
	}

	if gitRepo.IsAccessible() {
		ctxt.EnqueueActivity(activities.RepositoryDetectedAsPublic(repository), &self.user.Uuid)
	} else {
		ctxt.EnqueueActivity(activities.RepositoryDetectedAsPrivate(repository), &self.user.Uuid)
	}

	return nil
}

func (self *capistranoHandler) createProject(params *CapistranoSignUpParameters, ctxt RequestContext) error {
	project := &domain.Project{
		Name:             params.ProjectName(),
		Public:           false,
		OrganizationUuid: self.organization.Uuid,
	}

	if _, err := stores.NewDbProjectStore(ctxt.Tx()).Create(project); err != nil {
		return err
	}

	stencils := stores.NewDbStencilStore(ctxt.SecretKeyValueStore(), ctxt.Tx(), ctxt)
	configuration := stencils.ToConfiguration()
	configuration.NotifyViaEmail = params.Email
	configuration.UserUuid = self.user.Uuid
	configuration.UrlHost = self.user.UrlHost
	configuration.ProjectUuid = project.Uuid
	if err := stencil.NewProjectDefaults(configuration).Create(); err != nil {
		return err
	}

	if err := stencil.NewCapistranoRails(configuration).Create(); err != nil {
		return err
	}

	self.project = project
	ctxt.EnqueueActivity(activities.ProjectCreated(project), &self.user.Uuid)

	return nil
}

func (self *capistranoHandler) createOrganization(params *CapistranoSignUpParameters, ctxt RequestContext) error {
	organization := &domain.Organization{
		Name:   params.OrganizationName(),
		Public: false,
	}

	if _, err := stores.NewDbOrganizationStore(ctxt.Tx()).Create(organization); err != nil {
		return err
	}
	orgMembership := &domain.OrganizationMembership{
		OrganizationUuid: organization.Uuid,
		UserUuid:         self.user.Uuid,
		Type:             domain.MembershipTypeOwner,
	}
	if err := stores.NewDbOrganizationMembershipStore(ctxt.Tx()).Create(orgMembership); err != nil {
		return err
	}

	subscriptionId := fmt.Sprintf("free:%s", organization.Uuid)
	events := stores.NewDbBillingEventStore(ctxt.Tx())
	eventData := &domain.BillingPlanSelected{
		UserUuid:       self.user.Uuid,
		PlanUuid:       domain.FreePlanUuid,
		SubscriptionId: subscriptionId,
	}
	eventData.FillFromPlan(domain.FreePlan)
	planSelected := organization.NewBillingEvent(eventData)
	if _, err := events.Create(planSelected); err != nil {
		ctxt.Log().Warn().Msgf("error saving billing event %s (event data %v)", err, planSelected)
	}

	ctxt.EnqueueActivity(activities.OrganizationCreated(organization, domain.FreePlan), &self.user.Uuid)

	self.organization = organization

	return nil
}

func (self *capistranoHandler) signUpUser(params *CapistranoSignUpParameters, ctxt RequestContext) error {
	existingUser, err := self.users.FindByEmailAddress(params.Email)
	if existingUser != nil {
		return domain.NewValidationError("email", "not_unique")
	} else if !domain.IsNotFound(err) {
		return err
	}

	self.user = &domain.User{
		Name:     params.Name,
		Email:    params.Email,
		Password: params.Password,
		UrlHost:  ctxt.R().Host,
	}

	if err := domain.ValidateUser(self.user); err != nil {
		return err
	}

	if _, err := self.users.Create(self.user); err != nil {
		return err
	}
	ctxt.EnqueueActivity(activities.UserSignedUp(self.user), &self.user.Uuid)
	blockValidFrom := domain.Clock.Now().Add(12 * time.Hour)
	block, err := self.user.NewBlock("email_unverified")
	if err != nil {
		return NewInternalError(err)
	}

	block.BlockForever(blockValidFrom)
	if err := stores.NewDbUserBlockStore(ctxt.Tx()).Create(block); err != nil {
		return err
	}

	remoteHost, _, err := net.SplitHostPort(ctxt.R().RemoteAddr)
	if err != nil {
		return err
	}

	session := &domain.Session{
		UserUuid:      self.user.Uuid,
		UserAgent:     ctxt.R().UserAgent(),
		ClientAddress: remoteHost,
	}

	sessionUuid, err := stores.NewDbSessionStore(ctxt.Tx()).Create(session)
	if err != nil {
		return err
	}

	self.sessionUuid = sessionUuid

	ctxt.EnqueueActivity(activities.UserLoggedIn(self.user), &self.user.Uuid)
	ctxt.EnqueueActivity(activities.UserSignupParameterSet(self.user, "utm_campaign", "CAM_ITM_V3500_00018"), &self.user.Uuid)
	return nil
}
