package aws_test

import (
	"testing"

	"github.com/cloudskiff/driftctl/test/acceptance"
)

func TestAcc_AwsS3Bucket_BucketInUsEast1(t *testing.T) {
	acceptance.Run(t, acceptance.AccTestCase{
		Path: "./testdata/acc/aws_s3_bucket",
		Args: []string{"scan", "--filter", "Type=='aws_s3_bucket'"},
		Checks: []acceptance.AccCheck{
			{
				Check: func(result *acceptance.ScanResult, stdout string, err error) {
					if err != nil {
						t.Fatal(err)
					}
					result.Equal(1, result.Summary().TotalManaged)
					result.Equal("aws_s3_bucket", result.Analysis.Managed()[0].TerraformType())
					result.Equal("foobar.driftctl-test.com", result.Analysis.Managed()[0].TerraformId())
				},
			},
		},
	})
}
