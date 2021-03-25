package deserializer

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

type S3BucketMetricDeserializer struct {
}

func NewS3BucketMetricDeserializer() *S3BucketMetricDeserializer {
	return &S3BucketMetricDeserializer{}
}

func (s S3BucketMetricDeserializer) HandledType() resource.ResourceType {
	return aws.AwsS3BucketMetricResourceType
}

func (s S3BucketMetricDeserializer) Deserialize(rawMetrics []cty.Value) ([]resource.Resource, error) {
	var metrics []resource.Resource
	for _, metric := range rawMetrics {
		metric := metric
		var me aws.AwsS3BucketMetric
		if err := gocty.FromCtyValue(metric, &me); err == nil {
			me.CtyVal = &metric
			metrics = append(metrics, &me)
		} else {
			logrus.Warnf("Cannot read s3 bucket metric %s: %+v", metric.GoString(), err)
		}
	}
	return metrics, nil
}
