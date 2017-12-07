package awslib

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/harrowio/harrow/config"
)

var c *config.Config

var (
	ErrNoInstancesGive = errors.New("no instances given")
)

var ec2Service *ec2.EC2
var cwService *cloudwatch.CloudWatch
var asService *autoscaling.AutoScaling

func init() {
	c = config.GetConfig()
	ec2Service = c.GetEC2Service()
	cwService = c.GetCWService()
	asService = c.GetASService()
}

// Instead of maintaing these values as constants, we use the AS itself
// to store them.
//
// This prevents a situation in where the environments (i.e. dev and alpha)
// don't agree on the constant values and their watchdogs induce flapping on
// the shared AS.
type ASMeta struct {
	MinSize, MaxSize int                     // stored as "MinSize" and "MaxSize" of the AS
	MinFree, MaxFree int                     // stored as tags "min-free" and "max-free"
	Instances        []*autoscaling.Instance // list of *all* instances
	Healthy          []*autoscaling.Instance // list of instances that are InService and Healthy
}

// given a []autoscaling.Instance and a tag-key, filter instances that have
// the given tag-key and return a []ec2.Instance or an error
func FilterInstancesByTag(asInstances []*autoscaling.Instance, tag string) ([]*ec2.Instance, error) {
	if len(asInstances) == 0 {
		return []*ec2.Instance{}, ErrNoInstancesGive
	}
	instanceIDs := make([]*string, 0)
	for _, i := range asInstances {
		instanceIDs = append(instanceIDs, i.InstanceId)
	}
	instances := make([]*ec2.Instance, 0)
	var nextToken *string

	for {
		params := &ec2.DescribeInstancesInput{
			InstanceIds: instanceIDs,
			Filters: []*ec2.Filter{
				{Name: aws.String("tag-key"), Values: []*string{aws.String(tag)}},
			},
			NextToken: nextToken,
		}

		res, err := ec2Service.DescribeInstances(params)
		if err != nil {
			return []*ec2.Instance{}, err
		}
		for _, reservation := range res.Reservations {
			instances = append(instances, reservation.Instances...)
		}
		if res.NextToken == nil {
			// no more pages follow, break the loop
			break
		}
		nextToken = res.NextToken
	}

	return instances, nil
}

func GetTags(instanceId string) ([]*ec2.TagDescription, error) {
	input := &ec2.DescribeTagsInput{
		Filters: []*ec2.Filter{
			{Name: aws.String("resource-id"), Values: []*string{aws.String(instanceId)}},
		},
	}
	out, err := ec2Service.DescribeTags(input)
	if err != nil {
		return []*ec2.TagDescription{}, err
	}
	return out.Tags, nil
}

// add a tag to the given instance
func CreateTag(instanceId, tagKey, tagValue string) error {
	tag := &ec2.Tag{
		Key:   aws.String(tagKey),
		Value: aws.String(tagValue),
	}
	input := &ec2.CreateTagsInput{
		Resources: []*string{aws.String(instanceId)},
		Tags:      []*ec2.Tag{tag},
	}
	_, err := ec2Service.CreateTags(input)
	return err
}

// remove a tag from the given instance
func RemoveTag(instanceId, tagKey string) error {
	tag := &ec2.Tag{
		Key:   aws.String(tagKey),
		Value: aws.String(""),
	}
	input := &ec2.DeleteTagsInput{
		Resources: []*string{aws.String(instanceId)},
		Tags:      []*ec2.Tag{tag},
	}
	_, err := ec2Service.DeleteTags(input)
	return err
}

// Terminate the given instance
func TerminateInstance(instanceId string) error {
	input := &ec2.TerminateInstancesInput{
		InstanceIds: []*string{aws.String(instanceId)},
	}
	_, err := ec2Service.TerminateInstances(input)
	return err
}

func PutMetricData(metricName string, value float64) error {
	input := &cloudwatch.PutMetricDataInput{
		Namespace: aws.String("harrow"),
		MetricData: []*cloudwatch.MetricDatum{
			{
				MetricName: aws.String(metricName),
				Value:      aws.Float64(value),
			},
		},
	}
	_, err := cwService.PutMetricData(input)
	return err
}

func GetASMeta(name string) (*ASMeta, error) {
	params := &autoscaling.DescribeAutoScalingGroupsInput{
		AutoScalingGroupNames: []*string{aws.String(name)},
	}
	res, err := asService.DescribeAutoScalingGroups(params)
	if err != nil {
		return nil, err
	}
	if len(res.AutoScalingGroups) != 1 {
		return nil, fmt.Errorf("got %d AutoScalingGroups with name '%s', want 1", len(res.AutoScalingGroups))
	}
	as := res.AutoScalingGroups[0]
	asMeta := &ASMeta{
		MinSize: int(*as.MinSize),
		MaxSize: int(*as.MaxSize),
	}
	for _, tag := range as.Tags {
		if *tag.Key == "min-free" {
			v, err := strconv.ParseInt(*tag.Value, 0, 0)
			if err != nil {
				return nil, err
			}
			asMeta.MinFree = int(v)
		}
		if *tag.Key == "max-free" {
			v, err := strconv.ParseInt(*tag.Value, 0, 0)
			if err != nil {
				return nil, err
			}
			asMeta.MaxFree = int(v)
		}
	}

	asMeta.Instances = as.Instances
	asMeta.Healthy = make([]*autoscaling.Instance, 0)
	for _, i := range asMeta.Instances {
		if *i.LifecycleState == "InService" && *i.HealthStatus == "Healthy" {
			asMeta.Healthy = append(asMeta.Healthy, i)
		}
	}

	return asMeta, nil
}

func ScaleAS(name string, desired int) error {
	params := &autoscaling.UpdateAutoScalingGroupInput{
		AutoScalingGroupName: aws.String(name),
		DesiredCapacity:      aws.Int64(int64(desired)),
	}
	_, err := asService.UpdateAutoScalingGroup(params)
	return err
}

func DetachInstance(asName, instanceId string) error {
	//
	input := &autoscaling.DetachInstancesInput{
		AutoScalingGroupName:           aws.String(asName),
		InstanceIds:                    []*string{aws.String(instanceId)},
		ShouldDecrementDesiredCapacity: aws.Bool(false),
	}
	_, err := asService.DetachInstances(input)
	return err
}
