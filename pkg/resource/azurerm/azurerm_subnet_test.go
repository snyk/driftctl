package azurerm_test

// Flacky test
// func TestAcc_Azure_Subnet(t *testing.T) {
// 	acceptance.Run(t, acceptance.AccTestCase{
// 		TerraformVersion: "0.15.5",
// 		Paths:            []string{"./testdata/acc/azurerm_subnet"},
// 		Args: []string{
// 			"scan",
// 			"--to", "azure+tf",
// 		},
// 		Checks: []acceptance.AccCheck{
// 			{
// 				Check: func(result *test.ScanResult, stdout string, err error) {
// 					if err != nil {
// 						t.Fatal(err)
// 					}
// 					result.AssertInfrastructureIsInSync()
// 					// We should have one azurerm_virtual_network and two azurerm_subnet
// 					result.AssertManagedCount(3)
// 				},
// 			},
// 		},
// 	})
// }
