package azurerm_test

// Flacky test
// func TestAcc_Azure_ContainerRegistry(t *testing.T) {
// 	acceptance.Run(t, acceptance.AccTestCase{
// 		TerraformVersion: "0.15.5",
// 		Paths:            []string{"./testdata/acc/azurerm_container_registry"},
// 		Args: []string{
// 			"scan",
// 			"--to", "azure+tf",
// 			"--filter", "contains(Id, 'containerRegistryAcc')",
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
