package stores

import (
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/logger"

	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
)

type DiskLogStore struct {
	logDir string
	log    logger.Logger
}

func NewDiskLogStore(logDir string) LogStore {
	return &DiskLogStore{logDir: logDir}
}

func (self *DiskLogStore) Log() logger.Logger {
	if self.log == nil {
		self.log = logger.Discard
	}
	return self.log
}

func (self *DiskLogStore) SetLogger(l logger.Logger) {
	self.log = l
}

// This is not thread-safe! During operation run time, the redis log store will be used instead
func (self *DiskLogStore) FindByOperationUuid(uuid string, tepy string) (*domain.Loggable, error) {

	logFileName := self.getLogFileName(uuid, tepy)

	file, err := os.Open(logFileName)
	if err != nil {
		return nil, err
	}
	defer func() {
		file.Close()
	}()

	logLines := domain.LogLinesFromFile(file)

	log := domain.Loggable{
		Uuid:     uuid,
		Status:   domain.LoggableOK,
		LogLines: logLines,
	}

	return &log, nil
}

func (self *DiskLogStore) FindByRange(uuid string, tepy string, from, to int) (*domain.Loggable, error) {

	return nil, errors.New("DiskLogStore can't get partial logs")
}

func (self *DiskLogStore) PersistLogLine(operationUuid string, logLine *domain.LogLine) error {

	self.createAppendLogLine(operationUuid, domain.LoggableVerbose, logLine)
	if !logLine.IsInternal() {
		self.createAppendLogLine(operationUuid, domain.LoggableWorkspace, logLine)
	}
	return nil
}

func (self *DiskLogStore) OnFinished(operationUuid string) error {

	return nil // noop
}

func (self *DiskLogStore) Close() error {

	return nil // noop
}

func (self *DiskLogStore) createAppendLogLine(operationUuid, tepy string, logLine *domain.LogLine) {
	msg := logLine.Msg + "\n"

	var fileName = self.getLogFileName(operationUuid, tepy)

	if err := os.MkdirAll(filepath.Dir(fileName), 0722); err != nil {
		panic(err)
	}

	f, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0622)
	if err != nil {
		panic(err)
	}
	defer func() {
		f.Close()
	}()

	n, err := f.WriteString(msg)
	if err == nil && n < len(msg) {
		err = io.ErrShortWrite
	}
	if err != nil {
		panic(err)
	}

}

func (self *DiskLogStore) getLogFileName(operationUuid, tepy string) string {
	var pathParts = []string{self.logDir, "", "", operationUuid + "-" + tepy + ".txt"}
	fmt.Sscanf(pathParts[3], "%4s%4s", &pathParts[1], &pathParts[2])

	return path.Join(pathParts...)
}
