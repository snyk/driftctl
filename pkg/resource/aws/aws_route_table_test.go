package aws_test

import (
	"testing"

	"github.com/snyk/driftctl/test"
	"github.com/snyk/driftctl/test/acceptance"
)

func TestAcc_Aws_RouteTable(t *testing.T) {
	t.Skip("flake")

	acceptance.Run(t, acceptance.AccTestCase{
		TerraformVersion: "1.0.2",
		Paths:            []string{"./testdata/acc/aws_route_table"},
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
					result.Equal(2, result.Summary().TotalManaged)
				},
			},
		},
	})
}
