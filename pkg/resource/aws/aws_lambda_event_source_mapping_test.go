package aws_test

import (
	"testing"

	"github.com/snyk/driftctl/test"
	"github.com/snyk/driftctl/test/acceptance"
)

// aws lambda list-event-source-mappings to list all event source mappings
// aws lambda delete-event-source-mapping --uuid xxx to delete a specific event source mapping

func TestAcc_Aws_LambdaEventSourceMapping(t *testing.T) {
	acceptance.Run(t, acceptance.AccTestCase{
		TerraformVersion: "1.4.6",
		Paths:            []string{"./testdata/acc/aws_lambda_event_source_mapping"},
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
					result.AssertManagedCount(3)
				},
			},
		},
	})
}
