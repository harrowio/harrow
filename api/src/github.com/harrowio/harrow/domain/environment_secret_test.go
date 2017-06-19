package domain

import "testing"

var envSecretJson = []byte(`{"Value":"foo"}`)

func Test_EnvironmentSecret_AsEnvironmentSecret(t *testing.T) {
	s := &Secret{
		SecretBytes: envSecretJson,
		Type:        SecretEnv,
	}
	eS, err := AsEnvironmentSecret(s)
	if err != nil {
		t.Fatal(err)
	}

	if have, want := eS.Value, "foo"; have != want {
		t.Fatalf("eS.Value == %s, want %s", have, want)
	}

}

func Test_EnvironmentSecret_AsSecret_WithExistingSecret(t *testing.T) {
	eS := &EnvironmentSecret{
		Secret: &Secret{Name: "Testing", Type: SecretEnv},
		Value:  "foo",
	}
	s, err := eS.AsSecret()
	if err != nil {
		t.Fatal(err)
	}

	if have, want := s.Name, "Testing"; have != want {
		t.Errorf(`s.Name == "%s", want "%s"`, have, want)
	}
	if have, want := string(s.SecretBytes), string(envSecretJson); have != want {
		t.Errorf(`s.SecretBytes == "%s", want "%s"`, have, want)
	}
}

func Test_EnvironmentSecret_AsSecret_WithNewSecret(t *testing.T) {
	eS := &EnvironmentSecret{
		Value: "foo",
	}
	s, err := eS.AsSecret()
	if err != nil {
		t.Fatal(err)
	}

	if !s.IsEnv() {
		t.Error("s should be a env secret")
	}
	if have, want := string(s.SecretBytes), string(envSecretJson); have != want {
		t.Errorf(`s.SecretBytes == "%s", want "%s"`, have, want)
	}
}
