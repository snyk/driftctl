package aws_test

import (
	"testing"

	"github.com/snyk/driftctl/test"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/snyk/driftctl/test/acceptance"
	"github.com/snyk/driftctl/test/acceptance/awsutils"
)

func TestAcc_Aws_Instance(t *testing.T) {
	acceptance.Run(t, acceptance.AccTestCase{
		TerraformVersion: "0.15.5",
		Paths:            []string{"./testdata/acc/aws_instance_default"},
		Args:             []string{"scan"},
		Checks: []acceptance.AccCheck{
			{
				Env: map[string]string{
					"AWS_REGION": "us-east-1",
				},
				Check: func(result *test.ScanResult, stdout string, err error) {
					if err != nil {
						t.Fatal(err)
					}
					result.AssertInfrastructureIsInSync()
					result.AssertManagedCount(2)
				},
			},
		},
	})
}

func TestAcc_Aws_Instance_WithBlockDevices(t *testing.T) {
	var mutatedInstanceId string
	acceptance.Run(t, acceptance.AccTestCase{
		TerraformVersion: "0.15.5",
		Paths:            []string{"./testdata/acc/aws_instance"},
		Args:             []string{"scan"},
		Checks: []acceptance.AccCheck{
			{
				Env: map[string]string{
					"AWS_REGION": "us-east-1",
				},
				Check: func(result *test.ScanResult, stdout string, err error) {
					if err != nil {
						t.Fatal(err)
					}
				},
			},
			{
				Env: map[string]string{
					"AWS_REGION": "us-east-1",
				},
				PreExec: func() {
					client := ec2.New(awsutils.Session())
					response, err := client.DescribeInstances(&ec2.DescribeInstancesInput{
						Filters: []*ec2.Filter{
							{
								Name: aws.String("instance-state-name"),
								Values: []*string{
									aws.String("running"),
								},
							},
							{
								Name: aws.String("tag:Name"),
								Values: []*string{
									aws.String("test_instance_1"),
								},
							},
						},
					})
					if err != nil {
						t.Fatal(err)
					}
					if len(response.Reservations) != 1 || len(response.Reservations[0].Instances) != 1 {
						t.Fatal("Error, unexpected number of instances found, manual check required")
					}
					mutatedInstanceId = *response.Reservations[0].Instances[0].InstanceId
					_, err = client.CreateTags(&ec2.CreateTagsInput{
						Resources: []*string{&mutatedInstanceId},
						Tags: []*ec2.Tag{
							{
								Key:   aws.String("Env"),
								Value: aws.String("Production"),
							},
						},
					})
					if err != nil {
						t.Fatal(err)
					}
				},
				Check: func(result *test.ScanResult, stdout string, err error) {
					if err != nil {
						t.Fatal(err)
					}
					result.Equal(0, result.Summary().TotalDeleted)
				},
			},
		},
	})
}
