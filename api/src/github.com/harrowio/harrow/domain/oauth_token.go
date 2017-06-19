package domain

import (
	"time"
)

type OAuthToken struct {
	defaultSubject
	Uuid        string     `json:"uuid"`
	UserUuid    string     `json:"userUuid"    db:"user_uuid"`
	Provider    string     `json:"provider"`
	Scope       string     `json:"scope"       db:"scopes"`
	AccessToken string     `json:"accessToken" db:"access_token"`
	TokenType   string     `json:"tokenType"   db:"token_type"`
	CreatedAt   *time.Time `json:"createdAt"   db:"created_at"`
}

func (self *OAuthToken) OwnUrl(requestScheme, requestBaseUri string) string {
	return ""
}

func (self *OAuthToken) Links(response map[string]map[string]string, requestScheme, requestBaseUri string) map[string]map[string]string {
	return response
}

func (self *OAuthToken) AuthorizationName() string { return "oauth-token" }

// OwnedBy satisfies authz.Ownable by identifying the user for which
// this OAuth token is valid.
func (self *OAuthToken) OwnedBy(user *User) bool {
	return self.UserUuid == user.Uuid
}
