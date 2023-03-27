package scaleway_test

import (
	"testing"
	"time"

	"github.com/snyk/driftctl/test"
	"github.com/snyk/driftctl/test/acceptance"
)

func TestAcc_Scaleway_FunctionNamespace(t *testing.T) {
	acceptance.Run(t, acceptance.AccTestCase{
		TerraformVersion: "1.4.2",
		Paths:            []string{"./testdata/acc/scaleway_function_namespace"},
		Args: []string{
			"scan",
			"--to", "scaleway+tf",
		},
		Checks: []acceptance.AccCheck{
			{
				ShouldRetry: func(result *test.ScanResult, retryDuration time.Duration, retryCount uint8) bool {
					return !result.IsSync() && retryDuration < time.Minute
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
