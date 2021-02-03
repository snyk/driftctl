package aws_test

import (
	"testing"

	"github.com/cloudskiff/driftctl/pkg/analyser"
	awsresources "github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/r3labs/diff/v2"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/cloudskiff/driftctl/test/acceptance"
	"github.com/cloudskiff/driftctl/test/acceptance/awsutils"
)

func TestAcc_AwsSqsQueue(t *testing.T) {
	var mutatedQueue string
	acceptance.Run(t, acceptance.AccTestCase{
		Path: "./testdata/acc/aws_sqs_queue",
		Args: []string{"scan", "--filter", "Type=='aws_sqs_queue'"},
		Checks: []acceptance.AccCheck{
			{
				Env: map[string]string{
					"AWS_REGION": "us-east-1",
				},
				Check: func(result *acceptance.ScanResult, stdout string, err error) {
					if err != nil {
						t.Fatal(err)
					}
					result.AssertInfrastructureIsInSync()
					result.Equal(2, result.Summary().TotalManaged)
					mutatedQueue = result.Managed()[0].TerraformId()
				},
			},
			{
				Env: map[string]string{
					"AWS_REGION": "us-east-1",
				},
				PreExec: func() {
					client := sqs.New(awsutils.Session())
					attributes := make(map[string]*string)
					attributes["DelaySeconds"] = aws.String("200")
					_, err := client.SetQueueAttributes(&sqs.SetQueueAttributesInput{
						Attributes: attributes,
						QueueUrl:   aws.String(mutatedQueue),
					})
					if err != nil {
						t.Fatal(err)
					}
				},
				Check: func(result *acceptance.ScanResult, stdout string, err error) {
					if err != nil {
						t.Fatal(err)
					}
					result.AssertDriftCountTotal(1)
					result.AssertResourceHasDrift(
						mutatedQueue,
						awsresources.AwsSqsQueueResourceType,
						analyser.Change{
							Change: diff.Change{
								Type: diff.UPDATE,
								Path: []string{"DelaySeconds"},
								From: float64(0),
								To:   float64(200),
							},
						},
					)
				},
			},
		},
	})
}
