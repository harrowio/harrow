package projector

import (
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/logger"
)

type ActivityHandler interface {
	SubscribedTo() []string
	HandleActivity(IndexTransaction, *domain.Activity) error
}

type BroadcastHandler struct {
	log       logger.Logger
	handlers  []ActivityHandler
	histogram map[string]int
	seen      int
}

func NewBroadcastHandler(log logger.Logger) *BroadcastHandler {
	return &BroadcastHandler{
		log:       log,
		histogram: map[string]int{},
	}
}

func (self *BroadcastHandler) SubscribedTo() []string {
	result := []string{}
	for _, handler := range self.handlers {
		result = append(result, handler.SubscribedTo()...)
	}

	return result
}

func (self *BroadcastHandler) Add(handler ActivityHandler) *BroadcastHandler {
	self.handlers = append(self.handlers, handler)
	return self
}

func (self *BroadcastHandler) LogStatus() {
	self.log.Info().Msgf("seen=%d", self.seen)
	for name, count := range self.histogram {
		self.log.Info().Msgf("%s=%d", name, count)
	}
}

func (self *BroadcastHandler) HandleActivity(tx IndexTransaction, activity *domain.Activity) error {
	defer func() {
		if r := recover(); r != nil {
			self.LogStatus()
			self.log.Info().Msgf("last_processed=%s@%d", activity.Name, activity.Id)
			panic(r)
		}
	}()

	self.seen++
	self.histogram[activity.Name]++
	if self.seen%1000 == 0 {
		self.LogStatus()
	}

	for _, handler := range self.handlers {
		if err := handler.HandleActivity(tx, activity); err != nil {
			self.log.Error().Msgf("%T: %s", handler, err)
		}
	}

	return nil
}
