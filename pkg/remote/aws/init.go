package aws

import (
	"github.com/cloudskiff/driftctl/pkg/alerter"
	"github.com/cloudskiff/driftctl/pkg/output"
	"github.com/cloudskiff/driftctl/pkg/remote/aws/client"
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	"github.com/cloudskiff/driftctl/pkg/remote/cache"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/cloudskiff/driftctl/pkg/terraform"
)

const RemoteAWSTerraform = "aws+tf"

/**
 * Initialize remote (configure credentials, launch tf providers and start gRPC clients)
 * Required to use Scanner
 */

func Init(version string, alerter *alerter.Alerter,
	providerLibrary *terraform.ProviderLibrary,
	supplierLibrary *resource.SupplierLibrary,
	progress output.Progress,
	resourceSchemaRepository *resource.SchemaRepository,
	factory resource.ResourceFactory) error {

	provider, err := NewAWSTerraformProvider(version, progress)
	if err != nil {
		return err
	}
	err = provider.Init()
	if err != nil {
		return err
	}

	repositoryCache := cache.New(100)

	s3Repository := repository.NewS3Repository(client.NewAWSClientFactory(provider.session))
	ec2repository := repository.NewEC2Repository(provider.session, repositoryCache)
	route53repository := repository.NewRoute53Repository(provider.session)
	lambdaRepository := repository.NewLambdaRepository(provider.session)
	rdsRepository := repository.NewRDSRepository(provider.session)
	sqsRepository := repository.NewSQSClient(provider.session)
	snsRepository := repository.NewSNSClient(provider.session)
	dynamoDBRepository := repository.NewDynamoDBRepository(provider.session)
	cloudfrontRepository := repository.NewCloudfrontClient(provider.session)
	kmsRepository := repository.NewKMSRepository(provider.session)
	ecrRepository := repository.NewECRRepository(provider.session)

	deserializer := resource.NewDeserializer(factory)
	providerLibrary.AddProvider(terraform.AWS, provider)

	supplierLibrary.AddSupplier(NewS3BucketSupplier(provider, s3Repository, deserializer))
	supplierLibrary.AddSupplier(NewS3BucketAnalyticSupplier(provider, s3Repository, deserializer))
	supplierLibrary.AddSupplier(NewS3BucketInventorySupplier(provider, s3Repository, deserializer))
	supplierLibrary.AddSupplier(NewS3BucketMetricSupplier(provider, s3Repository, deserializer))
	supplierLibrary.AddSupplier(NewS3BucketNotificationSupplier(provider, s3Repository, deserializer))
	supplierLibrary.AddSupplier(NewS3BucketPolicySupplier(provider, s3Repository, deserializer))
	supplierLibrary.AddSupplier(NewEC2EipSupplier(provider, ec2repository, deserializer))
	supplierLibrary.AddSupplier(NewEC2EipAssociationSupplier(provider, deserializer, ec2repository))
	supplierLibrary.AddSupplier(NewEC2EbsVolumeSupplier(provider, deserializer, ec2repository))
	supplierLibrary.AddSupplier(NewEC2EbsSnapshotSupplier(provider, deserializer, ec2repository))
	supplierLibrary.AddSupplier(NewRoute53ZoneSupplier(provider, deserializer, route53repository))
	supplierLibrary.AddSupplier(NewRoute53RecordSupplier(provider, deserializer, route53repository))
	supplierLibrary.AddSupplier(NewEC2InstanceSupplier(provider, deserializer, ec2repository))
	supplierLibrary.AddSupplier(NewEC2AmiSupplier(provider, deserializer, ec2repository))
	supplierLibrary.AddSupplier(NewEC2KeyPairSupplier(provider, deserializer, ec2repository))
	supplierLibrary.AddSupplier(NewLambdaFunctionSupplier(provider, deserializer, lambdaRepository))
	supplierLibrary.AddSupplier(NewDBSubnetGroupSupplier(provider, deserializer, rdsRepository))
	supplierLibrary.AddSupplier(NewDBInstanceSupplier(provider, deserializer, rdsRepository))
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
	supplierLibrary.AddSupplier(NewSqsQueueSupplier(provider, deserializer, sqsRepository))
	supplierLibrary.AddSupplier(NewSqsQueuePolicySupplier(provider, deserializer, sqsRepository))
	supplierLibrary.AddSupplier(NewSNSTopicSupplier(provider, deserializer, snsRepository))
	supplierLibrary.AddSupplier(NewSNSTopicPolicySupplier(provider, deserializer, snsRepository))
	supplierLibrary.AddSupplier(NewSNSTopicSubscriptionSupplier(provider, alerter, deserializer, snsRepository))
	supplierLibrary.AddSupplier(NewDynamoDBTableSupplier(provider, deserializer, dynamoDBRepository))
	supplierLibrary.AddSupplier(NewRoute53HealthCheckSupplier(provider, deserializer, route53repository))
	supplierLibrary.AddSupplier(NewCloudfrontDistributionSupplier(provider, deserializer, cloudfrontRepository))
	supplierLibrary.AddSupplier(NewECRRepositorySupplier(provider, deserializer, ecrRepository))
	supplierLibrary.AddSupplier(NewKMSKeySupplier(provider, deserializer, kmsRepository))
	supplierLibrary.AddSupplier(NewKMSAliasSupplier(provider, deserializer, kmsRepository))
	supplierLibrary.AddSupplier(NewLambdaEventSourceMappingSupplier(provider, deserializer, lambdaRepository))

	resourceSchemaRepository.Init(provider.Schema())
	aws.InitResourcesMetadata(resourceSchemaRepository)

	return nil
}
