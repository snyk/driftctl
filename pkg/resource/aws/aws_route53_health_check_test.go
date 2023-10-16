package aws_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/route53"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/snyk/driftctl/test"
	"github.com/snyk/driftctl/test/acceptance"
	"github.com/snyk/driftctl/test/acceptance/awsutils"
)

func TestAcc_Aws_Route53HealthCheck(t *testing.T) {
	var mutatedHealthCheckID string
	acceptance.Run(t, acceptance.AccTestCase{
		TerraformVersion: "0.15.5",
		Paths:            []string{"./testdata/acc/aws_route53_health_check"},
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
					result.AssertManagedCount(2)

					mutatedHealthCheckID = result.Managed()[0].ResourceId()
				},
			},
			{
				Env: map[string]string{
					"AWS_REGION": "us-east-1",
				},
				PreExec: func() {
					client := route53.New(awsutils.Session())
					_, err := client.UpdateHealthCheck(&route53.UpdateHealthCheckInput{
						Disabled:      aws.Bool(true),
						HealthCheckId: &mutatedHealthCheckID,
						ResourcePath:  aws.String("/bad"),
					})
					if err != nil {
						t.Fatal(err)
					}
				},
				Check: func(result *test.ScanResult, stdout string, err error) {
					if err != nil {
						t.Fatal(err)
					}
					result.AssertDeletedCount(0)
					result.AssertManagedCount(2)
				},
			},
		},
	})
}
