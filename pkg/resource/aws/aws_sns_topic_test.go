package aws_test

import (
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/snyk/driftctl/pkg/remote/cache"

	"github.com/snyk/driftctl/pkg/remote/aws/repository"
	"github.com/snyk/driftctl/test"

	"github.com/aws/aws-sdk-go/service/sns"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/r3labs/diff/v2"
	"github.com/snyk/driftctl/pkg/analyser"
	awsresources "github.com/snyk/driftctl/pkg/resource/aws"
	"github.com/snyk/driftctl/test/acceptance"
	"github.com/snyk/driftctl/test/acceptance/awsutils"
)

func TestAcc_Aws_SNSTopic(t *testing.T) {
	var mutatedTopicArn string
	acceptance.Run(t, acceptance.AccTestCase{
		TerraformVersion: "0.15.5",
		Paths:            []string{"./testdata/acc/aws_sns_topic"},
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
					result.AssertManagedCount(3)

					for _, resource := range result.Analysis.Managed() {
						if strings.Contains(resource.ResourceId(), "user-updates-topic3") {
							mutatedTopicArn = resource.ResourceId()
						}
					}
				},
			},
			{
				Env: map[string]string{
					"AWS_REGION": "us-east-1",
				},
				PreExec: func() {
					client := sns.New(awsutils.Session())
					_, err := client.SetTopicAttributes(&sns.SetTopicAttributesInput{
						AttributeName:  aws.String("DisplayName"),
						AttributeValue: aws.String("CHANGED"),
						TopicArn:       &mutatedTopicArn,
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
					result.AssertManagedCount(3)

					result.AssertResourceHasDrift(
						mutatedTopicArn,
						awsresources.AwsSnsTopicResourceType,
						analyser.Change{
							Change: diff.Change{
								Type: diff.UPDATE,
								Path: []string{"display_name"},
								From: "user-updates-topic3",
								To:   "CHANGED",
							},
						},
					)
				},
			},
		},
	})
}
