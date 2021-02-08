package aws_test

import (
	"testing"

	"github.com/cloudskiff/driftctl/test/acceptance"
)

func TestAcc_AwsSNSTopicSubscription(t *testing.T) {
	acceptance.Run(t, acceptance.AccTestCase{
		Path: "./testdata/acc/aws_sns_topic_subscription",
		Args: []string{"scan", "--filter", "Type=='aws_sns_topic_subscription'"},
		Checks: []acceptance.AccCheck{
			{
				Env: map[string]string{
					"AWS_REGION": "us-east-1",
				},
				Check: func(result *acceptance.ScanResult, stdout string, err error) {
					if err != nil {
						t.Fatal(err)
					}
					result.AssertDriftCountTotal(0)
					result.AssertDeletedCount(0)
					result.AssertManagedCount(2)
				},
			},
		},
	})
}
