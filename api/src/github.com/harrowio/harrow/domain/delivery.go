package domain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"
)

type Delivery struct {
	defaultSubject
	Uuid        string     `json:"uuid" db:"uuid"`
	WebhookUuid string     `json:"webhookUuid" db:"webhook_uuid"`
	DeliveredAt time.Time  `json:"deliveredAt" db:"delivered_at"`
	ArchivedAt  *time.Time `json:"archivedAt" db:"archived_at"`
	// The Schedule this Delivery triggered, optional
	ScheduleUuid *string          `json:"scheduleUuid" db:"schedule_uuid"`
	Request      DeliveredRequest `json:"request" db:"request"`

	gitRef             string
	repositoryFullName string
}

func (self *Delivery) OwnUrl(requestScheme, requestBase string) string {
	return fmt.Sprintf("%s://%s/deliveries/%s", requestScheme, requestBase, self.Uuid)
}

func (self *Delivery) Links(response map[string]map[string]string, requestScheme, requestBase string) map[string]map[string]string {
	response["self"] = map[string]string{
		"href": self.OwnUrl(requestScheme, requestBase),
	}

	response["webhook"] = map[string]string{
		"href": fmt.Sprintf("%s://%s/webhooks/%s", requestScheme, requestBase, self.WebhookUuid),
	}

	if self.ScheduleUuid != nil {
		response["schedule"] = map[string]string{
			"href": fmt.Sprintf("%s://%s/schedules/%s", requestScheme, requestBase, *self.ScheduleUuid),
		}
	}

	return response
}

func (self *Delivery) AuthorizationName() string { return "delivery" }

func (self *Delivery) FindProject(projects ProjectStore) (*Project, error) {
	return projects.FindByWebhookUuid(self.WebhookUuid)
}

func (self *Delivery) GitRef() string {
	if self.gitRef == "" {
		self.parseDelivery()
	}

	return self.gitRef
}

func (self *Delivery) parseDelivery() {
	body, err := ioutil.ReadAll(self.Request.Body)
	if err != nil {
		panic(err)
	}

	// ensure that the request body can be read again
	self.Request.Body = ioutil.NopCloser(bytes.NewReader(body))

	gh := &githubWebhookDelivery{}
	if err := json.Unmarshal(body, &gh); err != nil {
		return
	}

	if gh.Ref == "" {
		self.parseBitBucketDelivery(body)
	} else {
		self.gitRef = gh.Ref
		self.repositoryFullName = gh.RepositoryName()
	}
}

func (self *Delivery) RepositoryName() string {
	if self.repositoryFullName == "" {
		self.parseDelivery()
	}
	return self.repositoryFullName
}

func (self *Delivery) parseBitBucketDelivery(body []byte) {
	bb := bitbucketWebhookDelivery{}
	if err := json.Unmarshal(body, &bb); err != nil {
		return
	}

	self.gitRef = bb.GitRef()
	self.repositoryFullName = bb.RepositoryName()
}

func (self *Delivery) OperationParameters(projectUuid string, repositories RepositoriesByName) *OperationParameters {
	result := &OperationParameters{}
	result.Init()
	result.Reason = OperationTriggeredByWebhook
	result.TriggeredByDelivery = self.Uuid
	repositoryName := self.RepositoryName()
	found, err := repositories.FindAllByProjectUuidAndRepositoryName(projectUuid, repositoryName)
	if err != nil {
		return result
	}

	for _, repository := range found {
		result.Checkout[repository.Uuid] = self.GitRef()
	}

	return result
}

type githubWebhookDelivery struct {
	Ref        string `json:"ref"`
	Repository struct {
		FullName string `json:"full_name"`
	} `json:"repository"`
}

func (self *githubWebhookDelivery) RepositoryName() string {
	return self.Repository.FullName
}

type bitbucketWebhookDelivery struct {
	Repository struct {
		FullName string `json:"full_name"`
	} `json:"repository"`
	Push struct {
		Changes []struct {
			New struct {
				Name string
			} `json:"new"`
		} `json:"changes"`
	} `json:"push"`
}

func (self *bitbucketWebhookDelivery) GitRef() string {
	if len(self.Push.Changes) == 0 {
		return ""
	}
	return self.Push.Changes[0].New.Name
}

func (self *bitbucketWebhookDelivery) RepositoryName() string {
	return self.Repository.FullName
}
