package stores

import (
	"encoding/json"
	"fmt"

	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/logger"
)

type KVActivityStreamStore struct {
	kv  KeyValueStore
	log logger.Logger
}

func NewKVActivityStreamStore(kv KeyValueStore) *KVActivityStreamStore {
	return &KVActivityStreamStore{kv: kv}
}

func (self *KVActivityStreamStore) Log() logger.Logger {
	if self.log == nil {
		self.log = logger.Discard
	}
	return self.log
}

func (self *KVActivityStreamStore) SetLogger(l logger.Logger) {
	self.log = l
}

func (self *KVActivityStreamStore) FindStreamForUser(uuid string) ([]*domain.ActivityOnStream, error) {
	keys, err := self.kv.LRange(self.keyForUserStream(uuid), 0, 50)
	if err != nil {
		return nil, err
	}

	result := []*domain.ActivityOnStream{}
	for _, key := range keys {
		value, err := self.kv.Get(key)
		if err != nil {
			self.Log().Warn().Msgf("failed to get %s: %s", key, err)
			continue
		}

		item := domain.ActivityOnStream{}
		if err := json.Unmarshal(value, &item); err != nil {
			self.Log().Warn().Msgf("corrupt activity data: %s", err)
			continue
		}

		self.customizeActivityForViewer(&item, uuid)

		result = append(result, &item)
	}

	return result, nil
}

func (self *KVActivityStreamStore) keyForUserStream(userUuid string) string {
	return fmt.Sprintf("%s.activities", userUuid)
}

func (self *KVActivityStreamStore) customizeActivityForViewer(activity *domain.ActivityOnStream, viewerUuid string) {
	activity.UserUuid = viewerUuid
	activity.Unread = true
	if unread, err := self.IsUnread(activity.Id, viewerUuid); err == nil {
		activity.Unread = unread
	} else {
		self.Log().Info().Msgf("CustomizeActivityForViewer: %s", err)
	}
}

func (self *KVActivityStreamStore) keyForActivity(activity *domain.ActivityOnStream) string {
	return fmt.Sprintf("activity.%d", activity.Id)
}

func (self *KVActivityStreamStore) StoreActivity(activity *domain.ActivityOnStream) error {

	data, err := json.Marshal(activity)
	if err != nil {
		return err
	}

	return self.kv.Set(self.keyForActivity(activity), data)
}

func (self *KVActivityStreamStore) AddActivityToUserStream(activity *domain.ActivityOnStream, userUuid string) error {

	key := fmt.Sprintf("%s.activities", userUuid)
	value := self.keyForActivity(activity)
	if err := self.MarkAsUnread(activity.Id, userUuid); err != nil {
		self.Log().Info().Msgf("AddActivityToUserStream[%d]: %s", activity.Id, userUuid)
	}

	return self.kv.LPush(key, value)
}

func (self *KVActivityStreamStore) MarkAsUnread(activityId int, viewerUuid string) error {

	return self.kv.Set(self.keyForReadStatus(activityId, viewerUuid), []byte(`unread`))
}

func (self *KVActivityStreamStore) keyForReadStatus(activityId int, viewerUuid string) string {

	return fmt.Sprintf(`%s.activity.%d.unread`, viewerUuid, activityId)
}

func (self *KVActivityStreamStore) MarkAsRead(activityId int, viewerUuid string) error {

	return self.kv.Del(self.keyForReadStatus(activityId, viewerUuid))
}

func (self *KVActivityStreamStore) IsUnread(activityId int, viewerUuid string) (bool, error) {

	return self.kv.Exists(self.keyForReadStatus(activityId, viewerUuid))
}
