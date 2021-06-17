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
	factory resource.ResourceFactory,
	configDir string) error {

	if version == "" {
		version = "3.19.0"
	}
	provider, err := NewAWSTerraformProvider(version, progress, configDir)
	if err != nil {
		return err
	}
	err = provider.Init()
	if err != nil {
		return err
	}

	repositoryCache := cache.New(100)

	s3Repository := repository.NewS3Repository(client.NewAWSClientFactory(provider.session), repositoryCache)
	ec2repository := repository.NewEC2Repository(provider.session, repositoryCache)
	route53repository := repository.NewRoute53Repository(provider.session, repositoryCache)
	lambdaRepository := repository.NewLambdaRepository(provider.session, repositoryCache)
	rdsRepository := repository.NewRDSRepository(provider.session, repositoryCache)
	sqsRepository := repository.NewSQSClient(provider.session, repositoryCache)
	snsRepository := repository.NewSNSClient(provider.session, repositoryCache)
	cloudfrontRepository := repository.NewCloudfrontClient(provider.session, repositoryCache)
	dynamoDBRepository := repository.NewDynamoDBRepository(provider.session, repositoryCache)
	ecrRepository := repository.NewECRRepository(provider.session, repositoryCache)
	kmsRepository := repository.NewKMSRepository(provider.session, repositoryCache)
	iamRepository := repository.NewIAMRepository(provider.session, repositoryCache)

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
	supplierLibrary.AddSupplier(NewVPCSecurityGroupSupplier(provider, deserializer, ec2repository))
	supplierLibrary.AddSupplier(NewIamUserSupplier(provider, deserializer, iamRepository))
	supplierLibrary.AddSupplier(NewIamUserPolicySupplier(provider, deserializer, iamRepository))
	supplierLibrary.AddSupplier(NewIamUserPolicyAttachmentSupplier(provider, deserializer, iamRepository))
	supplierLibrary.AddSupplier(NewIamAccessKeySupplier(provider, deserializer, iamRepository))
	supplierLibrary.AddSupplier(NewIamRoleSupplier(provider, deserializer, iamRepository))
	supplierLibrary.AddSupplier(NewIamPolicySupplier(provider, deserializer, iamRepository))
	supplierLibrary.AddSupplier(NewIamRolePolicySupplier(provider, deserializer, iamRepository))
	supplierLibrary.AddSupplier(NewIamRolePolicyAttachmentSupplier(provider, deserializer, iamRepository))
	supplierLibrary.AddSupplier(NewVPCSecurityGroupRuleSupplier(provider, deserializer, ec2repository))
	supplierLibrary.AddSupplier(NewVPCSupplier(provider, deserializer, ec2repository))
	supplierLibrary.AddSupplier(NewSubnetSupplier(provider, deserializer, ec2repository))
	supplierLibrary.AddSupplier(NewRouteTableSupplier(provider, deserializer, ec2repository))
	supplierLibrary.AddSupplier(NewRouteSupplier(provider, deserializer, ec2repository))
	supplierLibrary.AddSupplier(NewRouteTableAssociationSupplier(provider, deserializer, ec2repository))
	supplierLibrary.AddSupplier(NewNatGatewaySupplier(provider, deserializer, ec2repository))
	supplierLibrary.AddSupplier(NewInternetGatewaySupplier(provider, deserializer, ec2repository))
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

	err = resourceSchemaRepository.Init(version, provider.Schema())
	if err != nil {
		return err
	}
	aws.InitResourcesMetadata(resourceSchemaRepository)

	return nil
}
