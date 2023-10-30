package state_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/snyk/driftctl/test"
	"github.com/snyk/driftctl/test/acceptance"
	"github.com/snyk/driftctl/test/acceptance/awsutils"
)

func TestAcc_StateReader_WithMultipleStatesInDirectory(t *testing.T) {
	t.Skip("flake")

	acceptance.Run(t, acceptance.AccTestCase{
		TerraformVersion: "0.14.9",
		Paths: []string{
			"./testdata/acc/multiple_states_local/s3",
			"./testdata/acc/multiple_states_local/route53",
		},
		Args: []string{
			"scan",
			"--from", "tfstate://testdata/acc/multiple_states_local/states",
			"--filter", "(Type=='aws_s3_bucket' && Id != 'aws-cloudtrail-logs-994475276861-f6865496') || Type=='aws_route53_zone'",
		},
		Checks: []acceptance.AccCheck{
			{
				Check: func(result *test.ScanResult, stdout string, err error) {
					if err != nil {
						t.Fatal(err)
					}
					result.AssertInfrastructureIsInSync()
					result.AssertManagedCount(2)
					result.Equal("aws_route53_zone", result.Managed()[0].ResourceType())
					result.Equal("aws_s3_bucket", result.Managed()[1].ResourceType())
				},
			},
		},
	})
}

func TestAcc_StateReader_WithMultiplesStatesInS3(t *testing.T) {
	// Disabled since this test is not working
	// terraform_state_reader_acc_test.go:49: OperationAborted: A conflicting conditional operation is currently in progress against this resource. Please try again.
	//     status code: 409, request id: 1TJZX1RZYDZB38CG, host id: laXYB6Z6UXuLXDYYRCXpQOgfSl/PsDGpJFmXpIiDibK17Pd8y4H5aAhyuWd35aqHhnDzyyxj0HE=
	// see https://app.circleci.com/pipelines/github/snyk/driftctl/4279/workflows/360983a0-3253-45b0-8c78-daec16ba73ae/jobs/9402
	t.Skip()
	stateBucketName := "driftctl-acc-statereader-multiples-states"
	acceptance.Run(t, acceptance.AccTestCase{
		TerraformVersion: "0.14.9",
		OnStart: func() {
			err := createBucket(stateBucketName)
			if err != nil {
				t.Fatal(err)
			}
			time.Sleep(30 * time.Second)
		},
		Paths: []string{"./testdata/acc/multiple_states/s3", "./testdata/acc/multiple_states/route53"},
		Args: []string{
			"scan",
			"--from", fmt.Sprintf("tfstate+s3://%s/states", stateBucketName),
			"--filter", "(Type=='aws_s3_bucket' && Id != 'aws-cloudtrail-logs-994475276861-f6865496') || Type=='aws_route53_zone'",
		},
		Checks: []acceptance.AccCheck{
			{
				Check: func(result *test.ScanResult, stdout string, err error) {
					if err != nil {
						t.Fatal(err)
					}
					result.AssertUnmanagedCount(1)
					result.AssertDeletedCount(0)
					result.AssertResourceUnmanaged(stateBucketName, "aws_s3_bucket")
					result.AssertManagedCount(2)
					result.Equal("aws_route53_zone", result.Managed()[0].ResourceType())
					result.Equal("aws_s3_bucket", result.Managed()[1].ResourceType())
				},
			},
		},
		OnEnd: func() {
			err := removeStateBucket(stateBucketName)
			if err != nil {
				t.Fatal(err)
			}
		},
	})
}

func createBucket(bucket string) error {
	client := s3.New(awsutils.Session())
	_, err := client.CreateBucket(&s3.CreateBucketInput{
		Bucket: &bucket,
	})
	if err != nil {
		return err
	}
	return nil
}

func removeStateBucket(bucket string) error {
	client := s3.New(awsutils.Session())
	objects, err := client.ListObjectsV2(&s3.ListObjectsV2Input{Bucket: &bucket})
	if err != nil {
		return err
	}
	for _, object := range objects.Contents {
		_, err := client.DeleteObject(&s3.DeleteObjectInput{
			Bucket: &bucket,
			Key:    object.Key,
		})
		if err != nil {
			return err
		}
	}
	_, err = client.DeleteBucket(&s3.DeleteBucketInput{
		Bucket: &bucket,
	})
	if err != nil {
		return err
	}
	return nil
}
