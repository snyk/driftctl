package aws_test

import (
	"testing"
	"time"

	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	"github.com/cloudskiff/driftctl/pkg/remote/cache"
	"github.com/cloudskiff/driftctl/test"
	"github.com/cloudskiff/driftctl/test/acceptance"
	"github.com/cloudskiff/driftctl/test/acceptance/awsutils"
	"github.com/sirupsen/logrus"
)

func TestAcc_AwsSNSTopicPolicy(t *testing.T) {
	acceptance.Run(t, acceptance.AccTestCase{
		TerraformVersion: "0.15.5",
		Paths:            []string{"./testdata/acc/aws_sns_topic_policy"},
		Args:             []string{"scan", "--deep"},
		Checks: []acceptance.AccCheck{
			{
				Env: map[string]string{
					"AWS_REGION": "us-east-1",
				},
				PreExec: func() {
					err := acceptance.RetryFor(60*time.Second, func(doneCh chan struct{}) error {
						client := repository.NewSNSRepository(awsutils.Session(), cache.New(0))
						topics, err := client.ListAllTopics()
						if err != nil {
							logrus.Warnf("Cannot list topics: %+v", err)
							return err
						}
						if len(topics) == 3 {
							doneCh <- struct{}{}
						}
						return nil
					})
					if err != nil {
						t.Fatal("Timeout while fetching SNS TOPIC")
					}
				},
				Check: func(result *test.ScanResult, stdout string, err error) {
					if err != nil {
						t.Fatal(err)
					}
					result.AssertInfrastructureIsInSync()
					result.AssertManagedCount(6)
				},
			},
		},
	})
}
