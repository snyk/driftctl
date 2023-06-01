package aws_test

import (
    "testing"

    "github.com/snyk/driftctl/test"
    "github.com/snyk/driftctl/test/acceptance"
)

func TestAcc_Aws_SecurityGroupRule(t *testing.T) {
    acceptance.Run(t, acceptance.AccTestCase{
        TerraformVersion: "1.4.6",
        Paths:            []string{"./testdata/acc/aws_security_group_rule"},
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
                    result.AssertManagedCount(21)
                    result.AssertDeletedCount(2)
                    result.AssertUnmanagedCount(5)
                    result.AssertDriftCountTotal(0)
                },
            },
        },
    })
}
