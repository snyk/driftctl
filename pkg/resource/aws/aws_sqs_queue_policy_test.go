package aws_test

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/cloudskiff/driftctl/test"
	"github.com/cloudskiff/driftctl/test/acceptance/awsutils"
	"github.com/sirupsen/logrus"

	"github.com/cloudskiff/driftctl/test/acceptance"
)

func TestAcc_AwsSQSQueuePolicy(t *testing.T) {
	acceptance.Run(t, acceptance.AccTestCase{
		TerraformVersion: "0.15.5",
		Paths:            []string{"./testdata/acc/aws_sqs_queue_policy"},
		Args:             []string{"scan", "--filter", "Type=='aws_sqs_queue_policy'", "--deep"},
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
								if len(resp.QueueUrls) >= 3 {
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
					result.AssertManagedCount(2)
				},
			},
		},
	})
}
