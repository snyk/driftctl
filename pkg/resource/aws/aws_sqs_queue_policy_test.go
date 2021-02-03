package aws_test

import (
	"testing"

	"github.com/cloudskiff/driftctl/test/acceptance"
)

func TestAcc_AwsSqsQueuePolicy(t *testing.T) {
	acceptance.Run(t, acceptance.AccTestCase{
		Path: "./testdata/acc/aws_sqs_queue_policy",
		Args: []string{"scan", "--filter", "Type=='aws_sqs_queue_policy'"},
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
				},
			},
		},
	})
}
