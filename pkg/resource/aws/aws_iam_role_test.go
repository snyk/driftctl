package aws_test

import (
    "testing"

    "github.com/snyk/driftctl/test"
    "github.com/snyk/driftctl/test/acceptance"
)

func TestAcc_Aws_IamRole(t *testing.T) {
    acceptance.Run(t, acceptance.AccTestCase{
        TerraformVersion: "1.4.6",
        Paths:            []string{"./testdata/acc/aws_iam_role"},
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
                    result.AssertDriftCountTotal(0)
                    result.AssertDeletedCount(0)
                    result.AssertManagedCount(1)
                },
            },
        },
    })
}

func TestAcc_Aws_IamRole_WithManaged(t *testing.T) {
    acceptance.Run(t, acceptance.AccTestCase{
        TerraformVersion: "1.4.6",
        Paths:            []string{"./testdata/acc/aws_iam_role_with_managed_policies"},
        Args:             []string{"scan"},
        Checks: []acceptance.AccCheck{
            {
                Env: map[string]string{
                    "AWS_REGION": "us-east-1",
                },
                Check: func(result *test.ScanResult, stdout string, err error) {
                    if err != nil {
                        t.Fatal(err)
                    }
                    result.AssertDriftCountTotal(0)
                    result.AssertDeletedCount(0)
                    result.AssertManagedCount(2)
                },
            },
        },
    })
}
