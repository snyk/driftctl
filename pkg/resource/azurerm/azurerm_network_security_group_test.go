package azurerm_test

// Disabled flacky test
// func TestAcc_Azure_NetworkSecurityGroup(t *testing.T) {
// 	acceptance.Run(t, acceptance.AccTestCase{
// 		TerraformVersion: "0.15.5",
// 		Paths:            []string{"./testdata/acc/azurerm_network_security_group"},
// 		Args: []string{
// 			"scan",
// 			"--to", "azure+tf",
// 			"--filter", "contains(Id, 'acceptanceTestSecurityGroup-')",
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
