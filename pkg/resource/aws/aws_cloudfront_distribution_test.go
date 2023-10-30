package aws_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/cloudfront"

	"github.com/snyk/driftctl/test"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/snyk/driftctl/test/acceptance"
	"github.com/snyk/driftctl/test/acceptance/awsutils"
)

func TestAcc_Aws_CloudfrontDistribution(t *testing.T) {
	t.Skip("flake")

	var mutatedDistribution string
	acceptance.Run(t, acceptance.AccTestCase{
		TerraformVersion:           "0.15.5",
		Paths:                      []string{"./testdata/acc/aws_cloudfront_distribution"},
		Args:                       []string{"scan"},
		ShouldRefreshBeforeDestroy: true,
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
					mutatedDistribution = result.Managed()[0].ResourceId()
				},
			},
			{
				Env: map[string]string{
					"AWS_REGION": "us-east-1",
				},
				PreExec: func() {
					client := cloudfront.New(awsutils.Session())
					res, err := client.GetDistributionConfig(&cloudfront.GetDistributionConfigInput{
						Id: aws.String(mutatedDistribution),
					})
					if err != nil {
						t.Fatal(err)
					}
					res.DistributionConfig.SetIsIPV6Enabled(true)
					_, err = client.UpdateDistribution(&cloudfront.UpdateDistributionInput{
						Id:                 aws.String(mutatedDistribution),
						DistributionConfig: res.DistributionConfig,
						IfMatch:            res.ETag,
					})
					if err != nil {
						t.Fatal(err)
					}
				},
				Check: func(result *test.ScanResult, stdout string, err error) {
					if err != nil {
						t.Fatal(err)
					}
				},
			},
		},
	})
}
