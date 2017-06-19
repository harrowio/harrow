package domain

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/harrowio/harrow/uuidhelper"
)

type Webhook struct {
	defaultSubject
	Uuid        string     `json:"uuid" db:"uuid"`
	ProjectUuid string     `json:"projectUuid" db:"project_uuid"`
	CreatorUuid string     `json:"creatorUuid" db:"creator_uuid"`
	JobUuid     string     `json:"jobUuid" db:"job_uuid"`
	Name        string     `json:"name" db:"name"`
	Slug        string     `json:"slug" db:"slug"`
	CreatedAt   time.Time  `json:"createdAt" db:"created_at"`
	ArchivedAt  *time.Time `json:"archivedAt" db:"archived_at"`
}

func NewWebhook(projectUuid, creatorUuid, jobUuid, name string) *Webhook {
	result := &Webhook{
		ProjectUuid: projectUuid,
		CreatorUuid: creatorUuid,
		JobUuid:     jobUuid,
		Name:        name,
	}

	result.GenerateSlug()

	return result
}

func (self *Webhook) OwnUrl(requestScheme, requestBase string) string {
	return fmt.Sprintf("%s://%s/webhooks/%s", requestScheme, requestBase, self.Uuid)
}

func (self *Webhook) Links(response map[string]map[string]string, requestScheme, requestBase string) map[string]map[string]string {
	response["self"] = map[string]string{"href": self.OwnUrl(requestScheme, requestBase)}
	response["project"] = map[string]string{"href": fmt.Sprintf("%s://%s/projects/%s", requestScheme, requestBase, self.ProjectUuid)}
	response["deliver"] = map[string]string{"href": fmt.Sprintf("%s://%s/wh/%s", requestScheme, requestBase, self.Slug)}
	response["deliveries"] = map[string]string{"href": fmt.Sprintf("%s://%s/webhooks/%s/deliveries", requestScheme, requestBase, self.Uuid)}
	response["slug"] = map[string]string{"href": fmt.Sprintf("%s://%s/webhooks/%s/slug", requestScheme, requestBase, self.Uuid)}
	response["creator"] = map[string]string{"href": fmt.Sprintf("%s://%s/users/%s", requestScheme, requestBase, self.CreatorUuid)}
	response["job"] = map[string]string{"href": fmt.Sprintf("%s://%s/jobs/%s", requestScheme, requestBase, self.JobUuid)}
	return response
}

func (self *Webhook) AuthorizationName() string { return "webhook" }

func (self *Webhook) FindProject(projects ProjectStore) (*Project, error) {
	return projects.FindByUuid(self.ProjectUuid)
}

func (self *Webhook) NewDelivery(req *http.Request) *Delivery {
	return &Delivery{
		Uuid:        uuidhelper.MustNewV4(),
		Request:     DeliveredRequest{req},
		WebhookUuid: self.Uuid,
	}
}

func (self *Webhook) Validate() error {
	err := NewValidationError("", "")

	if self.Slug == "" {
		err.Add("slug", "empty")
	}

	if self.Name == "" {
		err.Add("name", "empty")
	}

	if self.JobUuid == "" {
		err.Add("jobUuid", "empty")
	}

	return err.ToError()
}

func (self *Webhook) GenerateSlug() {
	slug := make([]byte, 8)
	_, err := rand.Read(slug)
	if err != nil {
		panic("Webhook.GenerateSlug" + err.Error())
	}

	self.Slug = hex.EncodeToString(slug)
}

// IsInternal returns true if the webhook has been created internally
// by Harrow (e.g. for supporting job notifiers).
func (self *Webhook) IsInternal() bool {
	return strings.HasPrefix(self.Name, "urn:harrow:job-notifier")
}
