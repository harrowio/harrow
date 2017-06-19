package sshkey

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
)

func Generate(name, tepy string, length int) (string, string, error) {
	tmpdir, err := ioutil.TempDir("", "sshkey")
	if err != nil {
		return "", "", err
	}
	defer func() {
		err := os.RemoveAll(tmpdir)
		if err != nil {
			panic(err)
		}
	}()

	keyfileName := fmt.Sprintf("%s/key", tmpdir)

	cmd := exec.Command("/usr/bin/ssh-keygen",
		"-C", name,
		"-t", tepy,
		"-b", fmt.Sprintf("%d", length),
		"-f", keyfileName,
		"-N", "",
	)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return "", "", err
	}
	private, err := readKey(keyfileName)
	if err != nil {
		return "", "", err
	}

	public, err := readKey(keyfileName + ".pub")
	if err != nil {
		return "", "", err
	}

	return private, public, nil
}

func readKey(keyfileName string) (string, error) {
	keyfile, err := os.Open(keyfileName)
	if err != nil {
		return "", err
	}
	defer func() {
		keyfile.Close()
	}()
	keybytes, err := ioutil.ReadAll(keyfile)
	if err != nil {
		return "", err
	}
	return string(keybytes), nil
}
