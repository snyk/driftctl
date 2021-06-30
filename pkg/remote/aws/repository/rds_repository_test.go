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

func Test_rdsRepository_ListAllDbSubnetGroups(t *testing.T) {
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
			got, err := r.ListAllDbSubnetGroups()
			assert.Equal(t, tt.wantErr, err)

			if err == nil {
				// Check that results were cached
				cachedData, err := r.ListAllDbSubnetGroups()
				assert.Nil(t, err)
				assert.Equal(t, got, cachedData)
				assert.IsType(t, []*rds.DBSubnetGroup{}, store.Get("rdsListAllDbSubnetGroups"))
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
