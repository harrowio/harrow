package domain

import "testing"

var sshRcJson = []byte(`{"PrivateKey":"priv","PublicKey":"pub"}`)

func Test_SshRepositoryCredential_AsSshRepositoryCredential(t *testing.T) {
	rc := &RepositoryCredential{
		SecretBytes: sshRcJson,
		Type:        RepositoryCredentialSsh,
	}

	sshRc, err := AsSshRepositoryCredential(rc)
	if err != nil {
		t.Fatal(err)
	}

	if sshRc.PrivateKey != "priv" {
		t.Errorf("sshRc.PrivateKey=%s, want %s", sshRc.PrivateKey, "priv")
	}

	if sshRc.PublicKey != "pub" {
		t.Errorf("sshRc.PublicKey=%s, want %s", sshRc.PublicKey, "pub")
	}
}

func Test_SshRepositoryCredential_AsRepositoryCredential_WithExisitingRC(t *testing.T) {
	sshRc := &SshRepositoryCredential{
		RepositoryCredential: &RepositoryCredential{Name: "Testing", Type: RepositoryCredentialSsh},
		PublicKey:            "pub",
		PrivateKey:           "priv",
	}
	rc, err := sshRc.AsRepositoryCredential()
	if err != nil {
		t.Fatal(err)
	}

	if rc.Name != "Testing" {
		t.Errorf("rc.Name=%s, want %s", rc.Name, "Testing")
	}
	if string(rc.SecretBytes) != string(sshRcJson) {
		t.Errorf("rc.SecretBytes=%s, want %s", rc.SecretBytes, sshRcJson)
	}
}

func Test_SshRepositoryCredential_AsRepositoryCredential_WithNewRC(t *testing.T) {
	sshRc := &SshRepositoryCredential{
		PublicKey:  "pub",
		PrivateKey: "priv",
	}
	rc, err := sshRc.AsRepositoryCredential()
	if err != nil {
		t.Fatal(err)
	}

	if !rc.IsSsh() {
		t.Error("rc should be a ssh repository credential")
	}
	if string(rc.SecretBytes) != string(sshRcJson) {
		t.Errorf("rc.SecretBytes=%s, want %s", rc.SecretBytes, sshRcJson)
	}
}
