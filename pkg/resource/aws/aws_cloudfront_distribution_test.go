package aws_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/cloudfront"

	"github.com/cloudskiff/driftctl/pkg/analyser"
	awsresources "github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/cloudskiff/driftctl/test"

	"github.com/r3labs/diff/v2"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/cloudskiff/driftctl/test/acceptance"
	"github.com/cloudskiff/driftctl/test/acceptance/awsutils"
)

func TestAcc_Aws_CloudfrontDistribution(t *testing.T) {
	var mutatedDistribution string
	acceptance.Run(t, acceptance.AccTestCase{
		Paths:                      []string{"./testdata/acc/aws_cloudfront_distribution"},
		Args:                       []string{"scan", "--filter", "Type=='aws_cloudfront_distribution'"},
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
					mutatedDistribution = result.Managed()[0].TerraformId()
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
					result.AssertDriftCountTotal(1)
					result.AssertResourceHasDrift(
						mutatedDistribution,
						awsresources.AwsCloudfrontDistributionResourceType,
						analyser.Change{
							Change: diff.Change{
								Type: diff.UPDATE,
								Path: []string{"is_ipv6_enabled"},
								From: false,
								To:   true,
							},
						},
					)
				},
			},
		},
	})
}
