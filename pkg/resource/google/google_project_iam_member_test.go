package google_test

import (
	"testing"

	"github.com/snyk/driftctl/test"
	"github.com/snyk/driftctl/test/acceptance"
)

func TestAcc_Google_ProjectIAMMember(t *testing.T) {
	acceptance.Run(t, acceptance.AccTestCase{
		TerraformVersion: "0.15.5",
		Paths:            []string{"./testdata/acc/google_project_iam_member"},
		Args: []string{
			"scan",
			"--to", "gcp+tf",
		},
		Checks: []acceptance.AccCheck{
			{
				Check: func(result *test.ScanResult, stdout string, err error) {
					if err != nil {
						t.Fatal(err)
					}
					result.AssertManagedCount(2)
				},
			},
		},
	})
}
