package azurerm_test

// Disabled flacky test
// func TestAcc_Azure_PostgresqlServer(t *testing.T) {
// 	acceptance.Run(t, acceptance.AccTestCase{
// 		TerraformVersion: "0.15.5",
// 		Paths:            []string{"./testdata/acc/azurerm_postgresql_server"},
// 		Args: []string{
// 			"scan",
// 			"--to", "azure+tf",
// 			"--filter", "Type=='azurerm_postgresql_server' && contains(Id, 'acc-postgresql-server-')",
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
