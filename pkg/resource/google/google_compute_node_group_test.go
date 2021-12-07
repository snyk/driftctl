package google_test

import (
	"testing"

	"github.com/snyk/driftctl/test"
	"github.com/snyk/driftctl/test/acceptance"
)

func TestAcc_Google_ComputeNodeGroup(t *testing.T) {
	t.Skip("Skipping acc test for Google_ComputeNodeGroup due to pricing of the sole tenant resources")
	acceptance.Run(t, acceptance.AccTestCase{
		TerraformVersion: "0.15.5",
		Paths:            []string{"./testdata/acc/google_compute_node_group"},
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
					result.AssertInfrastructureIsInSync()
					result.AssertManagedCount(2)
				},
			},
		},
	})
}
