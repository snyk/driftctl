package aws_test

import (
    "testing"
    "time"

    "github.com/snyk/driftctl/test"
    "github.com/snyk/driftctl/test/acceptance"
)

func TestAcc_Aws_EipAssociation(t *testing.T) {
    acceptance.Run(t, acceptance.AccTestCase{
        TerraformVersion: "1.4.6",
        Paths:            []string{"./testdata/acc/aws_eip_association"},
        Args:             []string{"scan", "--deep"},
        RetryDestroy: acceptance.RetryConfig{
            Attempts: 3,
            Delay:    5 * time.Second,
        },
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
                    result.AssertManagedCount(2)
                },
            },
        },
    })
}
