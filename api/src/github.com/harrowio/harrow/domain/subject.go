package domain

// Subject defines the methods necessary for creating a JSON-HAL compliant
// response object.
type Subject interface {
	// Url returns the URL referencing the subject itself
	OwnUrl(requestScheme, requestBaseUri string) string
	// Links fills in the _links object in the HAL response
	Links(response map[string]map[string]string, requestScheme, requestBaseUri string) map[string]map[string]string
	// Embedded is used for embedding relations
	Embedded() map[string][]Subject
}

type defaultSubject struct {
	embedded map[string][]Subject
}

func (self *defaultSubject) Links(response map[string]map[string]string, requestScheme, requestBaseUri string) map[string]map[string]string {
	return response
}

func (self *defaultSubject) Embedded() map[string][]Subject {
	if self.embedded == nil {
		self.embedded = make(map[string][]Subject)
	}

	return self.embedded
}

func (self *defaultSubject) Embed(k string, subject Subject) {
	if self.embedded == nil {
		self.embedded = make(map[string][]Subject)
	}
	if _, ok := self.embedded[k]; !ok {
		self.embedded[k] = make([]Subject, 1)
		self.embedded[k][0] = subject
	} else {
		self.embedded[k] = append(self.embedded[k], subject)
	}
}
