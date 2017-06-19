package domain

import (
	"fmt"
	"reflect"

	"github.com/harrowio/harrow/config"
)

type Feature struct {
	defaultSubject
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}

func NewFeaturesFromConfig(features config.Features) []*Feature {
	result := []*Feature{}
	typeOfFeatures := reflect.TypeOf(features)
	valueOfFeatures := reflect.ValueOf(features)

	for i := 0; i < typeOfFeatures.NumField(); i++ {
		field := typeOfFeatures.Field(i)
		if field.Type.Kind() != reflect.Bool {
			continue
		}
		feature := &Feature{
			Name:    field.Tag.Get("json"),
			Enabled: valueOfFeatures.Field(i).Bool(),
		}

		result = append(result, feature)
	}

	return result
}

func (self *Feature) OwnUrl(requestScheme string, requestBaseUri string) string {
	return fmt.Sprintf("%s://%s/api-features/%s", requestScheme, requestBaseUri, self.Name)
}

func (self *Feature) Links(response map[string]map[string]string, requestScheme string, requestBaseUri string) map[string]map[string]string {
	response["self"] = map[string]string{"href": self.OwnUrl(requestScheme, requestBaseUri)}
	response["collection"] = map[string]string{"href": fmt.Sprintf("%s://%s/api-features", requestScheme, requestBaseUri)}
	return response
}
