package deserializer

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

type S3BucketNotificationDeserializer struct {
}

func NewS3BucketNotificationDeserializer() *S3BucketNotificationDeserializer {
	return &S3BucketNotificationDeserializer{}
}

func (s S3BucketNotificationDeserializer) HandledType() resource.ResourceType {
	return aws.AwsS3BucketNotificationResourceType
}

func (s S3BucketNotificationDeserializer) Deserialize(rawNotification []cty.Value) ([]resource.Resource, error) {
	var inventories []resource.Resource
	for _, notification := range rawNotification {
		notification := notification
		var inv aws.AwsS3BucketNotification
		if err := gocty.FromCtyValue(notification, &inv); err == nil {
			inv.CtyVal = &notification
			inventories = append(inventories, &inv)
		} else {
			logrus.Warnf("Cannot read s3 bucket notification %s: %+v", notification.GoString(), err)
		}
	}
	return inventories, nil
}
