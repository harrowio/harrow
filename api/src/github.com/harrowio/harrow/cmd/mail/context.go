package harrowMail

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/harrowio/harrow/activities"
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/hmail"
)

type File interface {
	Name() string
	Open() (io.ReadCloser, error)
	Children() ([]File, error)
}

type OSFile struct {
	File *os.File
}

func NewOSFile(f *os.File) *OSFile {
	return &OSFile{File: f}
}

func (self *OSFile) Name() string {
	return self.File.Name()
}

func (self *OSFile) Open() (io.ReadCloser, error) {
	return os.Open(self.File.Name())
}

func (self *OSFile) Children() ([]File, error) {
	childNames, err := self.File.Readdirnames(0)
	if err != nil {
		return nil, err
	}

	result := []File{}
	for _, childName := range childNames {
		path := filepath.Join(self.Name(), childName)
		child, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		result = append(result, NewOSFile(child))
	}

	return result, nil
}

type Context struct {
	ToAddress   string
	FromAddress string
	UrlHost     string

	Activity  *domain.Activity
	Project   *domain.Project
	Job       *domain.Job
	Operation *domain.Operation
	Notifier  struct {
		Type     string
		Notifier interface{}
	}
}

func NewContext(fromAddress string) *Context {
	return &Context{
		FromAddress: fromAddress,
	}
}

func (self *Context) LoadFromDirectory(dir File) error {
	children, err := dir.Children()
	if err != nil {
		return err
	}

	for _, child := range children {
		var dest interface{}
		switch filepath.Base(child.Name()) {
		case "activity.json":
			reader, err := child.Open()
			if err != nil {
				return err
			}
			data, err := ioutil.ReadAll(reader)
			reader.Close()
			if err != nil {
				return err
			}
			activity, err := activities.UnmarshalJSON(data)
			if err != nil {
				return err
			}

			self.Activity = activity
			continue
		case "project.json":
			dest = &self.Project
		case "job.json":
			dest = &self.Job
		case "operation.json":
			dest = &self.Operation
		case "notifier.json":
			dest = &self.Notifier
		default:
			continue
		}

		reader, err := child.Open()
		if err != nil {
			return err
		}
		defer reader.Close()
		if err := json.NewDecoder(reader).Decode(dest); err != nil {
			return err
		}
	}

	if self.Notifier.Type == "email_notifiers" {
		notifier, ok := self.Notifier.Notifier.(map[string]interface{})
		if ok {
			recipient, ok := notifier["recipient"]
			if ok {
				self.ToAddress = recipient.(string)
			}
			urlHost, ok := notifier["urlHost"]
			if ok {
				self.UrlHost = urlHost.(string)
				self.FromAddress = fmt.Sprintf("notifications@%s", urlHost)
			}
		}
	}

	return nil
}

func (self *Context) ToMailContext() (*hmail.MailContext, error) {
	handler, found := ActivityHandlers[self.Activity.Name]
	if !found {
		return nil, fmt.Errorf("No handler defined for %q", self.Activity.Name)
	}

	return handler(self)
}

func (self *Context) MailTemplateDir(basedir string) (File, error) {
	templateName := filepath.Join(basedir,
		strings.Replace(self.Activity.Name, ".", "/", -1),
	)

	dir, err := os.Open(templateName)
	if err != nil {
		return nil, err
	}

	return NewOSFile(dir), nil
}
