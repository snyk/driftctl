package google_test

import (
	"testing"

	"github.com/cloudskiff/driftctl/test"
	"github.com/cloudskiff/driftctl/test/acceptance"
)

func TestAcc_Google_ComputeFirewall(t *testing.T) {
	acceptance.Run(t, acceptance.AccTestCase{
		TerraformVersion: "0.15.5",
		Paths:            []string{"./testdata/acc/google_compute_firewall"},
		Args: []string{
			"scan",
			"--to", "gcp+tf",
			"--filter", "Type=='google_compute_firewall'",
			"--deep",
		},
		Checks: []acceptance.AccCheck{
			{
				Check: func(result *test.ScanResult, stdout string, err error) {
					if err != nil {
						t.Fatal(err)
					}
					result.AssertManagedCount(3)
					result.AssertDriftCountTotal(0)
					result.AssertUnmanagedCount(4) // Default VPCs
				},
			},
		},
	})
}
