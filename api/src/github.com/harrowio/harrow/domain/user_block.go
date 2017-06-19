package domain

import (
	"fmt"
	"time"
)

type UserBlock struct {
	defaultSubject
	Uuid      string     `json:"uuid" db:"uuid"`
	UserUuid  string     `json:"userUuid" db:"user_uuid"`
	Reason    string     `json:"reason" db:"reason"`
	ValidFrom *time.Time `json:"validFrom" db:"valid_from"`
	ValidTo   *time.Time `json:"validTo" db:"valid_to"`
}

func (self *UserBlock) BlockForever(now time.Time) {
	self.ValidFrom = &now
	self.ValidTo = nil
}

func (self *UserBlock) OwnUrl(requestScheme, requestBase string) string {
	return fmt.Sprintf("%s://%s/user-blocks/%s", requestScheme, requestBase, self.Uuid)
}

func (self *UserBlock) Links(response map[string]map[string]string, requestScheme, requestBase string) map[string]map[string]string {
	response["self"] = map[string]string{"href": self.OwnUrl(requestScheme, requestBase)}
	response["user"] = map[string]string{"href": fmt.Sprintf("%s://%s/users/%s", self.UserUuid)}
	return response
}
