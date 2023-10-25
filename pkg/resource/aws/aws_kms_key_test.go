package aws_test

import (
	"testing"
	"time"

	"github.com/snyk/driftctl/test"
	"github.com/snyk/driftctl/test/acceptance"
)

func TestAcc_Aws_KMSKey(t *testing.T) {
	acceptance.Run(t, acceptance.AccTestCase{
		TerraformVersion: "0.15.5",
		Paths:            []string{"./testdata/acc/aws_kms_key"},
		Args:             []string{"scan"},
		Checks: []acceptance.AccCheck{
			{
				Env: map[string]string{
					"AWS_REGION": "us-east-1",
				},
				ShouldRetry: acceptance.LinearBackoff(10 * time.Minute),
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
