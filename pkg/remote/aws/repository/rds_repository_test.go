package repository

import (
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/cloudskiff/driftctl/pkg/remote/cache"
	awstest "github.com/cloudskiff/driftctl/test/aws"
	"github.com/r3labs/diff/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_rdsRepository_ListAllDBInstances(t *testing.T) {
	tests := []struct {
		name    string
		mocks   func(client *awstest.MockFakeRDS)
		want    []*rds.DBInstance
		wantErr error
	}{
		{
			name: "List with 2 pages",
			mocks: func(client *awstest.MockFakeRDS) {
				client.On("DescribeDBInstancesPages",
					&rds.DescribeDBInstancesInput{},
					mock.MatchedBy(func(callback func(res *rds.DescribeDBInstancesOutput, lastPage bool) bool) bool {
						callback(&rds.DescribeDBInstancesOutput{
							DBInstances: []*rds.DBInstance{
								{DBInstanceIdentifier: aws.String("1")},
								{DBInstanceIdentifier: aws.String("2")},
								{DBInstanceIdentifier: aws.String("3")},
							},
						}, false)
						callback(&rds.DescribeDBInstancesOutput{
							DBInstances: []*rds.DBInstance{
								{DBInstanceIdentifier: aws.String("4")},
								{DBInstanceIdentifier: aws.String("5")},
								{DBInstanceIdentifier: aws.String("6")},
							},
						}, true)
						return true
					})).Return(nil).Once()
			},
			want: []*rds.DBInstance{
				{DBInstanceIdentifier: aws.String("1")},
				{DBInstanceIdentifier: aws.String("2")},
				{DBInstanceIdentifier: aws.String("3")},
				{DBInstanceIdentifier: aws.String("4")},
				{DBInstanceIdentifier: aws.String("5")},
				{DBInstanceIdentifier: aws.String("6")},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := cache.New(1)
			client := &awstest.MockFakeRDS{}
			tt.mocks(client)
			r := &rdsRepository{
				client: client,
				cache:  store,
			}
			got, err := r.ListAllDBInstances()
			assert.Equal(t, tt.wantErr, err)

			if err == nil {
				// Check that results were cached
				cachedData, err := r.ListAllDBInstances()
				assert.Nil(t, err)
				assert.Equal(t, got, cachedData)
				assert.IsType(t, []*rds.DBInstance{}, store.Get("rdsListAllDBInstances"))
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

func Test_rdsRepository_ListAllDBSubnetGroups(t *testing.T) {
	tests := []struct {
		name    string
		mocks   func(client *awstest.MockFakeRDS)
		want    []*rds.DBSubnetGroup
		wantErr error
	}{
		{
			name: "List with 2 pages",
			mocks: func(client *awstest.MockFakeRDS) {
				client.On("DescribeDBSubnetGroupsPages",
					&rds.DescribeDBSubnetGroupsInput{},
					mock.MatchedBy(func(callback func(res *rds.DescribeDBSubnetGroupsOutput, lastPage bool) bool) bool {
						callback(&rds.DescribeDBSubnetGroupsOutput{
							DBSubnetGroups: []*rds.DBSubnetGroup{
								{DBSubnetGroupName: aws.String("1")},
								{DBSubnetGroupName: aws.String("2")},
								{DBSubnetGroupName: aws.String("3")},
							},
						}, false)
						callback(&rds.DescribeDBSubnetGroupsOutput{
							DBSubnetGroups: []*rds.DBSubnetGroup{
								{DBSubnetGroupName: aws.String("4")},
								{DBSubnetGroupName: aws.String("5")},
								{DBSubnetGroupName: aws.String("6")},
							},
						}, true)
						return true
					})).Return(nil).Once()
			},
			want: []*rds.DBSubnetGroup{
				{DBSubnetGroupName: aws.String("1")},
				{DBSubnetGroupName: aws.String("2")},
				{DBSubnetGroupName: aws.String("3")},
				{DBSubnetGroupName: aws.String("4")},
				{DBSubnetGroupName: aws.String("5")},
				{DBSubnetGroupName: aws.String("6")},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := cache.New(1)
			client := &awstest.MockFakeRDS{}
			tt.mocks(client)
			r := &rdsRepository{
				client: client,
				cache:  store,
			}
			got, err := r.ListAllDBSubnetGroups()
			assert.Equal(t, tt.wantErr, err)

			if err == nil {
				// Check that results were cached
				cachedData, err := r.ListAllDBSubnetGroups()
				assert.Nil(t, err)
				assert.Equal(t, got, cachedData)
				assert.IsType(t, []*rds.DBSubnetGroup{}, store.Get("rdsListAllDBSubnetGroups"))
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

func Test_rdsRepository_ListAllDBClusters(t *testing.T) {
	tests := []struct {
		name    string
		mocks   func(*awstest.MockFakeRDS, *cache.MockCache)
		want    []*rds.DBCluster
		wantErr error
	}{
		{
			name: "should list with 2 pages",
			mocks: func(client *awstest.MockFakeRDS, store *cache.MockCache) {
				clusters := []*rds.DBCluster{
					{DBClusterIdentifier: aws.String("1")},
					{DBClusterIdentifier: aws.String("2")},
					{DBClusterIdentifier: aws.String("3")},
					{DBClusterIdentifier: aws.String("4")},
					{DBClusterIdentifier: aws.String("5")},
					{DBClusterIdentifier: aws.String("6")},
				}

				client.On("DescribeDBClustersPages",
					&rds.DescribeDBClustersInput{},
					mock.MatchedBy(func(callback func(res *rds.DescribeDBClustersOutput, lastPage bool) bool) bool {
						callback(&rds.DescribeDBClustersOutput{DBClusters: clusters[:3]}, false)
						callback(&rds.DescribeDBClustersOutput{DBClusters: clusters[3:]}, true)
						return true
					})).Return(nil).Once()

				store.On("Get", "rdsListAllDBClusters").Return(nil).Once()
				store.On("Put", "rdsListAllDBClusters", clusters).Return(false).Once()
			},
			want: []*rds.DBCluster{
				{DBClusterIdentifier: aws.String("1")},
				{DBClusterIdentifier: aws.String("2")},
				{DBClusterIdentifier: aws.String("3")},
				{DBClusterIdentifier: aws.String("4")},
				{DBClusterIdentifier: aws.String("5")},
				{DBClusterIdentifier: aws.String("6")},
			},
		},
		{
			name: "should hit cache",
			mocks: func(client *awstest.MockFakeRDS, store *cache.MockCache) {
				clusters := []*rds.DBCluster{
					{DBClusterIdentifier: aws.String("1")},
					{DBClusterIdentifier: aws.String("2")},
					{DBClusterIdentifier: aws.String("3")},
					{DBClusterIdentifier: aws.String("4")},
					{DBClusterIdentifier: aws.String("5")},
					{DBClusterIdentifier: aws.String("6")},
				}

				store.On("Get", "rdsListAllDBClusters").Return(clusters).Once()
			},
			want: []*rds.DBCluster{
				{DBClusterIdentifier: aws.String("1")},
				{DBClusterIdentifier: aws.String("2")},
				{DBClusterIdentifier: aws.String("3")},
				{DBClusterIdentifier: aws.String("4")},
				{DBClusterIdentifier: aws.String("5")},
				{DBClusterIdentifier: aws.String("6")},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &cache.MockCache{}
			client := &awstest.MockFakeRDS{}
			tt.mocks(client, store)
			r := &rdsRepository{
				client: client,
				cache:  store,
			}
			got, err := r.ListAllDBClusters()
			assert.Equal(t, tt.wantErr, err)

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
