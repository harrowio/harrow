package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/ec2"
)

var awsConfig *aws.Config

func (c *Config) AwsConfigFilePath() string {
	return filepath.Join(c.Root, "./config/aws.json")
}

func (c *Config) GetAwsConfig() *aws.Config {

	var environmentConfigs map[string]map[string]string

	if awsConfig == nil {
		file, err := ioutil.ReadFile(c.AwsConfigFilePath())
		if os.IsNotExist(err) {
			return new(aws.Config)
		}
		if err != nil {
			panic(err)
		}
		err = json.Unmarshal(file, &environmentConfigs)
		if err != nil {
			panic(err)
		}
		region := environmentConfigs[c.Environment()]["region"]
		keyId := environmentConfigs[c.Environment()]["key_id"]
		key := environmentConfigs[c.Environment()]["key"]
		awsConfig = &aws.Config{
			Region:      aws.String(region),
			Credentials: credentials.NewStaticCredentials(keyId, key, ""),
			MaxRetries:  aws.Int(1),
		}
	}
	return awsConfig
}

func (c *Config) GetEC2Service() *ec2.EC2 {
	return ec2.New(session.New(), c.GetAwsConfig())
}

func (c *Config) GetCWService() *cloudwatch.CloudWatch {
	return cloudwatch.New(session.New(), c.GetAwsConfig())
}

func (c *Config) GetASService() *autoscaling.AutoScaling {
	return autoscaling.New(session.New(), c.GetAwsConfig())
}
