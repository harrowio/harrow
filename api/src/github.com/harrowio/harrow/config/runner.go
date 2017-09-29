package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/crypto/ssh"
)

const devOpRunnerSSHKey string = "../../config-management/.vagrant/machines/dev/virtualbox/private_key "

func (self *Config) GetSshConfig() (*ssh.ClientConfig, error) {

	var (
		privateKey ssh.Signer
		err        error
	)

	keyFilename := os.Getenv("HAR_OPERATOR_SSH_KEY_FILENAME")
	if keyFilename == "" {
		keyFilename, err = filepath.Abs(devOpRunnerSSHKey)
		if err != nil {
			return nil, fmt.Errorf("can't make absolute path from %s: %s", devOpRunnerSSHKey, err)
		}
	}

	keyMaterial, err := ioutil.ReadFile(keyFilename)
	if err != nil {
		return nil, fmt.Errorf("error reading key (%s) material: %s", keyFilename, err)
	}

	privateKey, err = ssh.ParsePrivateKey(keyMaterial)
	if err != nil {
		return nil, fmt.Errorf("error parsing key (%s) material: %s", keyFilename, err)
	}

	return &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(privateKey),
		},
		Timeout: 5 * time.Second,
	}, nil
}
