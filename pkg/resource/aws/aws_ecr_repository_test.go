package aws_test

import (
    "testing"

    "github.com/aws/aws-sdk-go/aws"

    "github.com/aws/aws-sdk-go/service/ecr"

    "github.com/r3labs/diff/v2"
    awsresources "github.com/snyk/driftctl/enumeration/resource/aws"
    "github.com/snyk/driftctl/pkg/analyser"
    "github.com/snyk/driftctl/test"
    "github.com/snyk/driftctl/test/acceptance"
    "github.com/snyk/driftctl/test/acceptance/awsutils"
)

func TestAcc_Aws_ECRRepository(t *testing.T) {
    var mutatedRepositoryID string
    acceptance.Run(t, acceptance.AccTestCase{
        TerraformVersion: "1.4.6",
        Paths:            []string{"./testdata/acc/aws_ecr_repository"},
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
                    result.AssertManagedCount(1)

                    mutatedRepositoryID = result.Managed()[0].ResourceId()
                },
            },
            {
                Env: map[string]string{
                    "AWS_REGION": "us-east-1",
                },
                PreExec: func() {
                    client := ecr.New(awsutils.Session())
                    _, err := client.PutImageTagMutability(&ecr.PutImageTagMutabilityInput{
                        RepositoryName:     &mutatedRepositoryID,
                        ImageTagMutability: aws.String("IMMUTABLE"),
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
                    result.AssertDeletedCount(0)
                    result.AssertManagedCount(1)
                    result.AssertUnmanagedCount(0)

                    result.AssertResourceHasDrift(
                        mutatedRepositoryID,
                        awsresources.AwsEcrRepositoryResourceType,
                        analyser.Change{
                            Change: diff.Change{
                                Type: diff.UPDATE,
                                Path: []string{"image_tag_mutability"},
                                From: "MUTABLE",
                                To:   "IMMUTABLE",
                            },
                            Computed: false,
                        },
                    )

                },
            },
        },
    })
}
