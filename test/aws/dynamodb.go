package aws

import "github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"

type FakeDynamoDB interface {
	dynamodbiface.DynamoDBAPI
}
