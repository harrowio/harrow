package http

import (
	"io"

	"github.com/harrowio/harrow/domain"

	"encoding/json"

	"strconv"
)

type Links map[string]map[string]string

type halWrapper struct {
	Subject  interface{}                  `json:"subject"`
	Links    map[string]map[string]string `json:"_links"`
	Embedded map[string][]*halWrapper     `json:"_embedded"`
}

type halCollectionWrapper struct {
	Collection []interface{}                `json:"collection"`
	Meta       map[string]string            `json:"_meta"`
	Links      map[string]map[string]string `json:"_links"`
}

func decodeHALParams(r io.Reader, params interface{}) error {
	dec := json.NewDecoder(r)
	wrapper := &halWrapper{Subject: params}
	return dec.Decode(wrapper)
}

func mergeMeta(ctxt RequestContext, subject domain.Subject) *halWrapper {
	return &halWrapper{
		Subject:  subject,
		Links:    linksForSubject(ctxt.Auth(), ctxt.R(), subject),
		Embedded: embed(ctxt, subject.Embedded()),
	}
}

func marshalMergeMeta(ctxt RequestContext, subject domain.Subject) ([]byte, error) {
	return json.MarshalIndent(mergeMeta(ctxt, subject), "", "  ")
}

func embed(ctxt RequestContext, embedded map[string][]domain.Subject) map[string][]*halWrapper {
	res := make(map[string][]*halWrapper)
	if embedded == nil {
		return res
	}
	for k, c := range embedded {
		res[k] = make([]*halWrapper, len(embedded[k]))
		for idx, item := range c {
			res[k][idx] = mergeMeta(ctxt, item)
		}
	}
	return res
}

func writeAsJson(c RequestContext, subject domain.Subject) {
	if o, err := marshalMergeMeta(c, subject); err == nil {
		c.W().Header().Set("Content-Type", "application/json")
		c.W().Write(applyFilterTemplate(c.R(), subject, o))
	} else {
		handleError(c.W(), NewInternalError(err))
	}
}

type CollectionPage struct {
	Total      int
	Count      int
	Collection []interface{}
}

func writeCollectionPageAsJson(c RequestContext, page *CollectionPage) {

	var collectionWithMeta []interface{} = make([]interface{}, page.Count)
	for i, s := range page.Collection {
		collectionWithMeta[i] = mergeMeta(c, s.(domain.Subject))
	}

	halCollection := &halCollectionWrapper{
		Collection: collectionWithMeta,
		Meta: map[string]string{
			"total": strconv.Itoa(page.Total),
			"count": strconv.Itoa(page.Count),
		},
		Links: map[string]map[string]string{},
	}

	if o, err := json.MarshalIndent(halCollection, "", "  "); err == nil {
		c.W().Header().Set("Content-Type", "application/json")
		c.W().Write(o)
	} else {
		handleError(c.W(), NewInternalError(err))
	}
}
