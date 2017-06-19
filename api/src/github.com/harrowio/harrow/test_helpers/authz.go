package test_helpers

import (
	"fmt"
	"runtime"
	"strings"
	"testing"

	"github.com/harrowio/harrow/logger"
)

type MockAuthzService struct {
	reads                    int
	updates                  int
	archives                 int
	creates                  int
	CapabilitiesBySubjectMap map[string][]string
	log                      logger.Logger
}

func NewMockAuthzService() *MockAuthzService {
	return &MockAuthzService{
		CapabilitiesBySubjectMap: map[string][]string{},
	}
}

func (s *MockAuthzService) Log() logger.Logger {
	if s.log == nil {
		s.log = logger.Discard
	}
	return s.log
}

func (s *MockAuthzService) SetLogger(l logger.Logger) {
	s.log = l
}

func (s *MockAuthzService) CanRead(thing interface{}) (bool, error) {
	s.reads++
	return true, nil
}

func (s *MockAuthzService) CanUpdate(thing interface{}) (bool, error) {
	s.updates++
	return true, nil
}

func (s *MockAuthzService) CanArchive(thing interface{}) (bool, error) {
	s.archives++
	return true, nil
}

func (s *MockAuthzService) CanCreate(thing interface{}) (bool, error) {
	s.creates++
	return true, nil
}

func (s *MockAuthzService) CapabilitiesBySubject() map[string][]string {
	return s.CapabilitiesBySubjectMap
}

func (s *MockAuthzService) Can(action string, thing interface{}) (bool, error) {
	switch action {
	case "read":
		return s.CanRead(thing)
	case "update":
		return s.CanUpdate(thing)
	case "archive":
		return s.CanArchive(thing)
	case "create":
		return s.CanCreate(thing)
	default:
		return true, nil
	}
}

func (s *MockAuthzService) Expect(t *testing.T, method string, calls int) {
	_, file, line, _ := runtime.Caller(1)
	calledFrom := fmt.Sprintf("called from %s:%d", file, line)

	actual := 0
	switch method {
	case "create":
		actual = s.creates
	case "update":
		actual = s.updates
	case "archive":
		actual = s.archives
	case "read":
		actual = s.reads
	default:
		t.Fatalf("%s: No such authorization method: Can%s", calledFrom, strings.Title(method))
	}

	if actual != calls {
		t.Fatalf("%s: Expected %d calls to Can%s, got %d\n", calledFrom, calls, strings.Title(method), actual)
	}
}
