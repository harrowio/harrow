package operationRunner

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/harrowio/harrow/config"
	"github.com/harrowio/harrow/logger"
	"github.com/harrowio/harrow/uuidhelper"
	"golang.org/x/crypto/ssh"
)

type LXDAcquisition struct {
	ConnectTo *url.URL
	BaseImage string
	log       logger.Logger
}

func (lxd *LXDAcquisition) ContainerPrefix() string {
	pathComponents := strings.Split(lxd.ConnectTo.Path, "/")
	if len(pathComponents) <= 2 {
		return "machine"
	}

	return pathComponents[1]
}

func (lxd *LXDAcquisition) MustTakeInstance(c *config.Config) (string, string, error) {
	lxdHost := lxd.ConnectTo.Host
	sshConfig, err := c.GetSshConfig()
	if err != nil {
		return "", "", err
	}
	sshConfig.User = lxd.ConnectTo.User.Username()
	client, err := ssh.Dial("tcp", lxdHost, sshConfig)
	if err != nil {
		return "", "", err
	}

	session, err := client.NewSession()
	if err != nil {
		return "", "", err
	}
	defer session.Close()

	containerId := lxd.ContainerPrefix() + "-" + uuidhelper.MustNewV4()
	lxd.log.Info().Msgf("launching LXD container %s as %s on %s", containerId, lxd.ConnectTo.User.Username(), lxd.ConnectTo.Host)
	output, remoteErr := session.CombinedOutput("lxc launch " + lxd.BaseImage + " -p default -p docker " + containerId)
	if remoteErr != nil {
		lxd.log.Error().Msgf("error starting lxd container %s: %s\n%s\n", containerId, remoteErr, output)
		return containerId, lxdHost, remoteErr
	}
	if err := lxd.waitForContainerNetworking(client, containerId); err != nil {
		lxd.log.Error().Msgf("error waiting for lxd container networking %s: %s\n", containerId, err)
		return containerId, lxdHost, err
	}

	return containerId, lxdHost, nil
}

func (lxd *LXDAcquisition) waitForContainerNetworking(client *ssh.Client, containerId string) error {
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

	networkIsUp := func() (bool, error) {
		session, err := client.NewSession()
		if err != nil {
			return false, err
		}
		output, err := session.CombinedOutput("lxc list --format=json " + containerId)
		if err != nil {
			return false, err
		}

		if err := json.Unmarshal(output, &containerInfo); err != nil {
			return false, err
		}

		for _, info := range containerInfo {
			for _, addr := range info.State.Network.Eth0.Addresses {
				if addr.Family == "inet" {
					return true, nil
				}
			}
		}
		return false, nil
	}

	i := 1
	for ; i < 10; i++ {
		isUp, err := networkIsUp()
		if err != nil {
			lxd.log.Error().Msgf("container %s: Error checking network: %s", containerId, err)
		}
		if isUp {
			return nil
		}
		time.Sleep(time.Duration(i) * time.Second)
	}

	return fmt.Errorf("waitForContainerNetworking: timed out after %d tries", i)
}
