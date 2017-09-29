package controllerLXD

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/harrowio/harrow/logger"

	"golang.org/x/crypto/ssh"
)

type userScriptUploader struct {
	log logger.Logger
}

func (usu *userScriptUploader) uploadUserScript(client *ssh.Client, containerId string) error {

	session, err := usu.newSession(client)
	if err != nil {
		return err
	}
	defer session.Close()

	usu.log.Debug().Msg("got ssh session on server")
	errors := make(chan error)
	go func() {
		usu.log.Debug().Msg("connecting my stdin to the session's stdinpipe")
		w, err := session.StdinPipe()
		if err != nil {
			errors <- err
			return
		}
		usu.log.Debug().Msg("beore iocopy")
		io.Copy(w, os.Stdin)
		usu.log.Debug().Msg("after iocopy")
		w.Close()
		usu.log.Debug().Msg("after close")
		errors <- nil
	}()

	usu.log.Debug().Msg("starting untar command on container")
	tarCmd := fmt.Sprintf("lxc exec %s -- sudo -u ubuntu -i tar -xzf -", containerId)
	if containerId == "" {
		tarCmd = "sudo -u ubuntu -i tar -xzf -"
	}
	usu.log.Debug().Msgf("starting %s on container", tarCmd)
	err = session.Run(tarCmd)
	if err != nil {
		return err
	}
	usu.log.Debug().Msgf("finished upload")

	return <-errors
}

func (usu *userScriptUploader) newSession(client *ssh.Client) (*ssh.Session, error) {
	session, err := client.NewSession()
	if err != nil {
		return nil, err
	}
	err = usu.traceSsh(session)
	if err != nil {
		return nil, err
	}

	return session, nil
}

func (usu *userScriptUploader) traceSsh(session *ssh.Session) error {
	stdout, err := session.StdoutPipe()
	if err != nil {
		return err
	}

	stderr, err := session.StderrPipe()
	if err != nil {
		return err
	}
	go func() {
		lines := bufio.NewScanner(stderr)
		for lines.Scan() {
			usu.log.Warn().Msgf("%s", lines.Text())
		}
	}()
	go func() {
		lines := bufio.NewScanner(stdout)
		for lines.Scan() {
			usu.log.Debug().Msg(lines.Text())
		}
	}()

	return nil
}
