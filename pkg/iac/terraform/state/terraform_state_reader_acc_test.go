package state_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/snyk/driftctl/test"
	"github.com/snyk/driftctl/test/acceptance"
	"github.com/snyk/driftctl/test/acceptance/awsutils"
)

func TestAcc_StateReader_WithMultipleStatesInDirectory(t *testing.T) {
	acceptance.Run(t, acceptance.AccTestCase{
		TerraformVersion: "0.14.9",
		Paths: []string{
			"./testdata/acc/multiple_states_local/s3",
			"./testdata/acc/multiple_states_local/route53",
		},
		Args: []string{
			"scan",
			"--from", "tfstate://testdata/acc/multiple_states_local/states",
			"--filter", "Type=='aws_s3_bucket' || Type=='aws_route53_zone'",
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
	stateBucketName := "driftctl-acc-test-only"
	acceptance.Run(t, acceptance.AccTestCase{
		TerraformVersion: "0.14.9",
		OnStart: func() {
			err := createBucket(stateBucketName)
			if err != nil {
				t.Fatal(err)
			}
		},
		Paths: []string{"./testdata/acc/multiples_states/s3", "./testdata/acc/multiples_states/route53"},
		Args: []string{
			"scan",
			"--from", fmt.Sprintf("tfstate+s3://%s/states", stateBucketName),
			"--filter", "Type=='aws_s3_bucket' || Type=='aws_route53_zone'",
		},
		Checks: []acceptance.AccCheck{
			{
				Check: func(result *test.ScanResult, stdout string, err error) {
					if err != nil {
						t.Fatal(err)
					}
					result.AssertUnmanagedCount(1)
					result.AssertDeletedCount(0)
					result.AssertResourceUnmanaged("driftctl-acc-test-only", "aws_s3_bucket")
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
