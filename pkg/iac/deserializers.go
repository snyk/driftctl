package iac

import (
	"github.com/cloudskiff/driftctl/pkg/remote/deserializer"
	awsdeserializer "github.com/cloudskiff/driftctl/pkg/resource/aws/deserializer"
)

func Deserializers() []deserializer.CTYDeserializer {
	return []deserializer.CTYDeserializer{
		awsdeserializer.NewS3BucketDeserializer(),
		awsdeserializer.NewS3BucketAnalyticDeserializer(),
		awsdeserializer.NewS3BucketInventoryDeserializer(),
		awsdeserializer.NewS3BucketMetricDeserializer(),
		awsdeserializer.NewS3BucketNotificationDeserializer(),
		awsdeserializer.NewS3BucketPolicyDeserializer(),
		awsdeserializer.NewRoute53ZoneDeserializer(),
		awsdeserializer.NewRoute53RecordDeserializer(),
		awsdeserializer.NewEC2EipDeserializer(),
		awsdeserializer.NewEC2EipAssociationDeserializer(),
		awsdeserializer.NewEC2EbsVolumeDeserializer(),
		awsdeserializer.NewEC2EbsSnapshotDeserializer(),
		awsdeserializer.NewEC2InstanceDeserializer(),
		awsdeserializer.NewEC2AmiDeserializer(),
		awsdeserializer.NewEC2KeyPairDeserializer(),
		awsdeserializer.NewDBInstanceDeserializer(),
		awsdeserializer.NewLambdaFunctionDeserializer(),
		awsdeserializer.NewDBSubnetGroupDeserializer(),
		awsdeserializer.NewVPCSecurityGroupDeserializer(),
		awsdeserializer.NewDefaultSecurityGroupDeserializer(),
		awsdeserializer.NewIamUserDeserializer(),
		awsdeserializer.NewIamUserPolicyDeserializer(),
		awsdeserializer.NewIamUserPolicyAttachmentDeserializer(),
		awsdeserializer.NewIamAccessKeyDeserializer(),
		awsdeserializer.NewIamRoleDeserializer(),
		awsdeserializer.NewIamPolicyDeserializer(),
		awsdeserializer.NewIamPolicyAttachmentDeserializer(),
		awsdeserializer.NewIamRolePolicyDeserializer(),
		awsdeserializer.NewIamRolePolicyAttachmentDeserializer(),
		awsdeserializer.NewVPCSecurityGroupRuleDeserializer(),
		awsdeserializer.NewDefaultVPCDeserializer(),
		awsdeserializer.NewVPCDeserializer(),
		awsdeserializer.NewDefaultSubnetDeserializer(),
		awsdeserializer.NewSubnetDeserializer(),
		awsdeserializer.NewDefaultRouteTableDeserializer(),
		awsdeserializer.NewRouteTableDeserializer(),
		awsdeserializer.NewRouteDeserializer(),
		awsdeserializer.NewRouteTableAssociationDeserializer(),
		awsdeserializer.NewNatGatewayDeserializer(),
		awsdeserializer.NewInternetGatewayDeserializer(),
		awsdeserializer.NewSqsQueueDeserializer(),
	}
}
