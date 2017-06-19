package domain

import "fmt"

// Subscriptions is aggregates all event subscriptions of a single user
// on a given watchable.
type Subscriptions struct {
	defaultSubject
	WatcherUuid   string          `json:"watcherUuid"`
	WatchableUuid string          `json:"watchableUuid"`
	WatchableType string          `json:"watchableType"`
	Subscribed    map[string]bool `json:"subscribed"`
}

func (self *Subscriptions) OwnUrl(requestScheme, requestBase string) string {
	return fmt.Sprintf("%s://%s/%s/%s/subscriptions",
		requestScheme,
		requestBase,
		self.WatchableType,
		self.WatchableUuid,
	)
}

func (self *Subscriptions) Links(response map[string]map[string]string, requestScheme, requestBase string) map[string]map[string]string {
	response["self"] = map[string]string{
		"href": self.OwnUrl(requestScheme, requestBase),
	}

	response["watcher"] = map[string]string{
		"href": fmt.Sprintf("%s://%s/users/%s", requestScheme, requestBase, self.WatcherUuid),
	}

	response["watchable"] = map[string]string{
		"href": fmt.Sprintf("%s://%s/%s/%s", requestScheme, requestBase, self.WatchableType, self.WatchableUuid),
	}

	return response
}

func (self *Subscriptions) AuthorizationName() string { return "subscription" }

func (self *Subscriptions) FindUser(users UserStore) (*User, error) {
	return users.FindByUuid(self.WatcherUuid)
}

func (self *Subscriptions) OwnedBy(user *User) bool {
	if user == nil {
		return false
	}

	return self.WatcherUuid == user.Uuid
}
