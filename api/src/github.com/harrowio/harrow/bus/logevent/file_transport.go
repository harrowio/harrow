package logevent

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/harrowio/harrow/config"
	"github.com/harrowio/harrow/logger"
	"github.com/harrowio/harrow/loxer"
)

var fileNameTpl = "%s/%s.json.gz"

type FileTransport struct {
	config *config.Config
	log    logger.Logger
}

func NewFileTransport(c *config.Config, log logger.Logger) *FileTransport {
	return &FileTransport{
		config: c,
		log:    log,
	}
}

func (self *FileTransport) Consume(operationUUID string) (<-chan *Message, error) {
	f, err := os.Open(fmt.Sprintf(fileNameTpl, self.config.FilesystemConfig().OpLogDir, operationUUID))
	if err != nil {
		return nil, err
	}
	gz, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}
	res := make(chan *Message)
	scanner := bufio.NewScanner(gz)
	go func() {
		defer func() {
			err := gz.Close()
			if err != nil {
				self.log.Error().Msgf("gz.close(): %s", err)
			}
			err = f.Close()
			if err != nil {
				self.log.Error().Msgf("f.close(): %s", err)
			}
			close(res)
		}()
		for scanner.Scan() {
			msg := new(Message)
			msg.O = operationUUID
			msg.E = loxer.SerializedEvent{}
			peek := map[string]interface{}{}
			if err := json.Unmarshal(scanner.Bytes(), &peek); err != nil {
				self.log.Error().Msgf("json.unmarshal(%s): %s", scanner.Text(), err)
				continue
			}

			var dest interface{} = &msg

			if _, found := peek["type"]; found {
				dest = &msg.E
			} else {
				dest = &msg
			}

			err := json.Unmarshal(scanner.Bytes(), dest)
			if err != nil {
				self.log.Error().Msgf("json.unmarshal(%s): %s", scanner.Text(), err)
				return
			}

			res <- msg
		}
		if err := scanner.Err(); err != nil {
			self.log.Error().Msgf("scanner.err(): %s", err)
			return
		}
	}()
	return res, nil
}

func (self *FileTransport) Close() error {
	return nil
}

func (self *FileTransport) WriteLexemes(operationUUID string, lexemes []*Message) error {
	tmp, err := ioutil.TempDir(self.config.FilesystemConfig().OpLogDir, "log_transport")
	if err != nil {
		return fmt.Errorf("ioutil.TempDir(\"\", \"log_transport\"): %s", err)
	}
	defer os.RemoveAll(tmp)
	tmpFileName := fmt.Sprintf(fileNameTpl, tmp, operationUUID)
	f, err := os.Create(tmpFileName)
	if err != nil {
		return fmt.Errorf("os.Create(%s): %s", tmpFileName, err)
	}
	defer f.Close()
	gz := gzip.NewWriter(f)
	defer gz.Close()

	for _, l := range lexemes {
		pkt, err := json.Marshal(l)
		if err != nil {
			return fmt.Errorf("json.Marshal(%#v): %s", l, err)
		}
		_, err = gz.Write(pkt)
		if err != nil {
			return fmt.Errorf("f.Write(%s): %s", pkt, err)
		}
		_, err = gz.Write([]byte("\n"))
		if err != nil {
			return fmt.Errorf("gz.Write(%q): %s", "\n", err)
		}
	}
	err = gz.Close()
	if err != nil {
		return fmt.Errorf("gz.Close(): %s", err)
	}
	err = f.Close()
	if err != nil {
		return fmt.Errorf("f.Close(): %s", err)
	}
	fileName := fmt.Sprintf(fileNameTpl, self.config.FilesystemConfig().OpLogDir, operationUUID)
	err = os.Rename(tmpFileName, fileName)
	if err != nil {
		return fmt.Errorf("os.Rename(%q, %q): %s", tmpFileName, fileName, err)
	}
	return nil
}
