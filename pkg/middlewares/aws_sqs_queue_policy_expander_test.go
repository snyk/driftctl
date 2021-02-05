package middlewares

import (
	"strings"
	"testing"

	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/r3labs/diff/v2"
)

func TestAwsSqsQueuePolicyExpander_Execute(t *testing.T) {
	tests := []struct {
		name               string
		resourcesFromState []resource.Resource
		expected           []resource.Resource
	}{
		{
			"Inline policy, no aws_sqs_queue_policy attached",
			[]resource.Resource{
				&aws.AwsSqsQueue{
					Id:     "foo",
					Policy: awssdk.String("{\"Id\":\"MYINLINESQSPOLICY\",\"Statement\":[{\"Action\":\"sqs:SendMessage\",\"Effect\":\"Allow\",\"Principal\":\"*\",\"Resource\":\"arn:aws:sqs:eu-west-3:047081014315:foo\",\"Sid\":\"Stmt1611769527792\"}],\"Version\":\"2012-10-17\"}"),
				},
			},
			[]resource.Resource{
				&aws.AwsSqsQueue{
					Id:     "foo",
					Policy: nil,
				},
				&aws.AwsSqsQueuePolicy{
					Id:       "foo",
					QueueUrl: awssdk.String("foo"),
					Policy:   awssdk.String("{\"Id\":\"MYINLINESQSPOLICY\",\"Statement\":[{\"Action\":\"sqs:SendMessage\",\"Effect\":\"Allow\",\"Principal\":\"*\",\"Resource\":\"arn:aws:sqs:eu-west-3:047081014315:foo\",\"Sid\":\"Stmt1611769527792\"}],\"Version\":\"2012-10-17\"}"),
				},
			},
		},
		{
			"No inline policy, aws_sqs_queue_policy attached",
			[]resource.Resource{
				&aws.AwsSqsQueue{
					Id:     "foo",
					Policy: nil,
				},
				&aws.AwsSqsQueuePolicy{
					Id:       "foo",
					QueueUrl: awssdk.String("foo"),
					Policy:   awssdk.String("{\"Id\":\"MYSQSPOLICY\",\"Statement\":[{\"Action\":\"sqs:SendMessage\",\"Effect\":\"Allow\",\"Principal\":\"*\",\"Resource\":\"arn:aws:sqs:eu-west-3:047081014315:foo\",\"Sid\":\"Stmt1611769527792\"}],\"Version\":\"2012-10-17\"}"),
				},
			},
			[]resource.Resource{
				&aws.AwsSqsQueue{
					Id:     "foo",
					Policy: nil,
				},
				&aws.AwsSqsQueuePolicy{
					Id:       "foo",
					QueueUrl: awssdk.String("foo"),
					Policy:   awssdk.String("{\"Id\":\"MYSQSPOLICY\",\"Statement\":[{\"Action\":\"sqs:SendMessage\",\"Effect\":\"Allow\",\"Principal\":\"*\",\"Resource\":\"arn:aws:sqs:eu-west-3:047081014315:foo\",\"Sid\":\"Stmt1611769527792\"}],\"Version\":\"2012-10-17\"}"),
				},
			},
		},
		{
			"Inline policy and aws_sqs_queue_policy",
			[]resource.Resource{
				&aws.AwsSqsQueue{
					Id:     "foo",
					Policy: awssdk.String("{\"Id\":\"MYINLINESQSPOLICY\",\"Statement\":[{\"Action\":\"sqs:SendMessage\",\"Effect\":\"Allow\",\"Principal\":\"*\",\"Resource\":\"arn:aws:sqs:eu-west-3:047081014315:foo\",\"Sid\":\"Stmt1611769527792\"}],\"Version\":\"2012-10-17\"}"),
				},
				&aws.AwsSqsQueuePolicy{
					Id:       "foo",
					QueueUrl: awssdk.String("foo"),
					Policy:   awssdk.String("{\"Id\":\"MYSQSPOLICY\",\"Statement\":[{\"Action\":\"sqs:SendMessage\",\"Effect\":\"Allow\",\"Principal\":\"*\",\"Resource\":\"arn:aws:sqs:eu-west-3:047081014315:foo\",\"Sid\":\"Stmt1611769527792\"}],\"Version\":\"2012-10-17\"}"),
				},
			},
			[]resource.Resource{
				&aws.AwsSqsQueue{
					Id:     "foo",
					Policy: nil,
				},
				&aws.AwsSqsQueuePolicy{
					Id:       "foo",
					QueueUrl: awssdk.String("foo"),
					Policy:   awssdk.String("{\"Id\":\"MYSQSPOLICY\",\"Statement\":[{\"Action\":\"sqs:SendMessage\",\"Effect\":\"Allow\",\"Principal\":\"*\",\"Resource\":\"arn:aws:sqs:eu-west-3:047081014315:foo\",\"Sid\":\"Stmt1611769527792\"}],\"Version\":\"2012-10-17\"}"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewAwsSqsQueuePolicyExpander()
			err := m.Execute(&[]resource.Resource{}, &tt.resourcesFromState)
			if err != nil {
				t.Fatal(err)
			}
			changelog, err := diff.Diff(tt.expected, tt.resourcesFromState)
			if err != nil {
				t.Fatal(err)
			}
			if len(changelog) > 0 {
				for _, change := range changelog {
					t.Errorf("%s got = %v, want %v", strings.Join(change.Path, "."), awsutil.Prettify(change.From), awsutil.Prettify(change.To))
				}
			}
		})
	}
}
