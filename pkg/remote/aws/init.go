package aws

import (
	"github.com/cloudskiff/driftctl/pkg/alerter"
	"github.com/cloudskiff/driftctl/pkg/filter"
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
	ignore *filter.DriftIgnore) error {

	if version == "" {
		version = "3.19.0"
	}

	provider, err := NewAWSTerraformProvider(version, progress, ignore)
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

	var awsSuppliers []resource.SimpleTypeSupplier

	awsSuppliers = append(awsSuppliers, NewS3BucketSupplier(provider, s3Repository, deserializer))
	awsSuppliers = append(awsSuppliers, NewS3BucketAnalyticSupplier(provider, s3Repository, deserializer))
	awsSuppliers = append(awsSuppliers, NewS3BucketInventorySupplier(provider, s3Repository, deserializer))
	awsSuppliers = append(awsSuppliers, NewS3BucketMetricSupplier(provider, s3Repository, deserializer))
	awsSuppliers = append(awsSuppliers, NewS3BucketNotificationSupplier(provider, s3Repository, deserializer))
	awsSuppliers = append(awsSuppliers, NewS3BucketPolicySupplier(provider, s3Repository, deserializer))
	awsSuppliers = append(awsSuppliers, NewEC2EipSupplier(provider, ec2repository, deserializer))
	awsSuppliers = append(awsSuppliers, NewEC2EipAssociationSupplier(provider, deserializer, ec2repository))
	awsSuppliers = append(awsSuppliers, NewEC2EbsVolumeSupplier(provider, deserializer, ec2repository))
	awsSuppliers = append(awsSuppliers, NewEC2EbsSnapshotSupplier(provider, deserializer, ec2repository))
	awsSuppliers = append(awsSuppliers, NewRoute53ZoneSupplier(provider, deserializer, route53repository))
	awsSuppliers = append(awsSuppliers, NewRoute53RecordSupplier(provider, deserializer, route53repository))
	awsSuppliers = append(awsSuppliers, NewEC2InstanceSupplier(provider, deserializer, ec2repository))
	awsSuppliers = append(awsSuppliers, NewEC2AmiSupplier(provider, deserializer, ec2repository))
	awsSuppliers = append(awsSuppliers, NewEC2KeyPairSupplier(provider, deserializer, ec2repository))
	awsSuppliers = append(awsSuppliers, NewLambdaFunctionSupplier(provider, deserializer, lambdaRepository))
	awsSuppliers = append(awsSuppliers, NewDBSubnetGroupSupplier(provider, deserializer, rdsRepository))
	awsSuppliers = append(awsSuppliers, NewDBInstanceSupplier(provider, deserializer, rdsRepository))
	awsSuppliers = append(awsSuppliers, NewVPCSecurityGroupSupplier(provider, deserializer, ec2repository))
	awsSuppliers = append(awsSuppliers, NewIamUserSupplier(provider, deserializer, iamRepository))
	awsSuppliers = append(awsSuppliers, NewIamUserPolicySupplier(provider, deserializer, iamRepository))
	awsSuppliers = append(awsSuppliers, NewIamUserPolicyAttachmentSupplier(provider, deserializer, iamRepository))
	awsSuppliers = append(awsSuppliers, NewIamAccessKeySupplier(provider, deserializer, iamRepository))
	awsSuppliers = append(awsSuppliers, NewIamRoleSupplier(provider, deserializer, iamRepository))
	awsSuppliers = append(awsSuppliers, NewIamPolicySupplier(provider, deserializer, iamRepository))
	awsSuppliers = append(awsSuppliers, NewIamRolePolicySupplier(provider, deserializer, iamRepository))
	awsSuppliers = append(awsSuppliers, NewIamRolePolicyAttachmentSupplier(provider, deserializer, iamRepository))
	awsSuppliers = append(awsSuppliers, NewVPCSecurityGroupRuleSupplier(provider, deserializer, ec2repository))
	awsSuppliers = append(awsSuppliers, NewVPCSupplier(provider, deserializer, ec2repository))
	awsSuppliers = append(awsSuppliers, NewSubnetSupplier(provider, deserializer, ec2repository))
	awsSuppliers = append(awsSuppliers, NewRouteTableSupplier(provider, deserializer, ec2repository))
	awsSuppliers = append(awsSuppliers, NewRouteSupplier(provider, deserializer, ec2repository))
	awsSuppliers = append(awsSuppliers, NewRouteTableAssociationSupplier(provider, deserializer, ec2repository))
	awsSuppliers = append(awsSuppliers, NewNatGatewaySupplier(provider, deserializer, ec2repository))
	awsSuppliers = append(awsSuppliers, NewInternetGatewaySupplier(provider, deserializer, ec2repository))
	awsSuppliers = append(awsSuppliers, NewSqsQueueSupplier(provider, deserializer, sqsRepository))
	awsSuppliers = append(awsSuppliers, NewSqsQueuePolicySupplier(provider, deserializer, sqsRepository))
	awsSuppliers = append(awsSuppliers, NewSNSTopicSupplier(provider, deserializer, snsRepository))
	awsSuppliers = append(awsSuppliers, NewSNSTopicPolicySupplier(provider, deserializer, snsRepository))
	awsSuppliers = append(awsSuppliers, NewSNSTopicSubscriptionSupplier(provider, alerter, deserializer, snsRepository))
	awsSuppliers = append(awsSuppliers, NewDynamoDBTableSupplier(provider, deserializer, dynamoDBRepository))
	awsSuppliers = append(awsSuppliers, NewRoute53HealthCheckSupplier(provider, deserializer, route53repository))
	awsSuppliers = append(awsSuppliers, NewCloudfrontDistributionSupplier(provider, deserializer, cloudfrontRepository))
	awsSuppliers = append(awsSuppliers, NewECRRepositorySupplier(provider, deserializer, ecrRepository))
	awsSuppliers = append(awsSuppliers, NewKMSKeySupplier(provider, deserializer, kmsRepository))
	awsSuppliers = append(awsSuppliers, NewKMSAliasSupplier(provider, deserializer, kmsRepository))
	awsSuppliers = append(awsSuppliers, NewLambdaEventSourceMappingSupplier(provider, deserializer, lambdaRepository))

	for _, supplier := range awsSuppliers {
		if ignore.IsTypeIgnored(supplier.SuppliedType().String()) {
			continue
		}
		supplierLibrary.AddSupplier(supplier)
	}

	err = resourceSchemaRepository.Init(version, provider.Schema())
	if err != nil {
		return err
	}
	aws.InitResourcesMetadata(resourceSchemaRepository)

	return nil
}
