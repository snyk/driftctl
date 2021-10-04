package azurerm_test

import (
	"testing"

	"github.com/cloudskiff/driftctl/test"
	"github.com/cloudskiff/driftctl/test/acceptance"
)

func TestAcc_Azure_Subnet(t *testing.T) {
	acceptance.Run(t, acceptance.AccTestCase{
		TerraformVersion: "0.15.5",
		Paths:            []string{"./testdata/acc/azurerm_subnet"},
		Args: []string{
			"scan",
			"--to", "azure+tf",
			"--filter", "Type=='azurerm_subnet' || Type=='azurerm_virtual_network'",
		},
		Checks: []acceptance.AccCheck{
			{
				Check: func(result *test.ScanResult, stdout string, err error) {
					if err != nil {
						t.Fatal(err)
					}
					result.AssertInfrastructureIsInSync()
					// We should have one azurerm_virtual_network and two azurerm_subnet
					result.AssertManagedCount(3)
				},
			},
		},
	})
}
