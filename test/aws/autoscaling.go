package aws

import (
	"github.com/aws/aws-sdk-go/service/autoscaling/autoscalingiface"
)

type FakeAutoscaling interface {
	autoscalingiface.AutoScalingAPI
}
