package limits

import (
	"context"
	"fmt"

	"github.com/harrowio/harrow/config"
	"github.com/harrowio/harrow/domain"
	"github.com/harrowio/harrow/logger"
	"github.com/harrowio/harrow/protos"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

func NewDefaultClient(c *config.Config) *Client {
	limitsConfig := c.LimitsStoreConfig()
	return &Client{
		bind:     limitsConfig.Bind,
		port:     limitsConfig.Port,
		failmode: limitsConfig.FailMode,
		enabled:  limitsConfig.Enabled,
	}
}

type Client struct {
	bind     string
	port     string
	failmode string

	enabled bool

	logger logger.Logger
}

func (c *Client) OrganizationLimitsExceeded(org *domain.Organization) (bool, error) {

	if !c.enabled {
		c.log().Info().Msg("limits service disabled, allowing")
		return false, nil
	}

	conn, err := c.grpc_client()
	if err != nil {
		if c.failmode == "assume_allowed" {
			c.log().Info().Msg("can't dial grpc limits service, failmode is assume_allowed, allowing")
			return false, err
		}
	} else {
		defer conn.Close()
	}

	client := protos.NewLimitsServiceClient(conn)

	orgKey := protos.OrganizationKey{org.Uuid}
	orgLimitsExceededRes, err := client.Exceeded(context.Background(), &orgKey)
	if err != nil {
		if c.failmode == "assume_allowed" {
			c.log().Info().Msg("can't dial grpc limits service, failmode is assume_allowed, allowing")
			return false, nil
		}
		return true, errors.Wrap(err, fmt.Sprintf("calling LimitsServiceClient.Exceeded(%s)", orgKey))
	}

	return orgLimitsExceededRes.Exceeded, err
}

func (c *Client) ForOrganizationUuid(uuid string) (*domain.Limits, error) {

	if !c.enabled {
		c.log().Info().Msg("limits service disabled, allowing")
		return &domain.Limits{}, nil
	}

	conn, err := c.grpc_client()
	if err != nil {
		if c.failmode == "assume_allowed" {
			c.log().Info().Msg("can't dial grpc limits service, failmode is assume_allowed, returning nil limits object")
			return &domain.Limits{}, err
		}
	} else {
		defer conn.Close()
	}

	client := protos.NewLimitsServiceClient(conn)

	orgKey := protos.OrganizationKey{Uuid: uuid}
	orgLimitsRes, err := client.ForOrganization(context.Background(), &orgKey)
	if err != nil {
		if c.failmode == "assume_allowed" {
			c.log().Info().Msg("can't dial grpc limits service, failmode is assume_allowed, returning nil limits object")
			return &domain.Limits{}, nil
		}
		return &domain.Limits{}, errors.Wrap(err, fmt.Sprintf("calling LimitsServiceClient.ForOrganization(%s)", orgKey))
	}

	return &domain.Limits{
		Projects:            int(orgLimitsRes.Projects),
		Members:             int(orgLimitsRes.Members),
		PrivateRepositories: int(orgLimitsRes.PrivateRepositories),
		PublicRepositories:  int(orgLimitsRes.PublicRepositories),
		TrialDaysLeft:       int(orgLimitsRes.TrialDaysLeft),
		TrialEnabled:        orgLimitsRes.TrialEnabled,
	}, nil
}

func (c *Client) grpc_client() (*grpc.ClientConn, error) {
	fmt.Println("About to dial gRPC")
	conn, err := grpc.Dial(fmt.Sprintf("%s:%s", c.bind, c.port), grpc.WithInsecure())
	if err != nil {
		return nil, errors.Wrap(err, "dialing grpc service for limits.Client")
	}
	return conn, err
}

func (c *Client) log() logger.Logger {
	if c.logger == nil {
		c.logger = logger.Discard
	}
	return c.logger
}

func (c *Client) SetLogger(l logger.Logger) {
	c.logger = l
}
