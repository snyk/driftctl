package aws_test

// This test is commented because it will not destroy all created resources, check terraform documentation for more details
// https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/default_network_acl#removing-aws_default_network_acl-from-your-configuration
// You can run it on your side but it is not enabled in CI since it will make other tests to fail, more specifically the aws_network_acl_test one
// A fix can be to manually remove dangling rules through AWS SDK in PostExec hook

// func TestAcc_Aws_DefaultNetworkAcl(t *testing.T) {
// 	acceptance.Run(t, acceptance.AccTestCase{
// 		TerraformVersion: "0.15.5",
// 		Paths:            []string{"./testdata/acc/aws_default_network_acl"},
// 		Args:             []string{"scan"},
// 		Checks: []acceptance.AccCheck{
// 			{
// 				Env: map[string]string{
// 					"AWS_REGION": "us-east-1",
// 				},
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
