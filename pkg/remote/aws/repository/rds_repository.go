package repository

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/aws/aws-sdk-go/service/rds/rdsiface"
	"github.com/cloudskiff/driftctl/pkg/remote/cache"
)

type RDSClient interface {
	rdsiface.RDSAPI
}

type RDSRepository interface {
	ListAllDBInstances() ([]*rds.DBInstance, error)
	ListAllDbSubnetGroups() ([]*rds.DBSubnetGroup, error)
}

type rdsRepository struct {
	client rdsiface.RDSAPI
	cache  cache.Cache
}

func NewRDSRepository(session *session.Session, c cache.Cache) *rdsRepository {
	return &rdsRepository{
		rds.New(session),
		c,
	}
}

func (r *rdsRepository) ListAllDBInstances() ([]*rds.DBInstance, error) {
	if v := r.cache.Get("rdsListAllDBInstances"); v != nil {
		return v.([]*rds.DBInstance), nil
	}

	var result []*rds.DBInstance
	input := &rds.DescribeDBInstancesInput{}
	err := r.client.DescribeDBInstancesPages(input, func(res *rds.DescribeDBInstancesOutput, lastPage bool) bool {
		result = append(result, res.DBInstances...)
		return !lastPage
	})
	if err != nil {
		return nil, err
	}

	r.cache.Put("rdsListAllDBInstances", result)
	return result, nil
}

func (r *rdsRepository) ListAllDbSubnetGroups() ([]*rds.DBSubnetGroup, error) {
	if v := r.cache.Get("rdsListAllDbSubnetGroups"); v != nil {
		return v.([]*rds.DBSubnetGroup), nil
	}

	var subnetGroups []*rds.DBSubnetGroup
	input := rds.DescribeDBSubnetGroupsInput{}
	err := r.client.DescribeDBSubnetGroupsPages(&input,
		func(resp *rds.DescribeDBSubnetGroupsOutput, lastPage bool) bool {
			subnetGroups = append(subnetGroups, resp.DBSubnetGroups...)
			return !lastPage
		},
	)

	r.cache.Put("rdsListAllDbSubnetGroups", subnetGroups)
	return subnetGroups, err
}
