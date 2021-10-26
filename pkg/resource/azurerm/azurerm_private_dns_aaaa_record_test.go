package azurerm_test

import (
	"testing"

	"github.com/cloudskiff/driftctl/test"
	"github.com/cloudskiff/driftctl/test/acceptance"
)

func TestAcc_Azure_PrivateDNSAAAARecord(t *testing.T) {
	acceptance.Run(t, acceptance.AccTestCase{
		TerraformVersion: "0.15.5",
		Paths:            []string{"./testdata/acc/azurerm_private_dns_aaaa_record"},
		Args: []string{
			"scan",
			"--to", "azure+tf", "--deep",
		},
		Checks: []acceptance.AccCheck{
			{
				Check: func(result *test.ScanResult, stdout string, err error) {
					if err != nil {
						t.Fatal(err)
					}
					result.AssertInfrastructureIsNotSync()
					result.AssertManagedCount(3)
					result.AssertUnmanagedCount(1)
					result.AssertDriftCountTotal(0)
					result.AssertCoverage(60)
				},
			},
		},
	})
}
