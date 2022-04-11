package aws

import (
	"github.com/aws/aws-sdk-go/service/elasticache/elasticacheiface"
)

type FakeElastiCache interface {
	elasticacheiface.ElastiCacheAPI
}
