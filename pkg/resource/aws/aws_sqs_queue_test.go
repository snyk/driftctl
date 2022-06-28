package aws_test

import (
	"testing"
	"time"

	"github.com/sirupsen/logrus"

	awsresources "github.com/snyk/driftctl/enumeration/resource/aws"
	"github.com/snyk/driftctl/pkg/analyser"
	"github.com/snyk/driftctl/test"

	"github.com/r3labs/diff/v2"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/snyk/driftctl/test/acceptance"
	"github.com/snyk/driftctl/test/acceptance/awsutils"
)

func TestAcc_Aws_SQSQueue(t *testing.T) {
	var mutatedQueue string
	acceptance.Run(t, acceptance.AccTestCase{
		TerraformVersion: "0.15.5",
		Paths:            []string{"./testdata/acc/aws_sqs_queue"},
		Args:             []string{"scan", "--deep"},
		Checks: []acceptance.AccCheck{
			{
				Env: map[string]string{
					"AWS_REGION": "us-east-1",
				},
				PreExec: func() {
					err := acceptance.RetryFor(60*time.Second, func(doneCh chan struct{}) error {
						return sqs.New(awsutils.Session()).ListQueuesPages(&sqs.ListQueuesInput{},
							func(resp *sqs.ListQueuesOutput, lastPage bool) bool {
								logrus.Debugf("Retrieved %d SQS queues", len(resp.QueueUrls))
								if len(resp.QueueUrls) >= 2 {
									doneCh <- struct{}{}
								}
								return !lastPage
							},
						)
					})
					if err != nil {
						t.Fatal("Timeout while fetching SQS queues")
					}
				},
				Check: func(result *test.ScanResult, stdout string, err error) {
					if err != nil {
						t.Fatal(err)
					}
					result.AssertInfrastructureIsInSync()
					result.Equal(2, result.Summary().TotalManaged)
					mutatedQueue = result.Managed()[0].ResourceId()
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
				Check: func(result *test.ScanResult, stdout string, err error) {
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
								Path: []string{"delay_seconds"},
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
