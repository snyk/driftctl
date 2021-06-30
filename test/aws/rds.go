package aws

import "github.com/aws/aws-sdk-go/service/rds/rdsiface"

type FakeRDS interface {
	rdsiface.RDSAPI
}
