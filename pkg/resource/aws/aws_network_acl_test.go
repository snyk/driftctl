package aws_test

import (
    "testing"
    "time"

    "github.com/snyk/driftctl/test"
    "github.com/snyk/driftctl/test/acceptance"
)

// This test cover both aws_network_acl and `aws_network_acl_rule`
func TestAcc_Aws_NetworkAcl(t *testing.T) {
    acceptance.Run(t, acceptance.AccTestCase{
        TerraformVersion: "1.4.6",
        Paths:            []string{"./testdata/acc/aws_network_acl"},
        Args:             []string{"scan", "--deep"},
        Checks: []acceptance.AccCheck{
            {
                Env: map[string]string{
                    "AWS_REGION": "us-east-1",
                },
                Check: func(result *test.ScanResult, stdout string, err error) {
                    if err != nil {
                        t.Fatal(err)
                    }
                    result.AssertInfrastructureIsInSync()
                    result.AssertManagedCount(7)
                },
            },
        },
    })
}

// This test cover both aws_network_acl and `aws_network_acl_rule`
func TestAcc_Aws_NetworkAcl_NonDeep(t *testing.T) {
    acceptance.Run(t, acceptance.AccTestCase{
        TerraformVersion: "1.4.6",
        Paths:            []string{"./testdata/acc/aws_network_acl"},
        Args:             []string{"scan", "--filter", "Type=='aws_network_acl' ||  Type=='aws_network_acl_rule'"},
        Checks: []acceptance.AccCheck{
            {
                Env: map[string]string{
                    "AWS_REGION": "us-east-1",
                },
                ShouldRetry: acceptance.LinearBackoff(10 * time.Minute),
                Check: func(result *test.ScanResult, stdout string, err error) {
                    if err != nil {
                        t.Fatal(err)
                    }
                    result.AssertInfrastructureIsInSync()
                    result.AssertManagedCount(7)
                },
            },
        },
    })
}
