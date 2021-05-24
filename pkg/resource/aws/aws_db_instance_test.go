package aws_test

import (
	"testing"

	"github.com/cloudskiff/driftctl/test"
	"github.com/cloudskiff/driftctl/test/acceptance"
)

func TestAcc_AwsDbInstance_WithCharacterSetName(t *testing.T) {
	acceptance.Run(t, acceptance.AccTestCase{
		Paths: []string{"./testdata/acc/aws_db_instance"},
		Args:  []string{"scan", "--filter", "Type=='aws_db_instance'"},
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
