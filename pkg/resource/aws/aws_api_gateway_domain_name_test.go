package aws_test

import (
    "testing"

    "github.com/snyk/driftctl/test"
    "github.com/snyk/driftctl/test/acceptance"
)

func TestAcc_Aws_ApiGatewayDomainName(t *testing.T) {
    acceptance.Run(t, acceptance.AccTestCase{
        TerraformVersion: "1.4.6",
        Paths:            []string{"./testdata/acc/aws_api_gateway_domain_name"},
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
                    result.AssertInfrastructureIsInSync()
                    result.AssertManagedCount(1)
                },
            },
        },
    })
}
