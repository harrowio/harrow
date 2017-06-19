package domain

import (
	"bufio"
	"bytes"
	"database/sql/driver"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type DeliveredRequest struct {
	*http.Request
}

func (self *DeliveredRequest) Scan(value interface{}) error {

	var source []byte
	switch input := value.(type) {
	case []byte:
		source = input
	case string:
		source = []byte(input)
	}

	req, err := http.ReadRequest(bufio.NewReader(bytes.NewBuffer(source)))
	if err != nil {
		return err
	}

	self.Request = req

	return nil
}

func (self DeliveredRequest) Value() (driver.Value, error) {
	result := new(bytes.Buffer)
	err := self.Request.Write(result)
	if err != nil {
		return nil, err
	}

	return result.Bytes(), nil
}

func (self DeliveredRequest) MarshalJSON() ([]byte, error) {
	body, err := ioutil.ReadAll(self.Body)
	if err != nil {
		return nil, err
	}

	result := struct {
		Method string      `json:"method"`
		Path   string      `json:"path"`
		Header http.Header `json:"header"`
		Body   string      `json:"body"`
	}{
		Method: self.Method,
		Path:   self.URL.Path,
		Header: self.Header,
		Body:   string(body),
	}

	marshalled, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	return marshalled, nil
}
