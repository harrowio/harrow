package domain

import "fmt"

type ScriptCard struct {
	defaultSubject
	ProjectUuid                            string              `json:"projectUuid" db:"project_uuid"`
	ScriptUuid                             string              `json:"scriptUuid" db:"script_uuid"`
	ScriptName                             string              `json:"scriptName" db:"script_name"`
	LastOperation                          *Operation          `json:"lastOperation" db:"last_operation"`
	EnabledEnvironments                    []*Environment      `json:"enabledEnvironments"`
	RecentOperationStatusByEnvironmentUuid map[string][]string `json:"recentOperationStatusByEnvironmentUuid"`
}

func (self *ScriptCard) OwnUrl(requestScheme string, requestBaseUri string) string {
	return fmt.Sprintf("%s://%s/projects/%s/scripts/%s", requestScheme, requestBaseUri, self.ProjectUuid, self.ScriptUuid)
}

func (self *ScriptCard) Links(response map[string]map[string]string, requestScheme string, requestBaseUri string) map[string]map[string]string {
	response["self"] = map[string]string{
		"href": self.OwnUrl(requestScheme, requestBaseUri),
	}
	response["project"] = map[string]string{
		"href": fmt.Sprintf("%s://%s/projects/%s", requestScheme, requestBaseUri, self.ProjectUuid),
	}
	return response
}

func (self *ScriptCard) AuthorizationName() string { return "script-card" }
func (self *ScriptCard) FindProject(projects ProjectStore) (*Project, error) {
	return projects.FindByUuid(self.ProjectUuid)
}
