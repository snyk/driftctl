package aws_test

import (
	"testing"

	"github.com/cloudskiff/driftctl/test"
	"github.com/cloudskiff/driftctl/test/acceptance"
)

func TestAcc_Aws_IamAccessKey(t *testing.T) {
	acceptance.Run(t, acceptance.AccTestCase{
		TerraformVersion: "0.14.9",
		Paths:            []string{"./testdata/acc/aws_iam_access_key"},
		Args:             []string{"scan", "--filter", "Type=='aws_iam_access_key'"},
		Checks: []acceptance.AccCheck{
			{
				Env: map[string]string{
					"AWS_REGION": "us-east-1",
				},
				Check: func(result *test.ScanResult, stdout string, err error) {
					if err != nil {
						t.Fatal(err)
					}
					result.AssertDriftCountTotal(0)
					result.AssertDeletedCount(0)
					result.AssertManagedCount(1)
				},
			},
		},
	})
}
