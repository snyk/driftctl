package aws_test

import (
	"testing"

	"github.com/cloudskiff/driftctl/test"
	"github.com/cloudskiff/driftctl/test/acceptance"
)

func TestAcc_AwsIamPolicyAttachment_WithGroupsUsers(t *testing.T) {
	acceptance.Run(t, acceptance.AccTestCase{
		TerraformVersion: "0.15.5",
		Paths:            []string{"./testdata/acc/aws_iam_policy_attachment"},
		Args:             []string{"scan", "--filter", "Type=='aws_iam_policy_attachment'", "--deep"},
		Checks: []acceptance.AccCheck{
			{
				Check: func(result *test.ScanResult, stdout string, err error) {
					if err != nil {
						t.Fatal(err)
					}
					result.AssertDriftCountTotal(0)
					result.Equal(1, result.Summary().TotalManaged)
				},
			},
		},
	})
}
