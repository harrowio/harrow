package domain

import "testing"

func Test_RepositoryCredentialType_DriverInterface(t *testing.T) {
	var rcType RepositoryCredentialType = "test"
	value, err := rcType.Value()
	if err != nil {
		t.Error(err)
	}
	if value != "test" {
		t.Errorf("value=%s, want %s", value, "test")
	}
	err = rcType.Scan([]byte("new"))
	if err != nil {
		t.Error(err)
	}

	value, err = rcType.Value()
	if err != nil {
		t.Error(err)
	}
	if value != "new" {
		t.Errorf("value=%s, want %s", value, "new")
	}
}

func Test_RepositoryCredentialStatus_DriverInterface(t *testing.T) {
	var sStatus RepositoryCredentialStatus = "test"
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
