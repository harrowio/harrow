package domain

import (
	"fmt"
	"time"
)

type Session struct {
	defaultSubject
	Uuid          string     `json:"uuid"`
	UserUuid      string     `json:"userUuid"      db:"user_uuid"`
	EndedAt       time.Time  `json:"-"             db:"ended_at"`
	LoadedAt      time.Time  `json:"loadedAt"      db:"loaded_at"`
	ExpiresAt     time.Time  `json:"expiresAt"     db:"expires_at"`
	CreatedAt     time.Time  `json:"createdAt"     db:"created_at"`
	ValidatedAt   time.Time  `json:"validatedAt"   db:"validated_at"`
	LoggedOutAt   *time.Time `json:"loggedOutAt"   db:"logged_out_at"`
	InvalidatedAt *time.Time `json:"invalidatedAt" db:"invalidated_at"`
	Valid         bool       `json:"valid"         db:"-"`
	UserAgent     string     `json:"userAgent"     db:"user_agent"`
	ClientAddress string     `json:"clientAddress" db:"client_address"`
}

func (s *Session) Validate() error {
	s.Valid = false

	if s.IsExpired() {
		return NewValidationError("session", "expired")
	}

	if s.ValidatedAt.IsZero() || s.ValidatedAt.Before(s.CreatedAt) {
		return NewValidationError("validatedAt", "never_validated")
	}

	if s.LoggedOutAt != nil {
		return NewValidationError("session", "logged_out")
	}

	s.Valid = true
	return nil
}

func (s *Session) IsExpired() bool {
	return s.LoadedAt.After(s.ExpiresAt)
}

func (s *Session) IsInvalidated() bool {
	return s.InvalidatedAt != nil
}

func (s *Session) LogOut() {
	t := time.Now()
	s.LoggedOutAt = &t
}

func (self *Session) OwnUrl(requestScheme, requestBaseUri string) string {
	return fmt.Sprintf("%s://%s/sessions/%s", requestScheme, requestBaseUri, self.Uuid)
}

func (self *Session) Links(response map[string]map[string]string, requestScheme, requestBaseUri string) map[string]map[string]string {
	response["self"] = map[string]string{"href": self.OwnUrl(requestScheme, requestBaseUri)}
	response["user"] = map[string]string{"href": fmt.Sprintf("%s://%s/users/%s", requestScheme, requestBaseUri, self.UserUuid)}
	return response
}

func (self *Session) FindUser(store UserStore) (*User, error) {
	return store.FindByUuid(self.UserUuid)
}

func (self *Session) OwnedBy(user *User) bool {
	return user.Uuid == self.UserUuid
}

func (self *Session) AuthorizationName() string { return "session" }
