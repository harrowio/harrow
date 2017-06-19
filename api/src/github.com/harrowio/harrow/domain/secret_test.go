package domain

import (
	"reflect"
	"testing"
)

func Test_SecretType_DriverInterface(t *testing.T) {
	var sType SecretType = "test"
	value, err := sType.Value()
	if err != nil {
		t.Error(err)
	}
	if value != "test" {
		t.Errorf("value=%s, want %s", value, "test")
	}
	err = sType.Scan([]byte("new"))
	if err != nil {
		t.Error(err)
	}

	value, err = sType.Value()
	if err != nil {
		t.Error(err)
	}
	if value != "new" {
		t.Errorf("value=%s, want %s", value, "new")
	}
}

func Test_SecretStatus_DriverInterface(t *testing.T) {
	var sStatus SecretStatus = "test"
	value, err := sStatus.Value()
	if err != nil {
		t.Error(err)
	}
	if value != "test" {
		t.Errorf("value=%s, want %s", value, "test")
	}
	err = sStatus.Scan([]byte("new"))
	if err != nil {
		t.Error(err)
	}

	value, err = sStatus.Value()
	if err != nil {
		t.Error(err)
	}
	if value != "new" {
		t.Errorf("value=%s, want %s", value, "new")
	}
}

func Test_Secret_FindProject(t *testing.T) {
	secret := &Secret{
		EnvironmentUuid: "12345",
	}
	mockStore := &mockProjectStore{}
	secret.FindProject(mockStore)
	if len(mockStore.callsByEnv) != 1 {
		t.Errorf("len(callsByEnv)=%d, want %d", len(mockStore.callsByEnv), 1)
	}
	if mockStore.callsByEnv[0] != "12345" {
		t.Errorf("callsByEnv[0]=%s, want %s", mockStore.callsByEnv[0], "12345")
	}
}

func Test_Secret_OwnUrl(t *testing.T) {
	secret := &Secret{Uuid: "123"}
	url := secret.OwnUrl("https", "test.tld")
	want := "https://test.tld/secrets/123"
	if url != want {
		t.Errorf("url=%s, want %s", url, want)
	}
}

func Test_Secret_Links(t *testing.T) {
	secret := &Secret{Uuid: "456", EnvironmentUuid: "123"}
	resp := secret.Links(make(map[string]map[string]string), "https", "test.tld")
	want := map[string]map[string]string{
		"self": map[string]string{
			"href": "https://test.tld/secrets/456",
		},
		"environment": map[string]string{
			"href": "https://test.tld/environments/123",
		},
	}
	if !reflect.DeepEqual(resp, want) {
		t.Errorf("url=%#v, want %#v", resp, want)
	}
}
