package activities

import "github.com/harrowio/harrow/domain"

type ScriptEditorPayload struct {
	ProjectUuid     string `json:"projectUuid"`
	TaskUuid        string `json:"taskUuid"`
	EnvironmentUuid string `json:"environmentUuid"`
}

func init() {
	registerPayload(ScriptEditorSaved(&ScriptEditorPayload{}))
	registerPayload(ScriptEditorTested(&ScriptEditorPayload{}))
}

func ScriptEditorSaved(payload *ScriptEditorPayload) *domain.Activity {
	return &domain.Activity{
		Name:       "script-editor.saved",
		Extra:      map[string]interface{}{},
		OccurredOn: Clock.Now(),
		Payload:    payload,
	}
}

func ScriptEditorTested(payload *ScriptEditorPayload) *domain.Activity {
	return &domain.Activity{
		Name:       "script-editor.tested",
		Extra:      map[string]interface{}{},
		OccurredOn: Clock.Now(),
		Payload:    payload,
	}
}
