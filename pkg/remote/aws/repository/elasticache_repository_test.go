package repository

import (
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/pkg/errors"
	"github.com/r3labs/diff/v2"
	"github.com/snyk/driftctl/pkg/remote/cache"
	awstest "github.com/snyk/driftctl/test/aws"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_elasticacheRepository_ListAllCacheClusters(t *testing.T) {
	clusters := []*elasticache.CacheCluster{
		{CacheClusterId: aws.String("cluster1")},
		{CacheClusterId: aws.String("cluster2")},
		{CacheClusterId: aws.String("cluster3")},
		{CacheClusterId: aws.String("cluster4")},
		{CacheClusterId: aws.String("cluster5")},
		{CacheClusterId: aws.String("cluster6")},
	}

	remoteError := errors.New("remote error")

	tests := []struct {
		name    string
		mocks   func(client *awstest.MockFakeElastiCache, store *cache.MockCache)
		want    []*elasticache.CacheCluster
		wantErr error
	}{
		{
			name: "List cache clusters",
			mocks: func(client *awstest.MockFakeElastiCache, store *cache.MockCache) {
				client.On("DescribeCacheClustersPages",
					&elasticache.DescribeCacheClustersInput{},
					mock.MatchedBy(func(callback func(res *elasticache.DescribeCacheClustersOutput, lastPage bool) bool) bool {
						callback(&elasticache.DescribeCacheClustersOutput{
							CacheClusters: clusters[:3],
						}, false)
						callback(&elasticache.DescribeCacheClustersOutput{
							CacheClusters: clusters[3:],
						}, true)
						return true
					})).Return(nil).Once()
				store.On("Get", "elasticacheListAllCacheClusters").Return(nil).Times(1)
				store.On("Put", "elasticacheListAllCacheClusters", clusters).Return(false).Times(1)
			},
			want: clusters,
		},
		{
			name: "should hit cache",
			mocks: func(client *awstest.MockFakeElastiCache, store *cache.MockCache) {
				store.On("Get", "elasticacheListAllCacheClusters").Return(clusters).Times(1)
			},
			want: clusters,
		},
		{
			name: "should return remote error",
			mocks: func(client *awstest.MockFakeElastiCache, store *cache.MockCache) {
				client.On("DescribeCacheClustersPages",
					&elasticache.DescribeCacheClustersInput{},
					mock.AnythingOfType("func(*elasticache.DescribeCacheClustersOutput, bool) bool")).Return(remoteError).Once()
				store.On("Get", "elasticacheListAllCacheClusters").Return(nil).Times(1)
			},
			wantErr: remoteError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &cache.MockCache{}
			client := &awstest.MockFakeElastiCache{}
			tt.mocks(client, store)
			r := &elasticacheRepository{
				client: client,
				cache:  store,
			}
			got, err := r.ListAllCacheClusters()
			assert.Equal(t, tt.wantErr, err)

			changelog, err := diff.Diff(got, tt.want)
			assert.Nil(t, err)
			if len(changelog) > 0 {
				for _, change := range changelog {
					t.Errorf("%s: %s -> %s", strings.Join(change.Path, "."), change.From, change.To)
				}
				t.Fail()
			}
			store.AssertExpectations(t)
			client.AssertExpectations(t)
		})
	}
}
