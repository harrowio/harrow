package git

import (
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
)

type OperatingSystem struct {
	tempDir string
}

func NewOperatingSystem(tempDir string) *OperatingSystem {
	return &OperatingSystem{
		tempDir: tempDir,
	}
}

func (self *OperatingSystem) TempDir() (string, error) {
	return ioutil.TempDir(self.tempDir, "git-clone")
}

func (self *OperatingSystem) PersistentDir(key string) (string, error) {
	dirname := filepath.Join(self.tempDir, key)
	if err := os.MkdirAll(dirname, 0755); err != nil {
		if !os.IsExist(err) {
			return dirname, err
		}
	}
	return dirname, nil
}

func (self *OperatingSystem) DeleteFile(filename string) error {
	return os.RemoveAll(filename)
}

func (self *OperatingSystem) CreateFile(filename string) (io.WriteCloser, error) {
	return os.Create(filename)
}

func (self *OperatingSystem) SetPermissions(filename string, mode int) error {
	return os.Chmod(filename, os.FileMode(mode))
}

func (self *OperatingSystem) Run(cmd *SystemCommand) ([]byte, error) {
	osCommand := exec.Command(cmd.Exec, cmd.Args...)
	osCommand.Dir = cmd.Dir
	osCommand.Env = cmd.Env
	return osCommand.CombinedOutput()
}
