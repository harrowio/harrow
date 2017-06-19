package vmexLXD

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"time"

	"golang.org/x/crypto/ssh"

	"github.com/harrowio/harrow/config"
	"github.com/rs/zerolog"
)

const ProgramName = "vmex-lxd"

var log zerolog.Logger = zerolog.New(os.Stdout).With().Str("harrow", ProgramName).Timestamp().Logger()

func Main() {
	interval := 2 * time.Second
	containerId := flag.String("container-id", "", "Id of the container to remove")
	connectTo := flag.String("connect", "", "Host running the LXD container")

	flag.Parse()
	connectionInfo, err := url.Parse(*connectTo)
	if err != nil {
		panic(err)
	}

	host := connectionInfo.Host
	if *containerId == "" || host == "" {
		flag.Usage()
		os.Exit(2)
	}

	deadline := time.After(config.InstanceDeadline)
	go func() {
		<-deadline
		log.Error().Msgf("Deadline reached trying to kill instance %s, exiting", *containerId)
		// sysexits.h: #define EX_TEMPFAIL	75	/* temp failure; user is invited to retry */
		os.Exit(75)
	}()

	log.Info().Msgf("Terminating container %s", *containerId)
	addr := fmt.Sprintf("%s:22", host)
	conf, err := config.GetConfig().GetSshConfig()
	conf.User = connectionInfo.User.Username()
	if conf.User == "" {
		conf.User = "root"
	}

	if err != nil {
		log.Fatal().Msgf("Unable to get ssh config: %s", err)
	}
	client, err := ssh.Dial("tcp", addr, conf)
	if err != nil {
		log.Fatal().Msgf("Unable to open ssh connection: %s", err)
	}
	defer client.Close()

	for tries, try := 20, 0; try < tries; try++ {
		session, err := client.NewSession()
		if err != nil {
			log.Error().Msgf("try %d/%d: unable to start SSH session: %s", try+1, tries, err)
			time.Sleep(interval)
			continue
		}

		if err := session.Run("lxc delete --force " + *containerId); err != nil {
			log.Error().Msgf("try %d/%d: failed to delete container %s: %s", try+1, tries, *containerId, err)
			interval = interval * 2
			log.Debug().Msgf("Sleeping for %v", interval)
			time.Sleep(interval)

			continue
		} else {
			log.Info().Msgf("%s removed", *containerId)
			os.Exit(0)
		}
	}

	log.Error().Msgf("Failed to remove %s (tried 20 times)", *containerId)
}
