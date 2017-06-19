package authz

import (
	"sort"
	"strings"

	"github.com/harrowio/harrow/config"
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/logger"
	"github.com/harrowio/harrow/stores"
	"github.com/jmoiron/sqlx"
)

var (
	BasicCapabilities = func() map[string]bool {
		return map[string]bool{
			"read-public":                             true,
			"create-session":                          true,
			domain.CapabilityValidate + "-session":    true,
			domain.CapabilitySignUp + "-user":         true,
			domain.CapabilityCreate + "-organization": true,
		}
	}
)

type Ownable interface {
	OwnedBy(user *domain.User) bool
}

type BelongsToProject interface {
	FindProject(store domain.ProjectStore) (*domain.Project, error)
}

type BelongsToUser interface {
	FindUser(store domain.UserStore) (*domain.User, error)
}

type BelongsToOrganization interface {
	FindOrganization(store domain.OrganizationStore) (*domain.Organization, error)
}

// A Role is any source of capabilities.
type Role interface {
	// Capabilities returns the list of capabilities possessed by this
	// role.  A capability should take the form of <VERB>-<SUBJECT-NAME>,
	// for example "read-project".
	//
	// Verbs understood by Service are: "create", "read", "update", "archive".
	Capabilities() []string
}

// A Subject is anything against which capabilities can be checked.
type Subject interface {
	// AuthorizationName returns name for this subject to be used
	// in capabilities.  The value returned by this function is used
	// to construct the capability name to check against.
	AuthorizationName() string
}

type Service interface {
	CanRead(thing interface{}) (bool, error)
	CanCreate(thing interface{}) (bool, error)
	CanUpdate(thing interface{}) (bool, error)
	CanArchive(thing interface{}) (bool, error)
	Can(action string, thing interface{}) (bool, error)
	CapabilitiesBySubject() map[string][]string

	Log() logger.Logger
	SetLogger(logger.Logger)
}

var (
	ErrCapabilityMissing      = &Error{reason: "capability_missing"}
	ErrNoAuthorizationDefined = &Error{reason: "no_authorization_defined"}
	ErrBlocked                = &Error{reason: "blocked"}
)

type txService struct {
	tx                *sqlx.Tx
	organizationStore domain.OrganizationStore
	projectStore      domain.ProjectStore
	invitationStore   domain.InvitationStore
	userStore         domain.UserStore
	userBlockStore    UserBlockStore

	organizationMembershipStore *stores.DbOrganizationMembershipStore
	projectMembershipStore      *stores.DbProjectMembershipStore

	capabilities map[string]bool

	currentUser            *domain.User
	user                   *domain.User
	project                *domain.Project
	organization           *domain.Organization
	organizationMembership *domain.OrganizationMembership
	projectMembership      *domain.ProjectMembership

	addedRoles map[string]bool

	cachedOrganizationMemberships map[string]*domain.OrganizationMembership
	cachedProjectMemberships      map[string]*domain.ProjectMembership

	log logger.Logger
}

func NewService(tx *sqlx.Tx, currentUser *domain.User, c *config.Config) *txService {
	service := &txService{
		tx:         tx,
		addedRoles: map[string]bool{},
	}

	service.projectStore = stores.NewCachedProjectStore(
		stores.NewDbProjectStore(tx),
	)
	service.userStore = stores.NewCachedUserStore(
		stores.NewDbUserStore(tx, c),
	)
	service.userBlockStore = stores.NewCachedUserBlockStore(
		stores.NewDbUserBlockStore(tx),
	)
	service.organizationStore = stores.NewCachedOrganizationStore(
		stores.NewDbOrganizationStore(tx),
	)
	service.invitationStore = stores.NewCachedInvitationStore(
		stores.NewDbInvitationStore(tx),
	)

	service.organizationMembershipStore = stores.NewDbOrganizationMembershipStore(tx)
	service.cachedOrganizationMemberships = map[string]*domain.OrganizationMembership{}

	service.projectMembershipStore = stores.NewDbProjectMembershipStore(tx)
	service.cachedProjectMemberships = map[string]*domain.ProjectMembership{}

	service.currentUser = currentUser
	service.capabilities = BasicCapabilities()

	return service
}

func (s *txService) Can(action string, thing interface{}) (bool, error) {
	if err := s.currentUserIsBlocked(); err != nil {
		return false, err
	}

	if !s.HasCapabilities(thing) {
		return false, ErrNoAuthorizationDefined
	}

	s.useBasicCapablities()
	if s.HasAuthorization(thing) {
		err := s.load(thing)
		//s.logLoadedEntities()
		if err != nil {
			return false, err
		}

		s.determineRoles()
	}

	if s.owned(thing) {
		return true, nil
	}

	subject := thing.(Subject)
	capabilityName := action + "-" + subject.AuthorizationName()
	allowed := s.capabilities[capabilityName]
	if allowed {
		return true, nil
	} else {
		return false, NewMissingCapabilityError(s.capabilities, capabilityName)
	}
}

func (s *txService) useBasicCapablities() {
	s.capabilities = BasicCapabilities()
	for capability, _ := range s.addedRoles {
		s.capabilities[capability] = true
	}
}

func (s *txService) currentUserIsBlocked() error {
	if s.currentUser == nil {
		return nil
	}
	blocked, err := s.userBlockStore.UserIsBlocked(s.currentUser.Uuid)
	if err != nil {
		return &Error{internal: err, reason: "internal"}
	}

	if blocked {
		return ErrBlocked
	}

	return nil
}

func (s *txService) HasCapabilities(thing interface{}) bool {
	_, ok := thing.(Subject)
	return ok
}

func (s *txService) HasAuthorization(thing interface{}) bool {
	if _, ok := thing.(Ownable); ok {
		return true
	}
	if _, ok := thing.(BelongsToProject); ok {
		return true
	}
	if _, ok := thing.(BelongsToOrganization); ok {
		return true
	}
	if _, ok := thing.(BelongsToUser); ok {
		return true
	}

	return false
}

func (s *txService) load(thing interface{}) error {

	s.Log().Debug().Msgf("authz: load: thing: %T", thing)

	if err := s.loadForUser(thing); err != nil {
		return err
	}

	if err := s.loadForProject(thing); err != nil {
		return err
	}

	if err := s.loadForOrganization(thing); err != nil {
		return err
	}

	if err := s.loadAssociatedEntities(); err != nil {
		return err
	}

	return nil
}

func (s *txService) loadForUser(thing interface{}) error {
	object, ok := thing.(BelongsToUser)
	if !ok {
		return nil
	}

	s.Log().Debug().Msgf("authz: belongstouser: %t", object)

	user, err := object.FindUser(s.userStore)
	if err != nil {
		if _, ok := err.(*domain.NotFoundError); !ok {
			s.Log().Debug().Msgf("authz: loadforuser: %s", err)
			return &Error{internal: err, reason: "internal"}
		}
	}

	s.user = user

	return nil
}

func (s *txService) loadForProject(thing interface{}) error {
	object, ok := thing.(BelongsToProject)
	if !ok {
		return nil
	}

	s.Log().Debug().Msgf("authz: belongstoproject: %t", object)

	project, err := object.FindProject(s.projectStore)
	if err != nil {
		if _, ok := err.(*domain.NotFoundError); !ok {
			s.Log().Debug().Msgf("authz: loadforproject: %s", err)
			return &Error{internal: err, reason: "internal"}
		}
	}
	s.project = project

	return nil
}

func (s *txService) loadForOrganization(thing interface{}) error {
	object, ok := thing.(BelongsToOrganization)
	if !ok {
		return nil
	}

	s.Log().Debug().Msgf("authz: belongstoorganization: %t", object)

	organization, err := object.FindOrganization(s.organizationStore)
	if err != nil {
		if _, ok := err.(*domain.NotFoundError); !ok {
			s.Log().Debug().Msgf("authz: loadfororganization: %s", err)
			return &Error{internal: err, reason: "internal"}
		}
	}
	s.organization = organization

	return nil
}

func (s *txService) loadAssociatedEntities() error {

	if err := s.loadOrganization(); err != nil {
		return err
	}

	if err := s.loadOrganizationMembership(); err != nil {
		return err
	}

	if err := s.loadProjectMembership(); err != nil {
		return err
	}

	return nil
}

func (s *txService) loadOrganization() error {
	if s.project == nil || s.organization != nil {
		return nil
	}

	organization, err := s.organizationStore.FindByUuid(s.project.OrganizationUuid)
	if err != nil {
		if _, ok := err.(*domain.NotFoundError); !ok {
			return &Error{internal: err, reason: "internal"}
		}
	}

	s.organization = organization

	return nil
}

func (s *txService) loadOrganizationMembership() error {
	if s.organization == nil || s.currentUser == nil {
		return nil
	}

	organizationMembership, found := s.findCachedOrganizationMembership(s.organization.Uuid, s.currentUser.Uuid)
	if found {
		s.organizationMembership = organizationMembership
		return nil
	}

	organizationMembership, err := s.organizationMembershipStore.
		FindByOrganizationAndUserUuids(s.organization.Uuid, s.currentUser.Uuid)
	if err != nil {
		if _, ok := err.(*domain.NotFoundError); !ok {
			return &Error{internal: err, reason: "internal"}
		}
		organizationMembership, err = s.organizationMembershipStore.
			FindThroughProjectMemberships(s.organization.Uuid, s.currentUser.Uuid)
		if err != nil {
			if _, ok := err.(*domain.NotFoundError); !ok {
				return &Error{internal: err, reason: "internal"}
			}
		}
	}

	s.cacheOrganizationMembership(organizationMembership)
	s.organizationMembership = organizationMembership

	return nil
}

func (s *txService) findCachedOrganizationMembership(organizationUuid, userUuid string) (*domain.OrganizationMembership, bool) {
	membership, found := s.cachedOrganizationMemberships[organizationUuid+":"+userUuid]
	return membership, found
}

func (s *txService) cacheOrganizationMembership(membership *domain.OrganizationMembership) {
	if membership == nil {
		return
	}

	key := membership.OrganizationUuid + ":" + membership.UserUuid
	s.cachedOrganizationMemberships[key] = membership
}

func (s *txService) loadProjectMembership() error {
	if s.project == nil || s.currentUser == nil {
		return nil
	}

	projectMembership, found := s.findCachedProjectMembership(s.project.Uuid, s.currentUser.Uuid)
	if found {
		s.projectMembership = projectMembership
		return nil
	}

	projectMembership, err := s.projectMembershipStore.FindByUserAndProjectUuid(s.currentUser.Uuid, s.project.Uuid)
	if err != nil {
		if _, ok := err.(*domain.NotFoundError); !ok {
			return &Error{internal: err, reason: "internal"}
		}
	}

	s.cacheProjectMembership(projectMembership)
	s.projectMembership = projectMembership

	return nil
}

func (s *txService) findCachedProjectMembership(projectUuid, userUuid string) (*domain.ProjectMembership, bool) {
	membership, found := s.cachedProjectMemberships[projectUuid+":"+userUuid]
	return membership, found
}

func (s *txService) cacheProjectMembership(membership *domain.ProjectMembership) {
	if membership == nil {
		return
	}

	key := membership.ProjectUuid + ":" + membership.UserUuid
	s.cachedProjectMemberships[key] = membership
}

func (s *txService) logLoadedEntities() {
	s.Log().Debug().Msgf("authz: organization %#v", s.organization)
	s.Log().Debug().Msgf("authz: project %#v", s.project)
	s.Log().Debug().Msgf("authz: projectmembership %#v", s.projectMembership)
	s.Log().Debug().Msgf("authz: organizationmembership %#v", s.organizationMembership)
	if s.user != nil {
		s.Log().Debug().Msgf("authz: user uuid %s", s.user.Uuid)
	}
	if s.currentUser != nil {
		s.Log().Debug().Msgf("authz: current user uuid %s", s.currentUser.Uuid)
	}
}

func (s *txService) determineRoles() {
	if s.currentUser != nil && s.project != nil {
		projectMember := domain.NewProjectMember(s.currentUser, s.project, s.projectMembership, s.organizationMembership)
		if projectMember != nil {
			s.addSource(projectMember)
		}

		invited, err := s.currentUser.InvitedTo(s.project, s.invitationStore)
		if err == nil {
			if invited {
				s.addSource(domain.ProjectInvitee)
			}
		} else {
			s.Log().Debug().Msgf("user(%q).invitedto(%q): %s", s.currentUser.Uuid, s.project.Uuid, err)
		}
	} else if s.project != nil {
		if s.project.Public {
			s.addSource(domain.ProjectVisitor)
		}
	}

	if s.organizationMembership != nil {
		organizationMember := domain.NewOrganizationMember(s.currentUser, s.organizationMembership)
		if organizationMember != nil {
			s.addSource(organizationMember)
		}
	}
}

// AddRole adds the capabilities of the given role to all the capabilities
// the user already possesses.  Use this method to endow the user with
// additional capabilities that cannot be determined just by looking at
// the subject (e.g. logging in as an administrator).
func (s *txService) AddRole(role Role) {
	for _, cap := range role.Capabilities() {
		s.addedRoles[cap] = true
	}
}

func (s *txService) addSource(role Role) {
	for _, cap := range role.Capabilities() {
		s.capabilities[cap] = true
	}
}

func (s *txService) owned(thing interface{}) bool {
	owned, ok := thing.(Ownable)
	if ok && s.currentUser != nil && owned.OwnedBy(s.currentUser) {
		return true
	}

	if s.currentUser != nil && s.user != nil {
		return s.currentUser.Uuid == s.user.Uuid
	}

	return false
}

func (s *txService) CanRead(thing interface{}) (bool, error) {
	return s.Can("read", thing)
}

func (s *txService) CanCreate(thing interface{}) (bool, error) {
	return s.Can("create", thing)
}

func (s *txService) CanArchive(thing interface{}) (bool, error) {
	return s.Can("archive", thing)
}

func (s *txService) CanUpdate(thing interface{}) (bool, error) {
	return s.Can("update", thing)
}

func (self *txService) CapabilitiesBySubject() map[string][]string {
	result := map[string][]string{}
	for cap, _ := range self.capabilities {
		sep := strings.Index(cap, "-")
		verb := cap[0:sep]
		subject := cap[sep+1:]
		result[subject] = append(result[subject], verb)
		sort.Strings(result[subject])
	}

	return result
}

func (self *txService) Log() logger.Logger {
	if self.log == nil {
		self.log = logger.Discard
	}
	return self.log
}

func (self *txService) SetLogger(l logger.Logger) {
	self.log = l
}
