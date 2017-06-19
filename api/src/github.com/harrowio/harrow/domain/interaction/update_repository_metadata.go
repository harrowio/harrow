package interaction

import (
	"github.com/harrowio/harrow/activities"
	"github.com/harrowio/harrow/bus/activity"
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/logger"
)

// An ActivitySink is used for emitting activities over some form of
// messaging system.
type ActivitySink interface {
	EmitActivity(activity *domain.Activity)
}

type RepositoryMetaDataStore interface {
	UpdateMetadata(repositoryUuid string, metadata *domain.RepositoryMetaData) error
}

type UpdateRepositoryMetaData struct {
	activitySink  ActivitySink
	metadataStore RepositoryMetaDataStore
}

func NewUpdateRepositoryMetaData(activitySink ActivitySink, metadataStore RepositoryMetaDataStore) *UpdateRepositoryMetaData {
	return &UpdateRepositoryMetaData{
		activitySink:  activitySink,
		metadataStore: metadataStore,
	}
}

func (self *UpdateRepositoryMetaData) Update(repositoryUuid string, old, new *domain.RepositoryMetaData) error {
	if old != nil && !old.IsEmpty() {
		changes := old.Changes(new)
		activities := activities.FromRepositoryMetaDataChanges(changes, repositoryUuid)

		for _, activity := range activities {
			self.activitySink.EmitActivity(activity)
		}
	}

	return self.metadataStore.UpdateMetadata(repositoryUuid, new)
}

type BusActivitySink struct {
	sink activity.Sink
	log  logger.Logger
}

func NewBusActivitySink(publisher activity.Sink, log logger.Logger) *BusActivitySink {
	return &BusActivitySink{
		sink: publisher,
		log:  log,
	}
}

// EmitActivity emits an activity by sending it over the activity bus.
func (self *BusActivitySink) EmitActivity(activity *domain.Activity) {
	self.log.Info().Msgf("BusActivitySink: publish %s:\n%#v\n", activity.Name, activity.Payload)
	if err := self.sink.Publish(activity); err != nil {
		self.log.Error().Msgf("BusActivitySink: %s", err)
	}
}
