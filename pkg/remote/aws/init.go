package aws

import (
	"github.com/cloudskiff/driftctl/pkg/alerter"
	"github.com/cloudskiff/driftctl/pkg/output"
	"github.com/cloudskiff/driftctl/pkg/remote/aws/client"
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	"github.com/cloudskiff/driftctl/pkg/remote/cache"
	"github.com/cloudskiff/driftctl/pkg/remote/common"
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
	remoteLibrary *common.RemoteLibrary,
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
	sqsRepository := repository.NewSQSRepository(provider.session, repositoryCache)
	snsRepository := repository.NewSNSRepository(provider.session, repositoryCache)
	cloudfrontRepository := repository.NewCloudfrontRepository(provider.session, repositoryCache)
	dynamoDBRepository := repository.NewDynamoDBRepository(provider.session, repositoryCache)
	ecrRepository := repository.NewECRRepository(provider.session, repositoryCache)
	kmsRepository := repository.NewKMSRepository(provider.session, repositoryCache)
	iamRepository := repository.NewIAMRepository(provider.session, repositoryCache)

	deserializer := resource.NewDeserializer(factory)
	providerLibrary.AddProvider(terraform.AWS, provider)

	remoteLibrary.AddEnumerator(NewS3BucketEnumerator(s3Repository, factory, provider.Config))
	remoteLibrary.AddDetailsFetcher(aws.AwsS3BucketResourceType, NewS3BucketDetailsFetcher(provider, deserializer))
	remoteLibrary.AddEnumerator(NewS3BucketInventoryEnumerator(s3Repository, factory, provider.Config))
	remoteLibrary.AddDetailsFetcher(aws.AwsS3BucketInventoryResourceType, NewS3BucketInventoryDetailsFetcher(provider, deserializer))
	remoteLibrary.AddEnumerator(NewS3BucketNotificationEnumerator(s3Repository, factory, provider.Config))
	remoteLibrary.AddDetailsFetcher(aws.AwsS3BucketNotificationResourceType, NewS3BucketNotificationDetailsFetcher(provider, deserializer))
	remoteLibrary.AddEnumerator(NewS3BucketMetricsEnumerator(s3Repository, factory, provider.Config))
	remoteLibrary.AddDetailsFetcher(aws.AwsS3BucketMetricResourceType, NewS3BucketMetricsDetailsFetcher(provider, deserializer))

	remoteLibrary.AddEnumerator(NewEC2EbsVolumeEnumerator(ec2repository, factory))
	remoteLibrary.AddDetailsFetcher(aws.AwsEbsVolumeResourceType, common.NewGenericDetailsFetcher(aws.AwsEbsVolumeResourceType, provider, deserializer))
	remoteLibrary.AddEnumerator(NewEC2EbsSnapshotEnumerator(ec2repository, factory))
	remoteLibrary.AddDetailsFetcher(aws.AwsEbsSnapshotResourceType, common.NewGenericDetailsFetcher(aws.AwsEbsSnapshotResourceType, provider, deserializer))
	remoteLibrary.AddEnumerator(NewEC2EipEnumerator(ec2repository, factory))
	remoteLibrary.AddDetailsFetcher(aws.AwsEipResourceType, common.NewGenericDetailsFetcher(aws.AwsEipResourceType, provider, deserializer))
	remoteLibrary.AddEnumerator(NewEC2AmiEnumerator(ec2repository, factory))
	remoteLibrary.AddDetailsFetcher(aws.AwsAmiResourceType, common.NewGenericDetailsFetcher(aws.AwsAmiResourceType, provider, deserializer))
	remoteLibrary.AddEnumerator(NewEC2KeyPairEnumerator(ec2repository, factory))
	remoteLibrary.AddDetailsFetcher(aws.AwsKeyPairResourceType, common.NewGenericDetailsFetcher(aws.AwsKeyPairResourceType, provider, deserializer))
	remoteLibrary.AddEnumerator(NewEC2EipAssociationEnumerator(ec2repository, factory))
	remoteLibrary.AddDetailsFetcher(aws.AwsEipAssociationResourceType, common.NewGenericDetailsFetcher(aws.AwsEipAssociationResourceType, provider, deserializer))
	remoteLibrary.AddEnumerator(NewEC2InstanceEnumerator(ec2repository, factory))
	remoteLibrary.AddDetailsFetcher(aws.AwsInstanceResourceType, NewEC2InstanceDetailsFetcher(provider, deserializer))
	remoteLibrary.AddEnumerator(NewEC2InternetGatewayEnumerator(ec2repository, factory))
	remoteLibrary.AddDetailsFetcher(aws.AwsInternetGatewayResourceType, common.NewGenericDetailsFetcher(aws.AwsInternetGatewayResourceType, provider, deserializer))

	remoteLibrary.AddEnumerator(NewKMSKeyEnumerator(kmsRepository, factory))
	remoteLibrary.AddDetailsFetcher(aws.AwsKmsKeyResourceType, common.NewGenericDetailsFetcher(aws.AwsKmsKeyResourceType, provider, deserializer))
	remoteLibrary.AddEnumerator(NewKMSAliasEnumerator(kmsRepository, factory))
	remoteLibrary.AddDetailsFetcher(aws.AwsKmsAliasResourceType, common.NewGenericDetailsFetcher(aws.AwsKmsAliasResourceType, provider, deserializer))

	remoteLibrary.AddEnumerator(NewVPCEnumerator(ec2repository, factory))
	remoteLibrary.AddDetailsFetcher(aws.AwsVpcResourceType, common.NewGenericDetailsFetcher(aws.AwsVpcResourceType, provider, deserializer))

	remoteLibrary.AddEnumerator(NewDefaultVPCEnumerator(ec2repository, factory))
	remoteLibrary.AddDetailsFetcher(aws.AwsDefaultVpcResourceType, common.NewGenericDetailsFetcher(aws.AwsDefaultVpcResourceType, provider, deserializer))

	remoteLibrary.AddEnumerator(NewRoute53HealthCheckEnumerator(route53repository, factory))
	remoteLibrary.AddDetailsFetcher(aws.AwsRoute53HealthCheckResourceType, common.NewGenericDetailsFetcher(aws.AwsRoute53HealthCheckResourceType, provider, deserializer))
	remoteLibrary.AddEnumerator(NewS3BucketPolicyEnumerator(s3Repository, factory, provider.Config))
	remoteLibrary.AddDetailsFetcher(aws.AwsS3BucketPolicyResourceType, NewS3BucketPolicyDetailsFetcher(provider, deserializer))

	supplierLibrary.AddSupplier(NewS3BucketAnalyticSupplier(provider, s3Repository, deserializer))
	supplierLibrary.AddSupplier(NewRoute53ZoneSupplier(provider, deserializer, route53repository))
	supplierLibrary.AddSupplier(NewRoute53RecordSupplier(provider, deserializer, route53repository))
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
	supplierLibrary.AddSupplier(NewSubnetSupplier(provider, deserializer, ec2repository))
	supplierLibrary.AddSupplier(NewRouteTableSupplier(provider, deserializer, ec2repository))
	supplierLibrary.AddSupplier(NewRouteSupplier(provider, deserializer, ec2repository))
	supplierLibrary.AddSupplier(NewRouteTableAssociationSupplier(provider, deserializer, ec2repository))
	supplierLibrary.AddSupplier(NewNatGatewaySupplier(provider, deserializer, ec2repository))
	supplierLibrary.AddSupplier(NewSqsQueueSupplier(provider, deserializer, sqsRepository))
	supplierLibrary.AddSupplier(NewSqsQueuePolicySupplier(provider, deserializer, sqsRepository))
	supplierLibrary.AddSupplier(NewSNSTopicSupplier(provider, deserializer, snsRepository))
	supplierLibrary.AddSupplier(NewSNSTopicPolicySupplier(provider, deserializer, snsRepository))
	supplierLibrary.AddSupplier(NewSNSTopicSubscriptionSupplier(provider, alerter, deserializer, snsRepository))
	supplierLibrary.AddSupplier(NewDynamoDBTableSupplier(provider, deserializer, dynamoDBRepository))
	supplierLibrary.AddSupplier(NewCloudfrontDistributionSupplier(provider, deserializer, cloudfrontRepository))
	supplierLibrary.AddSupplier(NewECRRepositorySupplier(provider, deserializer, ecrRepository))
	supplierLibrary.AddSupplier(NewLambdaEventSourceMappingSupplier(provider, deserializer, lambdaRepository))

	err = resourceSchemaRepository.Init(version, provider.Schema())
	if err != nil {
		return err
	}
	aws.InitResourcesMetadata(resourceSchemaRepository)

	return nil
}
