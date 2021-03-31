package aws_test

import (
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	"github.com/cloudskiff/driftctl/test"

	"github.com/aws/aws-sdk-go/service/sns"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/cloudskiff/driftctl/pkg/analyser"
	awsresources "github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/cloudskiff/driftctl/test/acceptance"
	"github.com/cloudskiff/driftctl/test/acceptance/awsutils"
	"github.com/r3labs/diff/v2"
)

func TestAcc_AwsSNSTopic(t *testing.T) {
	var mutatedTopicArn string
	acceptance.Run(t, acceptance.AccTestCase{
		Paths: []string{"./testdata/acc/aws_sns_topic"},
		Args:  []string{"scan", "--filter", "Type=='aws_sns_topic'"},
		Checks: []acceptance.AccCheck{
			{
				Env: map[string]string{
					"AWS_REGION": "us-east-1",
				},
				PreExec: func() {
					err := acceptance.RetryFor(60*time.Second, func(doneCh chan struct{}) error {
						client := repository.NewSNSClient(awsutils.Session())
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
					result.AssertDriftCountTotal(0)
					result.AssertDeletedCount(0)
					result.AssertManagedCount(3)

					for _, resource := range result.Analysis.Managed() {
						if strings.Contains(resource.TerraformId(), "user-updates-topic3") {
							mutatedTopicArn = resource.TerraformId()
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
								Path: []string{"DisplayName"},
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
