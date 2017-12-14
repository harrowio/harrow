package runner

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"

	"github.com/carlescere/goback"
	"github.com/harrowio/harrow/config"
	"github.com/harrowio/harrow/logger"
	"github.com/pkg/errors"
)

const imageName = "harrow-baseimage"

type LXD struct {
	config  *config.Config
	connURL *url.URL

	log      logger.Logger
	reporter Reporter

	containerUUID string
}

func (lxd *LXD) MakeContainer() error {

	lxd.log.Info().Msgf("making new container: %s", lxd.containerName())

	session, err := lxd.sshSession()
	if err != nil {
		return err
	}
	defer session.Close()

	cmdStr := fmt.Sprintf("lxc launch %s -p default -p docker %s", imageName, lxd.containerName())
	lxd.log.Info().Msgf("running %s", cmdStr)

	output, err := session.CombinedOutput(cmdStr)
	if err != nil {
		return errors.Wrapf(err, "error starting lxd container: %s", string(output))
	}

	if !strings.Contains(string(output), fmt.Sprintf("Starting %s", lxd.containerName())) {
		return errors.Wrap(err, "expected note about 'Starting <name>', wasn't recieved")
	}

	lxd.reporter.MadeContainer()

	return nil
}

func (lxd *LXD) DestroyContainer() error {

	lxd.log.Info().Msgf("destroying container: %s", lxd.containerName())

	b := &goback.SimpleBackoff{Min: 1 * time.Minute, Max: 10 * time.Minute, Factor: 2}

	session, err := lxd.sshSession()
	if err != nil {
		return err
	}
	defer session.Close()

	cmdStr := fmt.Sprintf("lxc delete --force %s", lxd.containerName())

	for {
		lxd.log.Info().Msgf("running %s", cmdStr)
		output, err := session.CombinedOutput(cmdStr)

		if err == nil {
			lxd.reporter.DestroyedContainer()
			return nil // container deleted successfully
		}

		if err != nil {
			if strings.Contains(string(output), "error: not found") {
				return nil // container does not exist and error was "error: not found" (exit status: 1)
			}
			if strings.Contains(string(output), "error: Could not remove LV named") {
				// container exists in LXD DB but underlaying storage has been removed, newer versions
				// of LXD handle this more gracefully
				return nil
			}
			if errBack := goback.Wait(b); errBack != nil { // goback.ErrMaxAttemptsExceeded incase we're over-time
				return errors.Wrap(err, "max attempts exceeded waiting to destroy container")
			} else {
				lxd.reporter.DestroyContainerWillRetry()
				lxd.log.Info().Msgf("error destroying container, will retry: %s", err)
			}
		}
	}
	return nil
}

func (lxd *LXD) WaitForContainerNetworking(d time.Duration) error {
	lxd.log.Info().Msgf("waiting for container %s networking (max %s)", lxd.containerName(), d)

	b := &goback.SimpleBackoff{Min: 5 * time.Second, Max: d, Factor: 2}
	for {
		err := lxd.CheckContainerNetworking()
		if err != nil {
			if errBack := goback.Wait(b); errBack != nil { // goback.ErrMaxAttemptsExceeded incase we're over-time
				return errors.Wrap(err, "max attempts exceeded waiting for networking to come up")
			} else {
				lxd.log.Info().Msgf("error waiting for container networking, will retry: %s", err)
			}
		}
		break
	}
	lxd.log.Info().Msgf("container networking for %s is up!", lxd.containerName())
	return nil
}

func (lxd *LXD) ContainerExists() (bool, error) {

	lxd.log.Debug().Msgf("checking if container %s exists", lxd.containerName())

	session, err := lxd.sshSession()
	if err != nil {
		return false, errors.Wrap(err, "error getting session to check container existence")
	}
	defer session.Close()

	cmdStr := fmt.Sprintf("lxc info %s", lxd.containerName())

	output, err := session.CombinedOutput(cmdStr)
	if err == nil {
		return true, nil // container exists, we could get it's info
	}
	if err != nil {
		if strings.Contains(string(output), "error: not found") {
			return false, nil // container does not exist and error was "error: not found" (exit status: 1)
		}
		return false, errors.Wrapf(err, "error calling lxd info %s: %s", lxd.containerName(), output)
	}
	return false, nil
}

func (lxd *LXD) MaintainConnection(lost chan<- error) {
	lxd.log.Info().Msg("opening long lived connection to the container")

	session, err := lxd.sshSession()
	if err != nil {
		lost <- errors.Wrap(err, "unable to open ssh session")
	}
	defer session.Close()

	cmdStr := fmt.Sprintf("lxc exec %s -- cat", lxd.containerName())
	lxd.log.Info().Msgf("running %s", cmdStr)

	// Set up terminal modes
	modes := ssh.TerminalModes{
		ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	// Request pseudo terminal
	if err := session.RequestPty("xterm", 40, 80, modes); err != nil {
		lost <- errors.Wrapf(err, "upstream server rejected requets for a pty: %s", err)
	}

	err = session.Run(cmdStr)
	lost <- errors.Wrapf(err, "error maintaining long running command on container: %s", err)

}

func (lxd *LXD) CheckContainerNetworking() error {

	lxd.log.Info().Msg("checking for container networking")

	session, err := lxd.sshSession()
	if err != nil {
		return err
	}
	defer session.Close()

	cmdStr := fmt.Sprintf("lxc list %s --format json", lxd.containerName())
	lxd.log.Info().Msgf("running %s", cmdStr)

	output, err := session.CombinedOutput(cmdStr)
	if err != nil {
		return errors.Wrapf(err, "error getting container info from lxc: ", string(output))
	}

	// This messy struct maps to the LXC struct. We could import LXC's go package
	// to get this, but then we'd have a dependency on something the size of the
	// universe. #notworthit
	containerInfo := []struct {
		State struct {
			Network struct {
				Eth0 struct {
					Addresses []struct {
						Family  string
						Address string
					}
				}
			}
		}
	}{}

	if err := json.Unmarshal(output, &containerInfo); err != nil {
		return errors.Wrap(err, "error unmarshalling the container information from lxc")
	}

	for _, info := range containerInfo {
		for _, addr := range info.State.Network.Eth0.Addresses {
			if addr.Family == "inet" {
				return nil
			}
		}
	}

	return fmt.Errorf("no inet device found on container %s", lxd.containerName())
}

func (lxd *LXD) sshClient() (*ssh.Client, error) {

	b := &goback.SimpleBackoff{Min: 1 * time.Second, Max: 60 * time.Minute, Factor: 2}

	sshConfig, err := lxd.config.GetSshConfig()
	if err != nil {
		return nil, errors.Wrap(err, "could not get ssh configuration")
	}
	for {
		client, err := ssh.Dial("tcp", lxd.connURL.Host, sshConfig)
		if err != nil {
			lxd.reporter.SSHError(err)
			if errBack := goback.Wait(b); errBack != nil { // goback.ErrMaxAttemptsExceeded incase we're over-time
				return nil, errors.Wrap(err, "max attempts exceeded waiting to dial ssh to host")
			} else {
				lxd.log.Info().Msgf("error dialing ssh connection, will retry: %s", err)
			}
		} else {
			return client, nil
		}
	}
}

func (lxd *LXD) sshSession() (*ssh.Session, error) {

	b := &goback.SimpleBackoff{Min: 1 * time.Second, Max: 60 * time.Minute, Factor: 2}

	for {
		client, err := lxd.sshClient()
		if err != nil {
			return nil, errors.Wrap(err, "could not get ssh client")
		}
		session, err := client.NewSession()
		if err != nil {
			lxd.reporter.SSHError(err)
			if errBack := goback.Wait(b); errBack != nil { // goback.ErrMaxAttemptsExceeded incase we're over-time
				return nil, errors.Wrap(err, "max attempts exceeded waiting to start ssh session")
			} else {
				lxd.log.Info().Msgf("error starting ssh session, will retry: %s", err)
			}
		} else {
			return session, nil
		}
	}
}

func (lxd *LXD) containerName() string {
	return fmt.Sprintf("%s-%s", lxd.prefix(), lxd.containerUUID)
}

func (lxd *LXD) prefix() string {
	return "container"
}
