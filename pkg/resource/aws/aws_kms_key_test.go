package aws_test

import (
	"testing"

	"github.com/cloudskiff/driftctl/test/acceptance"
)

func TestAcc_Aws_KMSKey(t *testing.T) {
	acceptance.Run(t, acceptance.AccTestCase{
		Paths: []string{"./testdata/acc/aws_kms_key"},
		Args:  []string{"scan", "--filter", "Type=='aws_kms_key'"},
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
					result.AssertManagedCount(3)
				},
			},
		},
	})
}
