package aws_test

import (
	"testing"

	"github.com/cloudskiff/driftctl/test/acceptance"
)

func TestAcc_AwsInternetGateway(t *testing.T) {
	acceptance.Run(t, acceptance.AccTestCase{
		Path: "./testdata/acc/aws_internet_gateway",
		Args: []string{"scan", "--filter", "Type=='aws_internet_gateway'"},
		Checks: []acceptance.AccCheck{
			{
				Env: map[string]string{
					"AWS_REGION": "us-east-1",
				},
				Check: func(result *acceptance.ScanResult, stdout string, err error) {
					if err != nil {
						t.Fatal(err)
					}
					result.AssertInfrastructureIsInSync()
					result.Equal(1, result.Summary().TotalManaged)
				},
			},
		},
	})
}
