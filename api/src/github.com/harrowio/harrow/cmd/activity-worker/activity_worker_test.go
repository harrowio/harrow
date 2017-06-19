package activityWorker

import (
	"errors"
	"sync"
	"testing"

	"github.com/harrowio/harrow/bus/activity"
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/test_helpers"
)

func Test_ActivityWorker_receivesNewMessagesFromBus(t *testing.T) {
	bus := activity.NewMemoryTransport()
	payload := domain.NewActivity(1, "test.run")
	received := 0
	recorded := new(sync.WaitGroup)
	recorded.Add(1)
	record := func(msg activity.Message) {
		if got, want := msg.Activity().Name, payload.Name; got != want {
			t.Errorf("msg.Activity().Name = %q; want %q", got, want)
		} else {
			received++
		}
		recorded.Done()
	}
	worker := NewActivityWorker(bus, NewMemoryActivityStore()).
		AddMessageHandler(record)

	worker.Start()
	bus.Publish(payload)
	recorded.Wait()
	if got, want := received, 1; got != want {
		t.Errorf("received = %d; want %d", got, want)
	}
}

func Test_ActivityWorker_storesReceivedMessages(t *testing.T) {
	bus := activity.NewMemoryTransport()
	wait := new(sync.WaitGroup)

	payloads := []*domain.Activity{
		domain.NewActivity(1, "test.run"),
		domain.NewActivity(2, "test.run"),
	}

	wait.Add(len(payloads))

	store := NewMemoryActivityStore()
	worker := NewActivityWorker(bus, store).AddMessageHandler(
		func(msg activity.Message) { wait.Done() },
	)

	worker.Start()
	for _, payload := range payloads {
		bus.Publish(payload)
	}
	wait.Wait()
	if got, want := len(store.All), len(payloads); got != want {
		t.Errorf("len(store.All) = %d; want %d", got, want)
	}
}

func Test_ActivityWorker_doesNotAcknowledgeMessageIfStorageFails(t *testing.T) {
	bus := activity.NewMemoryTransport()
	payload := domain.NewActivity(1, "test.run")
	received := (*activity.MemoryMessage)(nil)
	wait := new(sync.WaitGroup)
	record := func(msg activity.Message) {
		received = msg.(*activity.MemoryMessage)
		wait.Done()
	}
	store := NewMemoryActivityStore().
		FailWith(errors.New("failed to store message"))

	worker := NewActivityWorker(bus, store).
		AddMessageHandler(record)

	worker.Start()
	wait.Add(1)
	bus.Publish(payload)
	wait.Wait()
	worker.Stop()

	if got, want := received, (*activity.MemoryMessage)(nil); got == want {
		t.Errorf(`received = %v; want not %v`, got, want)
	}

	if got, want := received.Acknowledged, false; got != want {
		t.Errorf("received.Acknowledged = %v; want %v", got, want)
	}
}

func Test_ActivityWorker_acknowledgesMessageIfStorageSucceeds(t *testing.T) {
	bus := activity.NewMemoryTransport()
	payload := domain.NewActivity(1, "test.run")
	received := (*activity.MemoryMessage)(nil)
	wait := new(sync.WaitGroup)
	wait.Add(1)
	record := func(msg activity.Message) {
		wait.Done()
		received = msg.(*activity.MemoryMessage)
	}
	store := NewMemoryActivityStore().
		FailWith(nil)

	worker := NewActivityWorker(bus, store).
		AddMessageHandler(record)

	worker.Start()
	bus.Publish(payload)
	wait.Wait()
	if got, want := received.Acknowledged, true; got != want {
		t.Errorf("received.Acknowledged = %v; want %v", got, want)
	}
}

func Test_ActivityWorker_enrichesActivityWithProjectMemberUuids_ifActivitySourceBelongsToProject(t *testing.T) {
	tx := test_helpers.GetDbTx(t)
	defer tx.Rollback()

	world := test_helpers.MustNewWorld(tx, t)
	bus := activity.NewMemoryTransport()
	activityStore := NewMemoryActivityStore()

	payload := domain.NewActivity(1, "test.belongs-to-project")
	payload.Payload = world.Project("public")

	wait := new(sync.WaitGroup)
	wait.Add(1)
	received := (*activity.MemoryMessage)(nil)
	record := func(msg activity.Message) {
		wait.Done()
		received = msg.(*activity.MemoryMessage)
	}
	worker := NewActivityWorker(bus, activityStore).
		AddMessageHandler(listProjectMembersTx(tx)).
		AddMessageHandler(record)

	worker.Start()
	bus.Publish(payload)
	wait.Wait()
	if got := len(received.Activity().Audience()); got <= 0 {
		t.Errorf(`len(received.Activity().Audience()) = %d; want > 0`, got)
	}
}

func Test_ActivityWorker_enrichesActivityWithProjectUuid_ifActivitySourceBelongsToProject(t *testing.T) {
	tx := test_helpers.GetDbTx(t)
	defer tx.Rollback()

	world := test_helpers.MustNewWorld(tx, t)
	bus := activity.NewMemoryTransport()
	activityStore := NewMemoryActivityStore()

	project := world.Project("public")
	payload := domain.NewActivity(1, "test.belongs-to-project")
	payload.Payload = project

	wait := new(sync.WaitGroup)
	wait.Add(1)
	received := (*activity.MemoryMessage)(nil)
	record := func(msg activity.Message) {
		wait.Done()
		received = msg.(*activity.MemoryMessage)
	}
	worker := NewActivityWorker(bus, activityStore).
		AddMessageHandler(listProjectMembersTx(tx)).
		AddMessageHandler(markProjectUuidTx(tx)).
		AddMessageHandler(record)

	worker.Start()
	bus.Publish(payload)
	wait.Wait()
	if got, want := received.Activity().ProjectUuid(), project.Uuid; got != want {
		t.Errorf(`received.Activity().ProjectUuid() = %v; want %v`, got, want)
	}
}
