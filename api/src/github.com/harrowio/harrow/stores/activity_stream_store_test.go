package stores_test

import (
	"testing"

	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/stores"
	"github.com/harrowio/harrow/test_helpers"
)

func TestKVActivityStreamStore_marksNewlyAddedActivityAsUnread(t *testing.T) {
	kv := test_helpers.NewMockKeyValueStore()
	subject := stores.NewKVActivityStreamStore(kv)
	activity := &domain.ActivityOnStream{
		Id: 1,
	}
	viewerUuid := "3b4b1369-93ad-4aee-8472-75e91fd95405"

	if err := subject.AddActivityToUserStream(activity, viewerUuid); err != nil {
		t.Fatal(err)
	}

	unread, err := subject.IsUnread(activity.Id, viewerUuid)
	if err != nil {
		t.Fatal(err)
	}

	if got, want := unread, true; got != want {
		t.Errorf(`unread = %v; want %v`, got, want)
	}
}

func TestKVActivityStreamStore_markingAnActivityAsUnreadWorks(t *testing.T) {
	kv := test_helpers.NewMockKeyValueStore()
	subject := stores.NewKVActivityStreamStore(kv)
	activity := &domain.ActivityOnStream{
		Id: 1,
	}
	viewerUuid := "3b4b1369-93ad-4aee-8472-75e91fd95405"

	if err := subject.AddActivityToUserStream(activity, viewerUuid); err != nil {
		t.Fatal(err)
	}

	if err := subject.MarkAsRead(activity.Id, viewerUuid); err != nil {
		t.Fatal(err)
	}

	unread, err := subject.IsUnread(activity.Id, viewerUuid)
	if err != nil {
		t.Fatal(err)
	}

	if got, want := unread, false; got != want {
		t.Errorf(`unread = %v; want %v`, got, want)
	}

}
