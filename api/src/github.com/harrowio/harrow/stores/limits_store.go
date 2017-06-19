package stores

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/harrowio/harrow/config"
	"github.com/harrowio/harrow/limits"
	"github.com/harrowio/harrow/logger"
	"github.com/jmoiron/sqlx"
)

type DbLimitsStore struct {
	tx        *sqlx.Tx
	cachePath string
	log       logger.Logger
}

func NewDbLimitsStore(tx *sqlx.Tx) *DbLimitsStore {
	c := config.GetConfig()
	return &DbLimitsStore{
		tx:        tx,
		cachePath: c.LimitsStoreConfig().CachePath,
	}
}

func (self *DbLimitsStore) Log() logger.Logger {
	if self.log == nil {
		self.log = logger.Discard
	}
	return self.log
}

func (self *DbLimitsStore) SetLogger(l logger.Logger) {
	self.log = l
}

// FindByOrganizationUuid returns the limits for organization
// identified by organization uuid.  This function makes uses of a
// cache when loading limits to avoid having to iterate over the full
// history.
func (self *DbLimitsStore) FindByOrganizationUuid(organizationUuid string) (*limits.Limits, error) {
	limits := self.LoadFromCache(organizationUuid)
	since := limits.Version()
	if since.IsZero() {
		organization, err := NewDbOrganizationStore(self.tx).FindByUuid(organizationUuid)
		if err != nil {
			return nil, err
		}

		since = organization.CreatedAt.Add(-1 * time.Minute)
	}
	activities := NewDbActivityStore(self.tx)
	if err := activities.AllSince(since, limits.HandleActivity); err != nil {
		return nil, err
	}

	self.CacheLimits(organizationUuid, limits)

	return limits, nil
}

func (self *DbLimitsStore) CacheLimits(organizationUuid string, limits *limits.Limits) {
	filename := filepath.Join("/mnt/gluster/harrow/limits", organizationUuid+".json")
	os.MkdirAll(filepath.Dir(filename), 0755)
	file, err := os.Create(filename)
	if err != nil {
		self.Log().Info().Msgf("failed to cache limits for organization %s: %s", organizationUuid, err)
		return
	}
	defer file.Close()

	if err := json.NewEncoder(file).Encode(limits); err != nil {
		os.Remove(filename)
		self.Log().Info().Msgf("failed to cache limits for organization %s: %s", organizationUuid, err)
	}
}

func (self *DbLimitsStore) LoadFromCache(organizationUuid string) *limits.Limits {
	filename := filepath.Join("/mnt/gluster/harrow/limits", organizationUuid+".json")
	file, err := os.Open(filename)
	if err != nil {
		return limits.NewLimits(organizationUuid, time.Now())
	}
	defer file.Close()

	result := limits.NewLimits(organizationUuid, time.Now())
	if err := json.NewDecoder(file).Decode(result); err != nil {
		return limits.NewLimits(organizationUuid, time.Now())
	}

	return result
}
