package google_test

import (
	"testing"
	"time"

	"github.com/snyk/driftctl/test"
	"github.com/snyk/driftctl/test/acceptance"
)

func TestAcc_Google_StorageBucket(t *testing.T) {
	acceptance.Run(t, acceptance.AccTestCase{
		TerraformVersion: "0.15.5",
		Paths:            []string{"./testdata/acc/google_storage_bucket"},
		Args: []string{
			"scan",
			"--to", "gcp+tf",
		},
		Checks: []acceptance.AccCheck{
			{
				// New resources are not visible immediately on GCP api after an apply
				// Logic below retry driftctl scan until we can retrieve the results (infra will be in sync) and for maximum 60 seconds
				ShouldRetry: func(result *test.ScanResult, retryDuration time.Duration, retryCount uint8) bool {
					return !result.IsSync() && retryDuration < time.Minute
				},
				Check: func(result *test.ScanResult, stdout string, err error) {
					if err != nil {
						t.Fatal(err)
					}
					result.AssertInfrastructureIsInSync()
					result.AssertManagedCount(3)
				},
			},
		},
	})
}
