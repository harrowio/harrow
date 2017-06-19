package domain

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"strings"
	"time"

	"github.com/harrowio/harrow/uuidhelper"
)

var (
	userCapabilities = newCapabilityList().
				reads("session", "invitation").
				writesFor("invitation").
				strings()

	ProjectInvitee = newCapabilityList().
			reads("project", "organization", "user")

	ErrTotpTokenNotValid = NewValidationError("totp", "not valid anymore")
)

// User is a basic type representing users in the system, the JSON marshalling
// of this type is intentionally disabled for private and public members
// as they should never be marshalled directly, everything should go through
// presentation.User marshaller.
type User struct {
	defaultSubject
	Uuid string `json:"uuid"`

	Email      string  `json:"email"`
	GhUsername *string `json:"ghUsername" db:"gh_username"`
	Name       string  `json:"name"`
	Password   string  `json:"password,omitempty"`
	Token      string  `json:"-" db:"token"`
	// If true, the user has no password. They can change user details without
	// providing a password (Name, Email) and cannot log in with username+password
	// This is true for example for users created as result of a OAuth login in
	WithoutPassword bool       `json:"withoutPassword"              db:"without_password"`
	CreatedAt       time.Time  `json:"createdAt"                    db:"created_at"`
	TotpSecret      string     `json:"-"                            db:"totp_secret"`
	TotpEnabledAt   *time.Time `json:"totpEnabledAt"                db:"totp_enabled_at"`
	PasswordHash    string     `json:"-"                            db:"password_hash"`
	// TODO(paul): unused, remove?
	PasswordResetToken string `json:"passwordResetToken,omitempty" db:"password_reset_token"`
	UrlHost            string `json:"urlHost"                      db:"url_host"`

	SignupParameters Dictionary `json:"signupParameters"  db:"signup_parameters"`

	totpToken TotpToken
}

func ValidateUser(u *User) error {

	result := EmptyValidationError()
	if len(u.Email) == 0 {
		result.Add("email", "too_short")
	}

	if len(u.Uuid) == 0 && len(u.Password) < 10 && !u.WithoutPassword {
		result.Add("password", "too_short")
	}

	if len(u.Password) > 0 && len(u.Password) < 10 && !u.WithoutPassword {
		result.Add("password", "too_short")
	}

	return result.ToError()
}

func (self *User) OwnUrl(requestScheme, requestBaseUri string) string {
	return fmt.Sprintf("%s://%s/users/%s", requestScheme, requestBaseUri, self.Uuid)
}

func (self *User) Links(response map[string]map[string]string, requestScheme, requestBaseUri string) map[string]map[string]string {
	response["oAuthTokens"] = map[string]string{"href": fmt.Sprintf("%s://%s/users/%s/oauth-tokens", requestScheme, requestBaseUri, self.Uuid)}
	response["organizations"] = map[string]string{"href": fmt.Sprintf("%s://%s/users/%s/organizations", requestScheme, requestBaseUri, self.Uuid)}
	response["projects"] = map[string]string{"href": fmt.Sprintf("%s://%s/users/%s/projects", requestScheme, requestBaseUri, self.Uuid)}
	response["jobs"] = map[string]string{"href": fmt.Sprintf("%s://%s/users/%s/jobs", requestScheme, requestBaseUri, self.Uuid)}
	response["repositories"] = map[string]string{"href": fmt.Sprintf("%s://%s/users/%s/repositories", requestScheme, requestBaseUri, self.Uuid)}
	response["activities"] = map[string]string{"href": fmt.Sprintf("%s://%s/users/%s/activities", requestScheme, requestBaseUri, self.Uuid)}
	response["self"] = map[string]string{"href": self.OwnUrl(requestScheme, requestBaseUri)}
	response["mfa"] = map[string]string{"href": fmt.Sprintf("%s://%s/users/%s/mfa", requestScheme, requestBaseUri, self.Uuid)}
	response["blocks"] = map[string]string{"href": fmt.Sprintf("%s://%s/users/%s/blocks", requestScheme, requestBaseUri, self.Uuid)}
	response["sessions"] = map[string]string{"href": fmt.Sprintf("%s://%s/users/%s/sessions", requestScheme, requestBaseUri, self.Uuid)}
	response["verify-email"] = map[string]string{"href": fmt.Sprintf("%s://%s/users/%s/verify-email", requestScheme, requestBaseUri, self.Uuid)}
	return response
}

func (self *User) FindUser(store UserStore) (*User, error) {
	return self, nil
}

func NewUser(name, email, password string) *User {
	return &User{
		Name:     name,
		Email:    email,
		Password: password,
	}
}

func (self *User) totp() TotpToken {
	if self.totpToken == nil {
		otp, err := NewTotpTokenWithClock(self.TotpSecret)
		if err != nil {
			panic(err)
		}
		self.totpToken = otp
	}

	return self.totpToken
}

func (self *User) CurrentTotpToken() int32 {
	return self.totp().Now()
}

func (self *User) IsValidTotpToken(token int32) bool {
	for i := int64(-1); i <= 1; i++ {
		if self.totp().FromNow(i) == token {
			return true
		}
	}

	return false
}

func (self *User) GenerateTotpSecret() {
	self.TotpSecret = RandomTotpSecret()
}

func (self *User) EnableTotp(token int32) error {
	if self.IsValidTotpToken(token) {
		now := Clock.Now()
		self.TotpEnabledAt = &now
	} else {
		return ErrTotpTokenNotValid
	}

	return nil
}

func (self *User) DisableTotp(token int32) error {
	if self.IsValidTotpToken(token) {
		self.TotpSecret = ""
		self.TotpEnabledAt = nil
		return nil
	} else {
		return ErrTotpTokenNotValid
	}
}

func (self *User) TotpEnabled() bool {
	return self.TotpEnabledAt != nil
}

func (self *User) FindProject(store ProjectStore) (*Project, error) {
	if self.Uuid == "" {
		return nil, new(NotFoundError)
	}
	return store.FindByMemberUuid(self.Uuid)
}

func (self *User) OwnedBy(user *User) bool {
	return self.Uuid == user.Uuid
}

func (self *User) AuthorizationName() string { return "user" }

func (self *User) SubscribeTo(watchable Watchable, event string, subscriptions SubscriptionStore) error {
	subscription := NewSubscription(watchable, event, self.Uuid)
	if _, err := subscriptions.Create(subscription); err != nil {
		return err
	}

	return nil
}

func (self *User) IsSubscribedTo(watchable Watchable, event string, subscriptions SubscriptionStore) (bool, error) {

	subscription, err := subscriptions.Find(watchable.Id(), event, self.Uuid)
	if err != nil {
		// Be on the safe side and do not send out any
		// notification unless we are *sure* the user solicited one.
		return false, err
	}

	return subscription != nil, nil
}

func (self *User) UnsubscribeFrom(watchable Watchable, event string, subscriptions SubscriptionStore) error {
	subscription, err := subscriptions.Find(watchable.Id(), event, self.Uuid)
	if err != nil {
		return err
	}
	if err := subscriptions.Delete(subscription.Uuid); err != nil {
		return err
	}
	return nil
}

func (self *User) Watch(watchable Watchable, subscriptions SubscriptionStore) error {
	existingSubscriptions, err := self.SubscriptionsFor(watchable, subscriptions)
	if err != nil {
		return err
	}

	subscribedTo := 0

	for event, subscribed := range existingSubscriptions.Subscribed {
		if subscribed {
			continue
		}

		err := self.SubscribeTo(watchable, event, subscriptions)
		if err != nil {
			return err
		}

		subscribedTo++
	}

	if subscribedTo == 0 {
		return NewValidationError("subscriptions", "already_watching")
	}
	return nil
}

func (self *User) Unwatch(watchable Watchable, subscriptions SubscriptionStore) error {
	for _, event := range watchable.WatchableEvents() {
		err := self.UnsubscribeFrom(watchable, event, subscriptions)
		if err != nil {
			return err
		}
	}
	return nil
}

func (self *User) SubscriptionsFor(watchable Watchable, allSubscriptions SubscriptionStore) (*Subscriptions, error) {
	result := &Subscriptions{
		WatcherUuid:   self.Uuid,
		WatchableUuid: watchable.Id(),
		WatchableType: watchable.WatchableType(),
		Subscribed:    map[string]bool{},
	}

	subscribed, err := allSubscriptions.FindEventsForUser(watchable.Id(), self.Uuid)
	if err != nil {
		return result, err
	}

	for _, event := range watchable.WatchableEvents() {
		result.Subscribed[event] = false
	}
	for _, event := range subscribed {
		result.Subscribed[event] = true
	}

	return result, nil
}

func (self *User) Capabilities() []string {
	return userCapabilities
}

func (self *User) InvitedTo(project *Project, invitations InvitationStore) (bool, error) {
	invitation, err := invitations.FindByUserAndProjectUuid(self.Uuid, project.Uuid)
	if _, ok := err.(NotFoundError); ok {
		return false, nil
	}
	if invitation == nil {
		return false, err
	} else {
		return true, nil
	}
}

func (self *User) NewSession(userAgent, clientAddress string) *Session {
	return &Session{
		Uuid:          uuidhelper.MustNewV4(),
		UserUuid:      self.Uuid,
		CreatedAt:     time.Now(),
		UserAgent:     userAgent,
		ClientAddress: clientAddress,
	}
}

func (self *User) NewBlock(reason string) (*UserBlock, error) {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return nil, NewValidationError("reason", "empty")
	}

	return &UserBlock{
		Uuid:     uuidhelper.MustNewV4(),
		UserUuid: self.Uuid,
		Reason:   reason,
	}, nil
}

func (self *User) HMAC(key []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(self.Uuid))
	mac.Write([]byte(self.PasswordHash))
	return mac.Sum(nil)
}

func (self *User) Scrub() *User {
	if self == nil {
		self = &User{}
	}
	copy := *self
	copy.Password = ""
	return &copy
}
