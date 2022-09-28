package repository

import (
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/service/s3control"
	"github.com/snyk/driftctl/enumeration/remote/aws/client"
	"github.com/snyk/driftctl/enumeration/remote/cache"
	"github.com/stretchr/testify/mock"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/r3labs/diff/v2"
	awstest "github.com/snyk/driftctl/test/aws"
	"github.com/stretchr/testify/assert"
)

func Test_s3ControlRepository_DescribeAccountPublicAccessBlock(t *testing.T) {

	tests := []struct {
		name    string
		mocks   func(client *awstest.MockFakeS3Control)
		want    *s3control.PublicAccessBlockConfiguration
		wantErr error
	}{
		{
			name: "describe account public accessblock",
			mocks: func(client *awstest.MockFakeS3Control) {
				client.On("GetPublicAccessBlock", mock.Anything).Return(
					&s3control.GetPublicAccessBlockOutput{
						PublicAccessBlockConfiguration: &s3control.PublicAccessBlockConfiguration{
							BlockPublicAcls:       aws.Bool(false),
							BlockPublicPolicy:     aws.Bool(true),
							IgnorePublicAcls:      aws.Bool(false),
							RestrictPublicBuckets: aws.Bool(true),
						},
					},
					nil,
				).Once()
			},
			want: &s3control.PublicAccessBlockConfiguration{
				BlockPublicAcls:       aws.Bool(false),
				BlockPublicPolicy:     aws.Bool(true),
				IgnorePublicAcls:      aws.Bool(false),
				RestrictPublicBuckets: aws.Bool(true),
			},
		},
		{
			name: "Error detting account public accessblock",
			mocks: func(client *awstest.MockFakeS3Control) {
				client.On("GetPublicAccessBlock", mock.Anything).Return(
					nil,
					awserr.NewRequestFailure(nil, 403, ""),
				).Once()
			},
			want:    nil,
			wantErr: awserr.NewRequestFailure(nil, 403, ""),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := cache.New(1)
			mockedClient := &awstest.MockFakeS3Control{}
			tt.mocks(mockedClient)
			factory := client.MockAwsClientFactoryInterface{}
			factory.On("GetS3ControlClient", (*aws.Config)(nil)).Return(mockedClient).Once()
			r := NewS3ControlRepository(&factory, "", store)
			got, err := r.DescribeAccountPublicAccessBlock()
			factory.AssertExpectations(t)
			assert.Equal(t, tt.wantErr, err)

			if err == nil {
				// Check that results were cached
				cachedData, err := r.DescribeAccountPublicAccessBlock()
				assert.NoError(t, err)
				assert.Equal(t, got, cachedData)
				assert.IsType(t, &s3control.PublicAccessBlockConfiguration{}, store.Get("S3DescribeAccountPublicAccessBlock"))
			}

			changelog, err := diff.Diff(got, tt.want)
			assert.Nil(t, err)
			if len(changelog) > 0 {
				for _, change := range changelog {
					t.Errorf("%s: %s -> %s", strings.Join(change.Path, "."), change.From, change.To)
				}
				t.Fail()
			}
		})
	}
}