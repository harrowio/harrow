package domain

import "testing"

var sshSecretJson = []byte(`{"PrivateKey":"priv","PublicKey":"pub"}`)

func Test_SshSecret_AsSshSecret(t *testing.T) {
	s := &Secret{
		SecretBytes: sshSecretJson,
		Type:        SecretSsh,
	}

	sshS, err := AsSshSecret(s)
	if err != nil {
		t.Fatal(err)
	}

	if sshS.PrivateKey != "priv" {
		t.Errorf("sshS.PrivateKey=%s, want %s", sshS.PrivateKey, "priv")
	}

	if sshS.PublicKey != "pub" {
		t.Errorf("sshS.PublicKey=%s, want %s", sshS.PublicKey, "pub")
	}
}

func Test_SshSecret_AsSecret_WithExistingSecret(t *testing.T) {
	sshS := &SshSecret{
		Secret:     &Secret{Name: "Testing", Type: SecretSsh},
		PublicKey:  "pub",
		PrivateKey: "priv",
	}
	s, err := sshS.AsSecret()
	if err != nil {
		t.Fatal(err)
	}

	if s.Name != "Testing" {
		t.Errorf("s.Name=%s, want %s", s.Name, "Testing")
	}
	if string(s.SecretBytes) != string(sshSecretJson) {
		t.Errorf("s.SecretBytes=%s, want %s", s.SecretBytes, sshSecretJson)
	}
}

func Test_SshSecret_AsSecret_WithNewSecret(t *testing.T) {
	sshS := &SshSecret{
		PublicKey:  "pub",
		PrivateKey: "priv",
	}
	s, err := sshS.AsSecret()
	if err != nil {
		t.Fatal(err)
	}

	if !s.IsSsh() {
		t.Error("s should be a ssh secret")
	}
	if string(s.SecretBytes) != string(sshSecretJson) {
		t.Errorf("s.SecretBytes=%s, want %s", s.SecretBytes, sshSecretJson)
	}
}

func Test_SshSecret_AsUnAndPrivileged(t *testing.T) {
	sshS := &SshSecret{
		PublicKey:  "pub",
		PrivateKey: "priv",
	}

	u, err := sshS.AsUnprivileged()
	if err != nil {
		t.Fatal(err)
	}
	if !u.IsSsh() {
		t.Error("u should be a ssh secret")
	}
	if u.PublicKey != "pub" {
		t.Errorf("u.PublicKey=%s, want %s", u.PublicKey, "pub")
	}

	p, err := sshS.AsPrivileged()
	if err != nil {
		t.Fatal(err)
	}
	if !p.IsSsh() {
		t.Error("p should be a ssh secret")
	}
	if p.PublicKey != "pub" {
		t.Errorf("p.PublicKey=%s, want %s", p.PublicKey, "pub")
	}
	if p.PrivateKey != "priv" {
		t.Errorf("p.PrivateKey=%s, want %s", p.PrivateKey, "priv")
	}
}
