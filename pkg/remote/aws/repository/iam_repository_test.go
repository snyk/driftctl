package repository

import (
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"

	"github.com/stretchr/testify/mock"

	"github.com/r3labs/diff/v2"
	"github.com/stretchr/testify/assert"

	"github.com/cloudskiff/driftctl/mocks"
)

func Test_IAMRepository_ListAllAccessKeys(t *testing.T) {
	tests := []struct {
		name    string
		users   []*iam.User
		mocks   func(client *mocks.FakeIAM)
		want    []*iam.AccessKeyMetadata
		wantErr error
	}{
		{
			name: "List only access keys with multiple pages",
			users: []*iam.User{
				{
					UserName: aws.String("test-driftctl"),
				},
				{
					UserName: aws.String("test-driftctl2"),
				},
			},
			mocks: func(client *mocks.FakeIAM) {

				client.On("ListAccessKeysPages",
					&iam.ListAccessKeysInput{
						UserName: aws.String("test-driftctl"),
					},
					mock.MatchedBy(func(callback func(res *iam.ListAccessKeysOutput, lastPage bool) bool) bool {
						callback(&iam.ListAccessKeysOutput{AccessKeyMetadata: []*iam.AccessKeyMetadata{
							{
								AccessKeyId: aws.String("AKIA5QYBVVD223VWU32A"),
								UserName:    aws.String("test-driftctl"),
							},
						}}, false)
						callback(&iam.ListAccessKeysOutput{AccessKeyMetadata: []*iam.AccessKeyMetadata{
							{
								AccessKeyId: aws.String("AKIA5QYBVVD2QYI36UZP"),
								UserName:    aws.String("test-driftctl"),
							},
						}}, true)
						return true
					})).Return(nil)
				client.On("ListAccessKeysPages",
					&iam.ListAccessKeysInput{
						UserName: aws.String("test-driftctl2"),
					},
					mock.MatchedBy(func(callback func(res *iam.ListAccessKeysOutput, lastPage bool) bool) bool {
						callback(&iam.ListAccessKeysOutput{AccessKeyMetadata: []*iam.AccessKeyMetadata{
							{
								AccessKeyId: aws.String("AKIA5QYBVVD26EJME25D"),
								UserName:    aws.String("test-driftctl2"),
							},
						}}, false)
						callback(&iam.ListAccessKeysOutput{AccessKeyMetadata: []*iam.AccessKeyMetadata{
							{
								AccessKeyId: aws.String("AKIA5QYBVVD2SWDFVVMG"),
								UserName:    aws.String("test-driftctl2"),
							},
						}}, true)
						return true
					})).Return(nil)
			},
			want: []*iam.AccessKeyMetadata{
				{
					AccessKeyId: aws.String("AKIA5QYBVVD223VWU32A"),
					UserName:    aws.String("test-driftctl"),
				},
				{
					AccessKeyId: aws.String("AKIA5QYBVVD2QYI36UZP"),
					UserName:    aws.String("test-driftctl"),
				},
				{
					AccessKeyId: aws.String("AKIA5QYBVVD223VWU32A"),
					UserName:    aws.String("test-driftctl"),
				},
				{
					AccessKeyId: aws.String("AKIA5QYBVVD2QYI36UZP"),
					UserName:    aws.String("test-driftctl"),
				},
				{
					AccessKeyId: aws.String("AKIA5QYBVVD26EJME25D"),
					UserName:    aws.String("test-driftctl2"),
				},
				{
					AccessKeyId: aws.String("AKIA5QYBVVD2SWDFVVMG"),
					UserName:    aws.String("test-driftctl2"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &mocks.FakeIAM{}
			tt.mocks(client)
			r := &iamRepository{
				client: client,
			}
			got, err := r.ListAllAccessKeys(tt.users)
			assert.Equal(t, tt.wantErr, err)
			changelog, err := diff.Diff(got, tt.want)
			assert.Nil(t, err)
			if len(changelog) > 0 {
				for _, change := range changelog {
					t.Errorf("%s: %v -> %v", strings.Join(change.Path, "."), change.From, change.To)
				}
				t.Fail()
			}
		})
	}
}
