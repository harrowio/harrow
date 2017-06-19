package stores

import "gopkg.in/redis.v2"

type RedisKeyValueStore struct {
	c *redis.Client
}

func NewRedisKeyValueStore(c *redis.Client) KeyValueStore {
	return &RedisKeyValueStore{c}
}

func (self *RedisKeyValueStore) Get(key string) ([]byte, error) {
	r := self.c.Get(key)
	if r.Err() == redis.Nil {
		return nil, ErrKeyNotFound
	}
	if r.Err() != nil {
		return nil, r.Err()
	}
	return []byte(r.Val()), nil
}

func (self *RedisKeyValueStore) Exists(key string) (bool, error) {
	r := self.c.Exists(key)
	if r.Err() != nil {
		return false, r.Err()
	}
	return r.Val(), nil
}

func (self *RedisKeyValueStore) Set(key string, data []byte) error {
	r := self.c.Set(key, string(data))
	return r.Err()
}

func (self *RedisKeyValueStore) Del(key string) error {
	r := self.c.Del(key)
	return r.Err()
}

func (self *RedisKeyValueStore) LRange(key string, start, stop int64) ([]string, error) {
	r := self.c.LRange(key, start, stop)
	if r.Err() != nil {
		return nil, r.Err()
	}
	return r.Val(), nil
}

func (self *RedisKeyValueStore) RPush(key string, data string) error {
	r := self.c.RPush(key, data)
	return r.Err()
}

func (self *RedisKeyValueStore) LPush(key string, data string) error {
	r := self.c.LPush(key, data)
	return r.Err()
}

func (self *RedisKeyValueStore) SMembers(key string) ([]string, error) {
	r := self.c.SMembers(key)
	if r.Err() != nil {
		return nil, r.Err()
	}
	return r.Val(), nil
}

func (self *RedisKeyValueStore) SIsMember(key, member string) (bool, error) {
	r := self.c.SIsMember(key, member)
	if r.Err() != nil {
		return false, r.Err()
	}
	return r.Val(), nil
}

func (self *RedisKeyValueStore) SAdd(key, member string) (int64, error) {
	r := self.c.SAdd(key, member)
	if r.Err() != nil {
		return 0, r.Err()
	}
	return r.Val(), nil
}

func (self *RedisKeyValueStore) SRem(key, member string) (int64, error) {
	r := self.c.SRem(key, member)
	if r.Err() != nil {
		return 0, r.Err()
	}
	return r.Val(), nil
}
