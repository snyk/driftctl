package repository

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

type DynamoDBRepository interface {
	ListAllTables() ([]*string, error)
}

type dynamoDBRepository struct {
	client dynamodbiface.DynamoDBAPI
}

func NewDynamoDBRepository(session *session.Session) *dynamoDBRepository {
	return &dynamoDBRepository{
		dynamodb.New(session),
	}
}

func (r *dynamoDBRepository) ListAllTables() ([]*string, error) {
	var tables []*string
	input := &dynamodb.ListTablesInput{}
	err := r.client.ListTablesPages(input, func(res *dynamodb.ListTablesOutput, lastPage bool) bool {
		tables = append(tables, res.TableNames...)
		return !lastPage
	})
	if err != nil {
		return nil, err
	}
	return tables, nil
}
