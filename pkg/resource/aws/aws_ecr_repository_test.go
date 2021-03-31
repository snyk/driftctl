package aws_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go/service/ecr"

	"github.com/cloudskiff/driftctl/pkg/analyser"
	awsresources "github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/cloudskiff/driftctl/test"
	"github.com/cloudskiff/driftctl/test/acceptance"
	"github.com/cloudskiff/driftctl/test/acceptance/awsutils"
	"github.com/r3labs/diff/v2"
)

func TestAcc_AwsECRRepository(t *testing.T) {
	var mutatedRepositoryID string
	acceptance.Run(t, acceptance.AccTestCase{
		Paths: []string{"./testdata/acc/aws_ecr_repository"},
		Args:  []string{"scan", "--filter", "Type=='aws_ecr_repository'"},
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

					mutatedRepositoryID = result.Managed()[0].TerraformId()
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
								Path: []string{"ImageTagMutability"},
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
