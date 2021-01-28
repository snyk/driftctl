package aws_test

import (
	"testing"

	"github.com/cloudskiff/driftctl/test/acceptance"
)

func TestAcc_AwsRoute53Record_WithFQDNAsId(t *testing.T) {
	acceptance.Run(t, acceptance.AccTestCase{
		Path: "./testdata/acc/aws_route53_record",
		Args: []string{"scan", "--filter", "Type=='aws_route53_record'"},
		Checks: []acceptance.AccCheck{
			{
				Env: map[string]string{
					"AWS_REGION": "us-east-1",
				},
				Check: func(result *acceptance.ScanResult, stdout string, err error) {
					if err != nil {
						t.Fatal(err)
					}
					result.AssertDriftCountTotal(0)
					result.Equal(0, result.Summary().TotalDeleted)
					result.Equal(8, result.Summary().TotalManaged)
				},
			},
		},
	})
}
