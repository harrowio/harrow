package domain

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

func Test_DeliveredRequest_Scan_readsHTTPRequest(t *testing.T) {
	req := &DeliveredRequest{}
	txt := []byte(strings.Join([]string{
		"POST /foo HTTP/1.1",
		"Host: example.com",
		"Content-Length: 4",
		"",
		"BAR\n",
	}, "\r\n"))

	err := req.Scan(txt)
	if err != nil {
		t.Fatal(err)
	}

	if req.URL.Path != "/foo" {
		t.Errorf("Unexpected path: %q", req.URL.Path)
	}

	if req.Method != "POST" {
		t.Errorf("Unexpected method: %q", req.Method)
	}

	if host := req.Host; host != "example.com" {
		t.Errorf("Unexpected host: %q", host)
	}

	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		t.Fatal(err)
	}

	if string(data) != "BAR\n" {
		t.Errorf("Unexpected body: %q", data)
	}
}

func Test_DeliveredRequest_Value_writesHTTPRequest(t *testing.T) {
	httpReq, err := http.NewRequest("POST", "http://example.com/foo", bytes.NewBufferString("BAR\n"))
	if err != nil {
		t.Fatal(err)
	}
	httpReq.Header.Set("User-Agent", "test")

	req := &DeliveredRequest{
		Request: httpReq,
	}

	value, err := req.Value()
	if err != nil {
		t.Fatal(err)
	}

	raw, ok := value.([]byte)
	if !ok {
		t.Fatalf("Expected value to be %T, got %T", []byte{}, value)
	}

	expected := strings.Join([]string{
		"POST /foo HTTP/1.1",
		"Host: example.com",
		"User-Agent: test",
		"Content-Length: 4",
		"",
		"BAR\n",
	}, "\r\n")

	if string(raw) != expected {
		t.Fatalf("Expected:\n%s\nGot:\n%s\n", expected, raw)
	}
}

func Test_DeliveredRequest_MarshalJSON_includesBody(t *testing.T) {
	httpReq, err := http.NewRequest("POST", "http://example.com/foo", bytes.NewBufferString("BAR\n"))
	if err != nil {
		t.Fatal(err)
	}
	httpReq.Header.Set("User-Agent", "test")

	req := &DeliveredRequest{
		Request: httpReq,
	}

	marshalled, err := req.MarshalJSON()
	result := struct {
		Body string
	}{}

	if err := json.Unmarshal(marshalled, &result); err != nil {
		t.Fatal(err)
	}

	if result.Body != "BAR\n" {
		t.Fatalf("Expected body to be %q, got %q", "BAR\n", result.Body)
	}
}
