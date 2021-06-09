package repository

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"github.com/cloudskiff/driftctl/pkg/remote/cache"
)

type DynamoDBRepository interface {
	ListAllTables() ([]*string, error)
}

type dynamoDBRepository struct {
	client dynamodbiface.DynamoDBAPI
	cache  cache.Cache
}

func NewDynamoDBRepository(session *session.Session, c cache.Cache) *dynamoDBRepository {
	return &dynamoDBRepository{
		dynamodb.New(session),
		c,
	}
}

func (r *dynamoDBRepository) ListAllTables() ([]*string, error) {
	if v := r.cache.Get("dynamodbListAllTables"); v != nil {
		return v.([]*string), nil
	}

	var tables []*string
	input := &dynamodb.ListTablesInput{}
	err := r.client.ListTablesPages(input, func(res *dynamodb.ListTablesOutput, lastPage bool) bool {
		tables = append(tables, res.TableNames...)
		return !lastPage
	})
	if err != nil {
		return nil, err
	}

	r.cache.Put("dynamodbListAllTables", tables)
	return tables, nil
}
