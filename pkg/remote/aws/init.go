package aws

import (
	"github.com/cloudskiff/driftctl/pkg/alerter"
	"github.com/cloudskiff/driftctl/pkg/output"
	"github.com/cloudskiff/driftctl/pkg/remote/aws/client"
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/cloudskiff/driftctl/pkg/terraform"
)

const RemoteAWSTerraform = "aws+tf"

/**
 * Initialize remote (configure credentials, launch tf providers and start gRPC clients)
 * Required to use Scanner
 */

func Init(alerter *alerter.Alerter,
	providerLibrary *terraform.ProviderLibrary,
	supplierLibrary *resource.SupplierLibrary,
	progress output.Progress,
	resourceSchemaRepository *resource.SchemaRepository,
	factory resource.ResourceFactory) error {

	provider, err := NewAWSTerraformProvider(progress)
	if err != nil {
		return err
	}
	err = provider.Init()
	if err != nil {
		return err
	}

	s3Repository := repository.NewS3Repository(client.NewAWSClientFactory(provider.session))
	deserializer := resource.NewDeserializer(factory)
	providerLibrary.AddProvider(terraform.AWS, provider)

	supplierLibrary.AddSupplier(NewS3BucketSupplier(provider, s3Repository, deserializer))
	supplierLibrary.AddSupplier(NewS3BucketAnalyticSupplier(provider, s3Repository, deserializer))
	supplierLibrary.AddSupplier(NewS3BucketInventorySupplier(provider, s3Repository, deserializer))
	supplierLibrary.AddSupplier(NewS3BucketMetricSupplier(provider, s3Repository, deserializer))
	supplierLibrary.AddSupplier(NewS3BucketNotificationSupplier(provider, s3Repository, deserializer))
	supplierLibrary.AddSupplier(NewS3BucketPolicySupplier(provider, s3Repository, deserializer))
	supplierLibrary.AddSupplier(NewEC2EipSupplier(provider, deserializer))
	supplierLibrary.AddSupplier(NewEC2EipAssociationSupplier(provider, deserializer))
	supplierLibrary.AddSupplier(NewEC2EbsVolumeSupplier(provider, deserializer))
	supplierLibrary.AddSupplier(NewEC2EbsSnapshotSupplier(provider, deserializer))
	supplierLibrary.AddSupplier(NewRoute53ZoneSupplier(provider, deserializer))
	supplierLibrary.AddSupplier(NewRoute53RecordSupplier(provider, deserializer))
	supplierLibrary.AddSupplier(NewEC2InstanceSupplier(provider, deserializer))
	supplierLibrary.AddSupplier(NewEC2AmiSupplier(provider, deserializer))
	supplierLibrary.AddSupplier(NewEC2KeyPairSupplier(provider, deserializer))
	supplierLibrary.AddSupplier(NewLambdaFunctionSupplier(provider, deserializer))
	supplierLibrary.AddSupplier(NewDBSubnetGroupSupplier(provider, deserializer))
	supplierLibrary.AddSupplier(NewDBInstanceSupplier(provider, deserializer))
	supplierLibrary.AddSupplier(NewVPCSecurityGroupSupplier(provider, deserializer))
	supplierLibrary.AddSupplier(NewIamUserSupplier(provider, deserializer))
	supplierLibrary.AddSupplier(NewIamUserPolicySupplier(provider, deserializer))
	supplierLibrary.AddSupplier(NewIamUserPolicyAttachmentSupplier(provider, deserializer))
	supplierLibrary.AddSupplier(NewIamAccessKeySupplier(provider, deserializer))
	supplierLibrary.AddSupplier(NewIamRoleSupplier(provider, deserializer))
	supplierLibrary.AddSupplier(NewIamPolicySupplier(provider, deserializer))
	supplierLibrary.AddSupplier(NewIamRolePolicySupplier(provider, deserializer))
	supplierLibrary.AddSupplier(NewIamRolePolicyAttachmentSupplier(provider, deserializer))
	supplierLibrary.AddSupplier(NewVPCSecurityGroupRuleSupplier(provider, deserializer))
	supplierLibrary.AddSupplier(NewVPCSupplier(provider, deserializer))
	supplierLibrary.AddSupplier(NewSubnetSupplier(provider, deserializer))
	supplierLibrary.AddSupplier(NewRouteTableSupplier(provider, deserializer))
	supplierLibrary.AddSupplier(NewRouteSupplier(provider, deserializer))
	supplierLibrary.AddSupplier(NewRouteTableAssociationSupplier(provider, deserializer))
	supplierLibrary.AddSupplier(NewNatGatewaySupplier(provider, deserializer))
	supplierLibrary.AddSupplier(NewInternetGatewaySupplier(provider, deserializer))
	supplierLibrary.AddSupplier(NewSqsQueueSupplier(provider, deserializer))
	supplierLibrary.AddSupplier(NewSqsQueuePolicySupplier(provider, deserializer))
	supplierLibrary.AddSupplier(NewSNSTopicSupplier(provider, deserializer))
	supplierLibrary.AddSupplier(NewSNSTopicPolicySupplier(provider, deserializer))
	supplierLibrary.AddSupplier(NewSNSTopicSubscriptionSupplier(provider, alerter, deserializer))
	supplierLibrary.AddSupplier(NewDynamoDBTableSupplier(provider, deserializer))
	supplierLibrary.AddSupplier(NewRoute53HealthCheckSupplier(provider, deserializer))
	supplierLibrary.AddSupplier(NewCloudfrontDistributionSupplier(provider, deserializer))
	supplierLibrary.AddSupplier(NewECRRepositorySupplier(provider, deserializer))
	supplierLibrary.AddSupplier(NewKMSKeySupplier(provider, deserializer))
	supplierLibrary.AddSupplier(NewKMSAliasSupplier(provider, deserializer))
	supplierLibrary.AddSupplier(NewLambdaEventSourceMappingSupplier(provider, deserializer))

	resourceSchemaRepository.Init(provider.Schema())
	aws.InitResourcesMetadata(resourceSchemaRepository)

	return nil
}
