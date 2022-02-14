package google_test

import (
	"testing"
	"time"

	"github.com/snyk/driftctl/test"
	"github.com/snyk/driftctl/test/acceptance"
)

func TestAcc_Google_ComputeInstance(t *testing.T) {
	acceptance.Run(t, acceptance.AccTestCase{
		TerraformVersion: "0.15.5",
		Paths:            []string{"./testdata/acc/google_compute_instance"},
		Args: []string{
			"scan",
			"--to", "gcp+tf",
		},
		Checks: []acceptance.AccCheck{
			{
				// New resources are not visible immediately through GCP API after an apply operation.
				// Logic below retries driftctl scan using a back-off strategy of retrying 'n' times
				// and doubling the amount of time waited after each one.
				ShouldRetry: func(result *test.ScanResult, retryDuration time.Duration, retryCount uint8) bool {
					if result.IsSync() || retryDuration > 15*time.Minute {
						return false
					}
					time.Sleep((2 * time.Duration(retryCount)) * time.Minute)
					return true
				},
				Check: func(result *test.ScanResult, stdout string, err error) {
					if err != nil {
						t.Fatal(err)
					}
					result.AssertInfrastructureIsInSync()
					result.AssertManagedCount(1)
				},
			},
		},
	})
}
