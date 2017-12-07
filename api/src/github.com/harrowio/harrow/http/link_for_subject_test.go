package http

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/test_helpers"
)

type testSubject struct {
	name  string
	links map[string]map[string]string
}

func NewTestSubject(name string) *testSubject {
	return &testSubject{
		name:  name,
		links: map[string]map[string]string{},
	}
}

func (self *testSubject) OwnUrl(requestScheme, requestBase string) string {
	return fmt.Sprintf("%s://%s/test-subject/%s", requestScheme, requestBase, self.name)
}

func (self *testSubject) Links(response map[string]map[string]string, requestScheme, requestBase string) map[string]map[string]string {
	for rel, link := range self.links {
		response[rel] = map[string]string{"href": link["href"]}
	}
	return response
}

func (self *testSubject) AddLink(rel, url string) {
	self.links[rel] = map[string]string{"href": url}
}

func (self *testSubject) AuthorizationName() string { return self.name }

func (self *testSubject) Embedded() map[string][]domain.Subject {
	return nil
}

func (self *testSubject) Embed(k string, subject domain.Subject) {
}

func Test_linksForSubject_mergesCapabilitiesIn(t *testing.T) {
	subject := NewTestSubject("foo")
	subject.AddLink("projects", "http://example.com/test-subjects/foo/projects")
	req, err := http.NewRequest("GET", "http://example.com/test-subject/foo", nil)
	if err != nil {
		t.Fatal(err)
	}

	auth := test_helpers.NewMockAuthzService()
	auth.CapabilitiesBySubjectMap = map[string][]string{
		"projects": {"create", "read"},
	}
	result := linksForSubject(auth, req, subject)

	if got, want := result["projects"]["create"], "POST"; !reflect.DeepEqual(got, want) {
		t.Errorf("projects.create = %v; want %v", got, want)
	}

	if got, want := result["projects"]["read"], "GET"; got != want {
		t.Errorf("projects.read = %v; want %v", got, want)
	}
}

func Test_linksForSubject_mergesCapabilitiesForSelfBasedOnAuthorizationName(t *testing.T) {
	subject := NewTestSubject("foo")
	subject.AddLink("self", "http://example.com/test-subjects/foo")
	req, err := http.NewRequest("GET", "http://example.com/test-subject/foo", nil)
	if err != nil {
		t.Fatal(err)
	}

	auth := test_helpers.NewMockAuthzService()
	auth.CapabilitiesBySubjectMap = map[string][]string{
		"foo": {"archive"},
	}
	result := linksForSubject(auth, req, subject)

	if got, want := result["self"]["archive"], "DELETE"; !reflect.DeepEqual(got, want) {
		t.Errorf("self.archive = %v; want %v", got, want)
	}
}

func Test_linksForSubject_mergesCapabilitiesForPluralResources(t *testing.T) {
	subject := NewTestSubject("foo")
	subject.AddLink("jobs", "http://example.com/test-subjects/foo/projects")
	req, err := http.NewRequest("GET", "http://example.com/test-subject/foo", nil)
	if err != nil {
		t.Fatal(err)
	}

	auth := test_helpers.NewMockAuthzService()

	auth.CapabilitiesBySubjectMap = map[string][]string{
		"job": {"create"},
	}
	result := linksForSubject(auth, req, subject)

	if got, want := result["jobs"]["create"], "POST"; !reflect.DeepEqual(got, want) {
		t.Errorf("jobs.create = %v; want %v", got, want)
	}
}

func Test_linksForSubject_pluralizesWordsEndingOnY(t *testing.T) {
	subject := NewTestSubject("foo")
	subject.AddLink("repositories", "http://example.com/test-subjects/foo/repositories")
	req, err := http.NewRequest("GET", "http://example.com/test-subject/foo", nil)
	if err != nil {
		t.Fatal(err)
	}

	auth := test_helpers.NewMockAuthzService()

	auth.CapabilitiesBySubjectMap = map[string][]string{
		"repository": {"create"},
	}
	result := linksForSubject(auth, req, subject)

	if got, want := result["repositories"]["create"], "POST"; !reflect.DeepEqual(got, want) {
		t.Errorf("repositories.create = %v; want %v", got, want)
	}
}
