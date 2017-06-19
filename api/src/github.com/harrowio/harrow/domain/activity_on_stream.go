package domain

import (
	"fmt"
	"time"
)

// ActivityOnStream represents an activity as it is displayed on an
// activity stream.
type ActivityActor struct {
	Name string `json:"name"`
	Uuid string `json:"uuid"`
}

type ActivityOnStream struct {
	defaultSubject
	Id         int            `json:"id"`
	UserUuid   string         `json:"userUuid"`
	Project    *ActivityActor `json:"project"`
	OccurredOn time.Time      `json:"occurredOn"`
	Actor      *ActivityActor `json:"actor"`
	Action     string         `json:"action"`
	Object     string         `json:"object"`
	Subject    *ActivityActor `json:"subject"`
	Unread     bool           `json:"unread"`
}

func (self *ActivityOnStream) OwnUrl(requestBase, requestScheme string) string {
	return fmt.Sprintf("%s://%s/users/%s/activities/%d",
		requestBase,
		requestScheme,
		self.UserUuid,
		self.Id,
	)
}

func (self *ActivityOnStream) Links(response map[string]map[string]string, requestScheme, requestBase string) map[string]map[string]string {
	response["self"] = map[string]string{
		"href": self.OwnUrl(requestScheme, requestBase),
	}
	response["read-status"] = map[string]string{
		"href": fmt.Sprintf("%s://%s/activities/%d/read-status", requestScheme, requestBase, self.Id),
	}

	return response
}

func (self *ActivityOnStream) FindUser(users UserStore) (*User, error) {
	return users.FindByUuid(self.UserUuid)
}
