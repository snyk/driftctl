package repository

import (
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/cloudskiff/driftctl/pkg/remote/cache"
	awstest "github.com/cloudskiff/driftctl/test/aws"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/stretchr/testify/mock"

	"github.com/r3labs/diff/v2"
	"github.com/stretchr/testify/assert"
)

func Test_ecrRepository_ListAllRepositories(t *testing.T) {

	tests := []struct {
		name    string
		mocks   func(client *awstest.MockFakeECR)
		want    []*ecr.Repository
		wantErr error
	}{
		{
			name: "List with 2 pages",
			mocks: func(client *awstest.MockFakeECR) {
				client.On("DescribeRepositoriesPages",
					&ecr.DescribeRepositoriesInput{},
					mock.MatchedBy(func(callback func(res *ecr.DescribeRepositoriesOutput, lastPage bool) bool) bool {
						callback(&ecr.DescribeRepositoriesOutput{
							Repositories: []*ecr.Repository{
								{RepositoryName: aws.String("1")},
								{RepositoryName: aws.String("2")},
								{RepositoryName: aws.String("3")},
							},
						}, false)
						callback(&ecr.DescribeRepositoriesOutput{
							Repositories: []*ecr.Repository{
								{RepositoryName: aws.String("4")},
								{RepositoryName: aws.String("5")},
								{RepositoryName: aws.String("6")},
							},
						}, true)
						return true
					})).Return(nil).Once()
			},
			want: []*ecr.Repository{
				{RepositoryName: aws.String("1")},
				{RepositoryName: aws.String("2")},
				{RepositoryName: aws.String("3")},
				{RepositoryName: aws.String("4")},
				{RepositoryName: aws.String("5")},
				{RepositoryName: aws.String("6")},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := cache.New(1)
			client := awstest.MockFakeECR{}
			tt.mocks(&client)
			r := &ecrRepository{
				client: &client,
				cache:  store,
			}
			got, err := r.ListAllRepositories()
			assert.Equal(t, tt.wantErr, err)

			if err == nil {
				// Check that results were cached
				cachedData, err := r.ListAllRepositories()
				assert.NoError(t, err)
				assert.Equal(t, got, cachedData)
				assert.IsType(t, []*ecr.Repository{}, store.Get("ecrListAllRepositories"))
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
