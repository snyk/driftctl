package aws_test

import (
    "testing"

    "github.com/aws/aws-sdk-go/service/route53"

    "github.com/aws/aws-sdk-go/aws"
    "github.com/r3labs/diff/v2"
    awsresources "github.com/snyk/driftctl/enumeration/resource/aws"
    "github.com/snyk/driftctl/pkg/analyser"
    "github.com/snyk/driftctl/test"
    "github.com/snyk/driftctl/test/acceptance"
    "github.com/snyk/driftctl/test/acceptance/awsutils"
)

func TestAcc_Aws_Route53HealthCheck(t *testing.T) {
    var mutatedHealthCheckID string
    acceptance.Run(t, acceptance.AccTestCase{
        TerraformVersion: "1.4.6",
        Paths:            []string{"./testdata/acc/aws_route53_health_check"},
        Args:             []string{"scan", "--deep"},
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
                    result.AssertDriftCountTotal(2)
                    result.AssertDeletedCount(0)
                    result.AssertManagedCount(2)

                    result.AssertResourceHasDrift(
                        mutatedHealthCheckID,
                        awsresources.AwsRoute53HealthCheckResourceType,
                        analyser.Change{
                            Change: diff.Change{
                                Type: diff.UPDATE,
                                Path: []string{"disabled"},
                                From: false,
                                To:   true,
                            },
                            Computed: false,
                        },
                    )

                    result.AssertResourceHasDrift(
                        mutatedHealthCheckID,
                        awsresources.AwsRoute53HealthCheckResourceType,
                        analyser.Change{
                            Change: diff.Change{
                                Type: diff.UPDATE,
                                Path: []string{"resource_path"},
                                From: "/",
                                To:   "/bad",
                            },
                            Computed: false,
                        },
                    )
                },
            },
        },
    })
}
