package azurerm_test

// Flacky test
// func TestAcc_Azure_PostgresqlDatabase(t *testing.T) {
// 	acceptance.Run(t, acceptance.AccTestCase{
// 		TerraformVersion: "0.15.5",
// 		Paths:            []string{"./testdata/acc/azurerm_postgresql_database"},
// 		Args: []string{
// 			"scan",
// 			"--to", "azure+tf",
// 			"--filter", "contains(Id, 'acc-test-db')",
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
