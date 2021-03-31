package aws_test

import (
	"testing"

	"github.com/cloudskiff/driftctl/pkg/analyser"
	"github.com/cloudskiff/driftctl/test"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	awsresources "github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/cloudskiff/driftctl/test/acceptance"
	"github.com/cloudskiff/driftctl/test/acceptance/awsutils"
	"github.com/r3labs/diff/v2"
)

func TestAcc_AwsInstance_WithBlockDevices(t *testing.T) {
	var mutatedInstanceId string
	acceptance.Run(t, acceptance.AccTestCase{
		Paths: []string{"./testdata/acc/aws_instance"},
		Args:  []string{"scan", "--filter", "Type=='aws_instance'"},
		Checks: []acceptance.AccCheck{
			{
				Env: map[string]string{
					"AWS_REGION": "us-east-1",
				},
				Check: func(result *test.ScanResult, stdout string, err error) {
					if err != nil {
						t.Fatal(err)
					}
					result.AssertDriftCountTotal(0)
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
					result.AssertResourceHasDrift(
						mutatedInstanceId,
						awsresources.AwsInstanceResourceType,
						analyser.Change{
							Change: diff.Change{
								Type: diff.CREATE,
								Path: []string{"Tags", "Env"},
								From: nil,
								To:   "Production",
							},
						},
					)
				},
			},
		},
	})
}
