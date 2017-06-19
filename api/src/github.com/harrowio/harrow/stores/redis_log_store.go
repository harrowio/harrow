package stores

import (
	"errors"

	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/logger"
)

type RedisLogStore struct {
	kv  KeyValueStore
	log logger.Logger
}

func NewRedisLogStore(kv KeyValueStore) *RedisLogStore {
	ls := &RedisLogStore{kv: kv}

	return ls
}

func (self *RedisLogStore) Log() logger.Logger {
	if self.log == nil {
		self.log = logger.Discard
	}
	return self.log
}

func (self *RedisLogStore) SetLogger(l logger.Logger) {
	self.log = l
}

func (self *RedisLogStore) FindByOperationUuid(uuid string, tepy string) (*domain.Loggable, error) {

	return self.FindByRange(uuid, tepy, 0, -1)
}

func (self *RedisLogStore) FindByRange(uuid string, tepy string, from, to int) (*domain.Loggable, error) {

	if tepy != domain.LoggableWorkspace {
		return nil, errors.New("RedisLogStore can only find workspace logs")
	}
	exists, err := self.kv.Exists(uuid)
	if err != nil || !exists {
		return nil, err
	}

	// lines are 1-indexed
	// fromIndex := (int64)(from - 1)
	lines, err := self.kv.LRange(uuid, int64(from), int64(to))
	if err != nil {
		return nil, err
	}
	logLines := domain.LogLinesFromSlice(lines, from)
	return domain.NewLoggable(uuid, logLines, domain.LoggableOK), nil
}

func (self *RedisLogStore) PersistLogLine(operationUuid string, logLine *domain.LogLine) error {

	if !logLine.IsInternal() {
		return self.kv.RPush(operationUuid, logLine.Msg)
	}
	return nil
}

func (self *RedisLogStore) OnFinished(operationUuid string) error {

	return self.kv.Del(operationUuid)
}
