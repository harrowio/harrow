package git

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"

	crypto "golang.org/x/crypto/ssh"
)

var (
	sshWrapperFormatTpl = "%s -o IdentitiesOnly=yes -o PasswordAuthentication=no -o StrictHostKeyChecking=no -i %s \"$@\"\n"
)

type Credential interface {
	String() string
	Username() string
	Secret() string
	Protocol() string
}

type httpCredential struct {
	User *url.Userinfo
}

func NewHTTPCredential(username, password string) (*httpCredential, error) {
	return &httpCredential{User: url.UserPassword(username, password)}, nil
}

func (self *httpCredential) Protocol() string { return "https" }

func (self *httpCredential) Username() string {
	if self.User == nil {
		return ""
	} else {
		return self.User.Username()
	}
}

func (self *httpCredential) Secret() string { return self.Password() }

func (self *httpCredential) Password() string {
	if self.User == nil {
		return ""
	} else {
		p, _ := self.User.Password()
		return p
	}
}

func (self *httpCredential) String() string {
	return fmt.Sprintf("httpCredential: %s:%s", self.Username(), self.Password())
}

type sshCredential struct {
	User    *url.Userinfo
	Signer  crypto.Signer
	keydata []byte
}

func (self *sshCredential) Protocol() string { return "ssh" }

func (self *sshCredential) Username() string {
	if self.User == nil {
		return ""
	} else {
		return self.User.Username()
	}
}

func (self *sshCredential) Secret() string {
	return string(self.keydata)
}

func NewSshCredential(username string, keydata []byte) (*sshCredential, error) {
	s, err := crypto.ParsePrivateKey(keydata)
	if err != nil {
		return nil, ErrUnparsablePrivateKey
	}
	return &sshCredential{Signer: s, User: url.User(username), keydata: keydata}, nil
}

func (self *sshCredential) String() string {
	return fmt.Sprintf("sshCredential: username:%s SHA256:%s", self.Username(), self.FingerprintSHA256())
}

type sensitiveTempFiles struct {
	wrapper    string
	privateKey string
}

func (self *sensitiveTempFiles) Remove() {
	os.Remove(self.privateKey)
	os.Remove(self.wrapper)
}

// writeSSHWrapperToDisk writes an SSH wrapper to disk. It writes the a wrapper
// that calls SSH with options to ignore known hosts, specify a key, etc see
// package var sshWrapperFormatTpl
func (self *sshCredential) writeWrapperScriptToDisk(sshExecutableAbsPath string) (*sensitiveTempFiles, error) {
	privateKeyTempfileName, err := self.writePrivateKeyToDisk()
	if err != nil {
		return nil, ErrWritingSshPrivateKeyToDisk
	}

	sensitiveFiles := &sensitiveTempFiles{privateKey: privateKeyTempfileName}
	file, err := ioutil.TempFile("", "sshWrapper")
	if err != nil {
		return nil, err
	}
	sensitiveFiles.wrapper = file.Name()

	if _, err := file.WriteString(fmt.Sprintf(sshWrapperFormatTpl, sshExecutableAbsPath, privateKeyTempfileName)); err != nil {
		return nil, err
	}
	if err := os.Chmod(file.Name(), 0700); err != nil {
		return nil, err
	}
	if err := file.Close(); err != nil {
		return nil, err
	}
	return sensitiveFiles, nil
}

// writePrivateKeyToDisk writes the private key bytes to disk in a
// tempfile. It returns the path to the file on disk. It is the
// responsibility of the caller to remove this file.
func (self *sshCredential) writePrivateKeyToDisk() (string, error) {
	file, err := ioutil.TempFile("", "sshCredential")
	if err != nil {
		return "", err
	}
	if _, err := file.Write(self.keydata); err != nil {
		return "", err
	}
	if err := os.Chmod(file.Name(), 0600); err != nil {
		return "", err
	}
	if err := file.Close(); err != nil {
		return "", err
	}
	return file.Name(), nil
}

// FingerprintSHA256 returns the SHA256 fingerprint of the public key,
// the output is intended to match the output of `$ ssh-keygen -l -f
// ./some-public.key` however this implementation (harmlessly)
// incorrectly pads the otuput with `=`.
func (self *sshCredential) FingerprintSHA256() string {
	hasher := sha256.New()
	hasher.Write(self.Signer.PublicKey().Marshal())
	return base64.StdEncoding.EncodeToString(hasher.Sum(nil))
}
