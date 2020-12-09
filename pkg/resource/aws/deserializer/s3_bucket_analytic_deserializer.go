package deserializer

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

type S3BucketAnalyticDeserializer struct {
}

func NewS3BucketAnalyticDeserializer() *S3BucketAnalyticDeserializer {
	return &S3BucketAnalyticDeserializer{}
}

func (s S3BucketAnalyticDeserializer) HandledType() resource.ResourceType {
	return aws.AwsS3BucketAnalyticsConfigurationResourceType
}

func (s S3BucketAnalyticDeserializer) Deserialize(rawAnalytic []cty.Value) ([]resource.Resource, error) {
	var inventories []resource.Resource
	for _, analytic := range rawAnalytic {
		var inv aws.AwsS3BucketAnalyticsConfiguration
		if err := gocty.FromCtyValue(analytic, &inv); err == nil {
			inventories = append(inventories, &inv)
		} else {
			logrus.Warnf("Cannot read s3 bucket analytic %s: %+v", analytic.GoString(), err)
		}
	}
	return inventories, nil
}
