package domain

import (
	"fmt"
	"strings"

	"github.com/harrowio/harrow/uuidhelper"
)

type Stencil struct {
	defaultSubject
	Id             string `json:"id"`
	ProjectUuid    string `json:"projectUuid,omitempty"`
	UserUuid       string `json:"userUuid,omitempty"`
	NotifyViaEmail string `json:"notifyViaEmail,omitempty"`
	UrlHost        string `json:"urlHost,omitempty"`
}

func (self *Stencil) OwnUrl(requestScheme, requestBase string) string {
	return fmt.Sprintf("%s://%s/stencils/%s", requestScheme, requestBase, self.Id)
}

func (self *Stencil) Links(response map[string]map[string]string, requestScheme, requestBase string) map[string]map[string]string {
	response["self"] = map[string]string{
		"href": self.OwnUrl(requestScheme, requestBase),
	}

	return response
}

func (self *Stencil) AuthorizationName() string { return "stencil" }

func (self *Stencil) FindProject(projects ProjectStore) (*Project, error) {
	return projects.FindByUuid(self.ProjectUuid)
}

func (self *Stencil) Validate() error {
	err := EmptyValidationError()

	if !uuidhelper.IsValid(self.ProjectUuid) {
		err.Add("projectUuid", "malformed")
	}

	if self.NotifyViaEmail != "" {
		if !strings.Contains(self.NotifyViaEmail, "@") {
			err.Add("notifyViaEmail", "malformed")
		}

		if self.UrlHost == "" {
			err.Add("urlHost", "empty")
		}
	}

	if !uuidhelper.IsValid(self.UserUuid) {
		err.Add("userUuid", "malformed")
	}

	return err.ToError()
}
