package repository

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/aws/aws-sdk-go/service/rds/rdsiface"
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
}

func NewRDSRepository(session *session.Session) *rdsRepository {
	return &rdsRepository{
		rds.New(session),
	}
}

func (r *rdsRepository) ListAllDBInstances() ([]*rds.DBInstance, error) {
	var result []*rds.DBInstance
	input := &rds.DescribeDBInstancesInput{}
	err := r.client.DescribeDBInstancesPages(input, func(res *rds.DescribeDBInstancesOutput, lastPage bool) bool {
		result = append(result, res.DBInstances...)
		return !lastPage
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (r *rdsRepository) ListAllDbSubnetGroups() ([]*rds.DBSubnetGroup, error) {
	var subnetGroups []*rds.DBSubnetGroup
	input := rds.DescribeDBSubnetGroupsInput{}
	err := r.client.DescribeDBSubnetGroupsPages(&input,
		func(resp *rds.DescribeDBSubnetGroupsOutput, lastPage bool) bool {
			subnetGroups = append(subnetGroups, resp.DBSubnetGroups...)
			return !lastPage
		},
	)
	return subnetGroups, err
}
