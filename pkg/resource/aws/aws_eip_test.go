package aws_test

import (
	"testing"
	"time"

	"github.com/snyk/driftctl/test"
	"github.com/snyk/driftctl/test/acceptance"
)

func TestAcc_Aws_Eip(t *testing.T) {
	acceptance.Run(t, acceptance.AccTestCase{
		TerraformVersion: "0.15.5",
		Paths:            []string{"./testdata/acc/aws_eip"},
		Args:             []string{"scan", "--deep"},
		RetryDestroy: acceptance.RetryConfig{
			Attempts: 3,
			Delay:    5 * time.Second,
		},
		Checks: []acceptance.AccCheck{
			{
				Env: map[string]string{
					"AWS_REGION": "us-east-1",
				},
				// New resources are not visible immediately through AWS API after an apply operation
				// (e.g. error attaching EC2 Internet Gateway)
				// Logic below retries driftctl scan using a back-off strategy of retrying 'n' times
				// and doubling the amount of time waited after each one.
				ShouldRetry: func(result *test.ScanResult, retryDuration time.Duration, retryCount uint8) bool {
					if result.IsSync() || retryDuration > 10*time.Minute {
						return false
					}
					time.Sleep((2 * time.Duration(retryCount)) * time.Minute)
					return true
				},
				Check: func(result *test.ScanResult, stdout string, err error) {
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
