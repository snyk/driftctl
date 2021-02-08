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
				Env: map[string]string{
					"AWS_REGION": "us-east-1",
				},
				Check: func(result *acceptance.ScanResult, stdout string, err error) {
					if err != nil {
						t.Fatal(err)
					}
					result.AssertManagedCount(1)
					result.Equal("aws_s3_bucket", result.Analysis.Managed()[0].TerraformType())
				},
			},
		},
	})
}
