package enumerator

import (
	"errors"
	"reflect"
	"testing"

	awssdk "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/cloudskiff/driftctl/mocks"
	"github.com/cloudskiff/driftctl/pkg/iac/config"
	"github.com/stretchr/testify/mock"
)

func TestS3Enumerator_Enumerate(t *testing.T) {
	tests := []struct {
		name   string
		config config.SupplierConfig
		mocks  func(client *mocks.FakeS3)
		want   []string
		err    string
	}{
		{
			name: "test results are returned",
			config: config.SupplierConfig{
				Path: "bucket-name/a/nested/prefix",
			},
			mocks: func(client *mocks.FakeS3) {
				input := &s3.ListObjectsV2Input{
					Bucket: awssdk.String("bucket-name"),
					Prefix: awssdk.String("a/nested/prefix"),
				}
				client.On(
					"ListObjectsV2Pages",
					input,
					mock.MatchedBy(func(callback func(res *s3.ListObjectsV2Output, lastPage bool) bool) bool {
						callback(&s3.ListObjectsV2Output{
							Contents: []*s3.Object{
								{
									Key: awssdk.String("a/nested/prefix/state1"),
								},
								{
									Key: awssdk.String("a/nested/prefix/state2"),
								},
								{
									Key: awssdk.String("a/nested/prefix/state3"),
								},
							},
						}, false)
						callback(&s3.ListObjectsV2Output{
							Contents: []*s3.Object{
								{
									Key: awssdk.String("a/nested/prefix/state4"),
								},
								{
									Key: awssdk.String("a/nested/prefix/state5"),
								},
								{
									Key: awssdk.String("a/nested/prefix/state6"),
								},
							},
						}, true)
						return true
					}),
				).Return(nil)
			},
			want: []string{
				"bucket-name/a/nested/prefix/state1",
				"bucket-name/a/nested/prefix/state2",
				"bucket-name/a/nested/prefix/state3",
				"bucket-name/a/nested/prefix/state4",
				"bucket-name/a/nested/prefix/state5",
				"bucket-name/a/nested/prefix/state6",
			},
		},
		{
			name: "test when invalid config used",
			config: config.SupplierConfig{
				Path: "bucket-name",
			},
			mocks: func(client *mocks.FakeS3) {
				client.On("ListObjectsV2Pages", mock.Anything, mock.Anything).Return(errors.New("error when listing"))
			},
			want: nil,
			err:  "Unable to parse S3 path: bucket-name. Must be BUCKET_NAME/PREFIX",
		},
		{
			name:   "test when empty config used",
			config: config.SupplierConfig{},
			mocks: func(client *mocks.FakeS3) {
				client.On("ListObjectsV2Pages", mock.Anything, mock.Anything).Return(errors.New("error when listing"))
			},
			want: nil,
			err:  "Unable to parse S3 path: . Must be BUCKET_NAME/PREFIX",
		},
		{
			name: "test enumeration return error",
			config: config.SupplierConfig{
				Path: "bucket-name/a/nested/prefix",
			},
			mocks: func(client *mocks.FakeS3) {
				client.On("ListObjectsV2Pages", mock.Anything, mock.Anything).Return(errors.New("error when listing"))
			},
			want: nil,
			err:  "error when listing",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeS3 := mocks.FakeS3{}
			tt.mocks(&fakeS3)
			s := &S3Enumerator{
				config: tt.config,
				client: &fakeS3,
			}
			got, err := s.Enumerate()
			if err != nil && err.Error() != tt.err {
				t.Fatalf("Expected error '%s', got '%s'", tt.err, err.Error())
			}
			if err != nil && tt.err == "" {
				t.Fatalf("Expected error '%s' but got nil", tt.err)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Enumerate() got = %v, want %v", got, tt.want)
			}
		})
	}
}
