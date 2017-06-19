package domain

import (
	"time"

	"github.com/harrowio/harrow/uuidhelper"
)

type Subscription struct {
	Uuid          string     `json:"uuid" db:"uuid"`
	UserUuid      string     `json:"userUuid" db:"user_uuid"`
	WatchableUuid string     `json:"watchableUuid" db:"watchable_uuid"`
	WatchableType string     `json:"watchableType" db:"watchable_type"`
	EventName     string     `json:"eventName" db:"event_name"`
	ArchivedAt    *time.Time `json:"archivedAt" db:"archived_at"`
	CreatedAt     *time.Time `json:"createdAt" db:"created_at"`
}

func NewSubscription(watchable Watchable, event, userUuid string) *Subscription {
	return &Subscription{
		Uuid:          uuidhelper.MustNewV4(),
		UserUuid:      userUuid,
		WatchableUuid: watchable.Id(),
		WatchableType: watchable.WatchableType(),
		EventName:     event,
	}
}

func (self *Subscription) AuthorizationName() string { return "subscription" }

func (self *Subscription) FindUser(users UserStore) (*User, error) {
	return users.FindByUuid(self.UserUuid)
}

func (self *Subscription) OwnedBy(user *User) bool {
	return user.Uuid == self.UserUuid
}
