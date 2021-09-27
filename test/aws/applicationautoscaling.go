package aws

import (
	"github.com/aws/aws-sdk-go/service/applicationautoscaling/applicationautoscalingiface"
)

type FakeApplicationAutoScaling interface {
	applicationautoscalingiface.ApplicationAutoScalingAPI
}
