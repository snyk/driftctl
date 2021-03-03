package github_test

import (
	"testing"

	"github.com/cloudskiff/driftctl/test/acceptance"
)

func TestAcc_Github_BranchProtection(t *testing.T) {
	acceptance.Run(t, acceptance.AccTestCase{
		Paths: []string{"./testdata/acc/github_branch_protection"},
		Args: []string{
			"scan",
			"--to", "github+tf",
			"--filter", "Type=='github_branch_protection'",
		},
		Checks: []acceptance.AccCheck{
			{
				Check: func(result *acceptance.ScanResult, stdout string, err error) {
					if err != nil {
						t.Fatal(err)
					}
					result.AssertInfrastructureIsInSync()
					result.AssertManagedCount(6)
				},
			},
		},
	})
}
