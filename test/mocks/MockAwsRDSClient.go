package mocks

import (
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/aws/aws-sdk-go/service/rds/rdsiface"
)

type DescribeSubnetGroupResponse []struct {
	LastPage bool
	Response *rds.DescribeDBSubnetGroupsOutput
}

type DescribeDBInstancesPagesOutput []struct {
	LastPage bool
	Response *rds.DescribeDBInstancesOutput
}

type MockAWSRDSClient struct {
	rdsiface.RDSAPI
	dbInstancesPages            DescribeDBInstancesPagesOutput
	describeSubnetGroupResponse DescribeSubnetGroupResponse
	err                         error
}

func NewMockAWSRDSErrorClient(err error) *MockAWSRDSClient {
	return &MockAWSRDSClient{err: err}
}

func NewMockAWSRDSClient(dbInstancesPages DescribeDBInstancesPagesOutput) *MockAWSRDSClient {
	return &MockAWSRDSClient{dbInstancesPages: dbInstancesPages}
}

func NewMockAWSRDSSubnetGroupClient(describeSubnetGroupResponse DescribeSubnetGroupResponse) *MockAWSRDSClient {
	return &MockAWSRDSClient{describeSubnetGroupResponse: describeSubnetGroupResponse}
}

func (m *MockAWSRDSClient) DescribeDBInstancesPages(_ *rds.DescribeDBInstancesInput, cb func(*rds.DescribeDBInstancesOutput, bool) bool) error {
	if m.err != nil {
		return m.err
	}
	for _, dbInstancesPage := range m.dbInstancesPages {
		cb(dbInstancesPage.Response, dbInstancesPage.LastPage)
	}
	return nil
}

func (m *MockAWSRDSClient) DescribeDBSubnetGroupsPages(input *rds.DescribeDBSubnetGroupsInput, callback func(*rds.DescribeDBSubnetGroupsOutput, bool) bool) error {
	if m.err != nil {
		return m.err
	}
	for _, response := range m.describeSubnetGroupResponse {
		callback(response.Response, response.LastPage)
	}
	return nil
}
