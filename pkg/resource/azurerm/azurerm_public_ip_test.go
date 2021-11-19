package azurerm_test

// Flacky test
// func TestAcc_Azure_PublicIP(t *testing.T) {
// 	acceptance.Run(t, acceptance.AccTestCase{
// 		TerraformVersion: "0.15.5",
// 		Paths:            []string{"./testdata/acc/azurerm_public_ip"},
// 		Args: []string{
// 			"scan",
// 			"--to", "azure+tf",
// 			"--filter", "Type=='azurerm_public_ip'",
// 		},
// 		Checks: []acceptance.AccCheck{
// 			{
// 				Check: func(result *test.ScanResult, stdout string, err error) {
// 					if err != nil {
// 						t.Fatal(err)
// 					}
// 					result.AssertInfrastructureIsInSync()
// 					result.AssertManagedCount(1)
// 				},
// 			},
// 		},
// 	})
// }
