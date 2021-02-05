package deserializer

import (
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/sirupsen/logrus"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

type DBInstanceDeserializer struct {
}

func NewDBInstanceDeserializer() *DBInstanceDeserializer {
	return &DBInstanceDeserializer{}
}

func (s DBInstanceDeserializer) HandledType() resource.ResourceType {
	return aws.AwsDbInstanceResourceType
}

func (s DBInstanceDeserializer) Deserialize(rawResourceList []cty.Value) ([]resource.Resource, error) {
	resources := make([]resource.Resource, 0)
	for _, raw := range rawResourceList {
		raw := raw
		res, err := decodeDBInstance(&raw)
		if err != nil {
			logrus.Warnf("error when deserializing aws_db_instance %+v : %+v", raw, err)
		}
		resources = append(resources, res)
	}
	return resources, nil
}

func decodeDBInstance(res *cty.Value) (resource.Resource, error) {
	var decoded aws.AwsDbInstance
	if err := gocty.FromCtyValue(*res, &decoded); err != nil {
		return nil, err
	}

	// On first apply this field is null in state and set to empty array in next terraform refresh
	if decoded.SecurityGroupNames != nil && len(*decoded.SecurityGroupNames) == 0 {
		decoded.SecurityGroupNames = nil
	}

	// On first apply this field is null in state and set to empty array in next terraform refresh
	if decoded.EnabledCloudwatchLogsExports != nil && len(*decoded.EnabledCloudwatchLogsExports) == 0 {
		decoded.EnabledCloudwatchLogsExports = nil
	}

	return &decoded, nil
}
