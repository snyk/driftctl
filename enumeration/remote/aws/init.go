package aws

import (
	"github.com/snyk/driftctl/enumeration"
	"github.com/snyk/driftctl/enumeration/alerter"
	client "github.com/snyk/driftctl/enumeration/remote/aws/client"
	"github.com/snyk/driftctl/enumeration/remote/aws/repository"
	"github.com/snyk/driftctl/enumeration/remote/cache"
	"github.com/snyk/driftctl/enumeration/remote/common"
	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/enumeration/terraform"
)

/**
 * Initialize remote (configure credentials, launch tf providers and start gRPC clients)
 * Required to use Scanner
 */

func Init(version string, alerter alerter.AlerterInterface, providerLibrary *terraform.ProviderLibrary, remoteLibrary *common.RemoteLibrary, progress enumeration.ProgressCounter, factory resource.ResourceFactory, configDir string) error {

	provider, err := NewAWSTerraformProvider(version, progress, configDir)
	if err != nil {
		return err
	}
	err = provider.CheckCredentialsExist()
	if err != nil {
		return err
	}
	err = provider.Init()
	if err != nil {
		return err
	}

	repositoryCache := cache.New(100)

	s3Repository := repository.NewS3Repository(client.NewAWSClientFactory(provider.session), repositoryCache)
	s3ControlRepository := repository.NewS3ControlRepository(client.NewAWSClientFactory(provider.session), repositoryCache)
	ec2repository := repository.NewEC2Repository(provider.session, repositoryCache)
	elbv2Repository := repository.NewELBV2Repository(provider.session, repositoryCache)
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
	cloudformationRepository := repository.NewCloudformationRepository(provider.session, repositoryCache)
	cloudtrailRepository := repository.NewCloudtrailRepository(provider.session, repositoryCache)
	apigatewayRepository := repository.NewApiGatewayRepository(provider.session, repositoryCache)
	appAutoScalingRepository := repository.NewAppAutoScalingRepository(provider.session, repositoryCache)
	apigatewayv2Repository := repository.NewApiGatewayV2Repository(provider.session, repositoryCache)
	autoscalingRepository := repository.NewAutoScalingRepository(provider.session, repositoryCache)
	elbRepository := repository.NewELBRepository(provider.session, repositoryCache)
	elasticacheRepository := repository.NewElastiCacheRepository(provider.session, repositoryCache)

	providerLibrary.AddProvider(terraform.AWS, provider)

	remoteLibrary.AddEnumerator(NewS3BucketEnumerator(s3Repository, factory, provider.Config, alerter))
	remoteLibrary.AddEnumerator(NewS3BucketInventoryEnumerator(s3Repository, factory, provider.Config, alerter))
	remoteLibrary.AddEnumerator(NewS3BucketNotificationEnumerator(s3Repository, factory, provider.Config, alerter))
	remoteLibrary.AddEnumerator(NewS3BucketMetricsEnumerator(s3Repository, factory, provider.Config, alerter))
	remoteLibrary.AddEnumerator(NewS3BucketPolicyEnumerator(s3Repository, factory, provider.Config, alerter))
	remoteLibrary.AddEnumerator(NewS3BucketAnalyticEnumerator(s3Repository, factory, provider.Config, alerter))
	remoteLibrary.AddEnumerator(NewS3BucketPublicAccessBlockEnumerator(s3Repository, factory, provider.Config, alerter))
	remoteLibrary.AddEnumerator(NewS3AccountPublicAccessBlockEnumerator(s3ControlRepository, factory, provider.accountId, alerter))

	remoteLibrary.AddEnumerator(NewEC2EbsVolumeEnumerator(ec2repository, factory))
	remoteLibrary.AddEnumerator(NewEC2EbsSnapshotEnumerator(ec2repository, factory))
	remoteLibrary.AddEnumerator(NewEC2EipEnumerator(ec2repository, factory))
	remoteLibrary.AddEnumerator(NewEC2AmiEnumerator(ec2repository, factory))
	remoteLibrary.AddEnumerator(NewEC2KeyPairEnumerator(ec2repository, factory))
	remoteLibrary.AddEnumerator(NewEC2EipAssociationEnumerator(ec2repository, factory))
	remoteLibrary.AddEnumerator(NewEC2InstanceEnumerator(ec2repository, factory))
	remoteLibrary.AddEnumerator(NewEC2InternetGatewayEnumerator(ec2repository, factory))
	remoteLibrary.AddEnumerator(NewVPCEnumerator(ec2repository, factory))
	remoteLibrary.AddEnumerator(NewDefaultVPCEnumerator(ec2repository, factory))
	remoteLibrary.AddEnumerator(NewEC2RouteTableEnumerator(ec2repository, factory))
	remoteLibrary.AddEnumerator(NewEC2DefaultRouteTableEnumerator(ec2repository, factory))
	remoteLibrary.AddEnumerator(NewEC2RouteTableAssociationEnumerator(ec2repository, factory))
	remoteLibrary.AddEnumerator(NewEC2SubnetEnumerator(ec2repository, factory))
	remoteLibrary.AddEnumerator(NewEC2DefaultSubnetEnumerator(ec2repository, factory))
	remoteLibrary.AddEnumerator(NewVPCSecurityGroupEnumerator(ec2repository, factory))
	remoteLibrary.AddEnumerator(NewVPCDefaultSecurityGroupEnumerator(ec2repository, factory))
	remoteLibrary.AddEnumerator(NewEC2NatGatewayEnumerator(ec2repository, factory))
	remoteLibrary.AddEnumerator(NewEC2NetworkACLEnumerator(ec2repository, factory))
	remoteLibrary.AddEnumerator(NewEC2NetworkACLRuleEnumerator(ec2repository, factory))
	remoteLibrary.AddEnumerator(NewEC2DefaultNetworkACLEnumerator(ec2repository, factory))
	remoteLibrary.AddEnumerator(NewEC2RouteEnumerator(ec2repository, factory))
	remoteLibrary.AddEnumerator(NewVPCSecurityGroupRuleEnumerator(ec2repository, factory))
	remoteLibrary.AddEnumerator(NewLaunchTemplateEnumerator(ec2repository, factory))
	remoteLibrary.AddEnumerator(NewEC2EbsEncryptionByDefaultEnumerator(ec2repository, factory))

	remoteLibrary.AddEnumerator(NewKMSKeyEnumerator(kmsRepository, factory))
	remoteLibrary.AddEnumerator(NewKMSAliasEnumerator(kmsRepository, factory))

	remoteLibrary.AddEnumerator(NewRoute53HealthCheckEnumerator(route53repository, factory))
	remoteLibrary.AddEnumerator(NewRoute53ZoneEnumerator(route53repository, factory))
	remoteLibrary.AddEnumerator(NewRoute53RecordEnumerator(route53repository, factory))

	remoteLibrary.AddEnumerator(NewCloudfrontDistributionEnumerator(cloudfrontRepository, factory))

	remoteLibrary.AddEnumerator(NewRDSDBInstanceEnumerator(rdsRepository, factory))
	remoteLibrary.AddEnumerator(NewRDSDBSubnetGroupEnumerator(rdsRepository, factory))

	remoteLibrary.AddEnumerator(NewSQSQueueEnumerator(sqsRepository, factory))
	remoteLibrary.AddEnumerator(NewSQSQueuePolicyEnumerator(sqsRepository, factory))

	remoteLibrary.AddEnumerator(NewSNSTopicEnumerator(snsRepository, factory))
	remoteLibrary.AddEnumerator(NewSNSTopicPolicyEnumerator(snsRepository, factory))
	remoteLibrary.AddEnumerator(NewSNSTopicSubscriptionEnumerator(snsRepository, factory, alerter))

	remoteLibrary.AddEnumerator(NewDynamoDBTableEnumerator(dynamoDBRepository, factory))

	remoteLibrary.AddEnumerator(NewIamPolicyEnumerator(iamRepository, factory))

	remoteLibrary.AddEnumerator(NewLambdaFunctionEnumerator(lambdaRepository, factory))
	remoteLibrary.AddEnumerator(NewLambdaEventSourceMappingEnumerator(lambdaRepository, factory))

	remoteLibrary.AddEnumerator(NewIamUserEnumerator(iamRepository, factory))
	remoteLibrary.AddEnumerator(NewIamUserPolicyEnumerator(iamRepository, factory))
	remoteLibrary.AddEnumerator(NewIamRoleEnumerator(iamRepository, factory))
	remoteLibrary.AddEnumerator(NewIamAccessKeyEnumerator(iamRepository, factory))
	remoteLibrary.AddEnumerator(NewIamRolePolicyAttachmentEnumerator(iamRepository, factory))
	remoteLibrary.AddEnumerator(NewIamRolePolicyEnumerator(iamRepository, factory))
	remoteLibrary.AddEnumerator(NewIamUserPolicyAttachmentEnumerator(iamRepository, factory))
	remoteLibrary.AddEnumerator(NewIamGroupPolicyEnumerator(iamRepository, factory))
	remoteLibrary.AddEnumerator(NewIamGroupEnumerator(iamRepository, factory))
	remoteLibrary.AddEnumerator(NewIamGroupPolicyAttachmentEnumerator(iamRepository, factory))

	remoteLibrary.AddEnumerator(NewECRRepositoryEnumerator(ecrRepository, factory))
	remoteLibrary.AddEnumerator(NewECRRepositoryPolicyEnumerator(ecrRepository, factory))

	remoteLibrary.AddEnumerator(NewRDSClusterEnumerator(rdsRepository, factory))

	remoteLibrary.AddEnumerator(NewCloudformationStackEnumerator(cloudformationRepository, factory))

	remoteLibrary.AddEnumerator(NewCloudtrailEnumerator(cloudtrailRepository, factory))

	remoteLibrary.AddEnumerator(NewApiGatewayRestApiEnumerator(apigatewayRepository, factory))
	remoteLibrary.AddEnumerator(NewApiGatewayAccountEnumerator(apigatewayRepository, factory))
	remoteLibrary.AddEnumerator(NewApiGatewayApiKeyEnumerator(apigatewayRepository, factory))
	remoteLibrary.AddEnumerator(NewApiGatewayAuthorizerEnumerator(apigatewayRepository, factory))
	remoteLibrary.AddEnumerator(NewApiGatewayStageEnumerator(apigatewayRepository, factory))
	remoteLibrary.AddEnumerator(NewApiGatewayResourceEnumerator(apigatewayRepository, factory))
	remoteLibrary.AddEnumerator(NewApiGatewayDomainNameEnumerator(apigatewayRepository, factory))
	remoteLibrary.AddEnumerator(NewApiGatewayVpcLinkEnumerator(apigatewayRepository, factory))
	remoteLibrary.AddEnumerator(NewApiGatewayRequestValidatorEnumerator(apigatewayRepository, factory))
	remoteLibrary.AddEnumerator(NewApiGatewayRestApiPolicyEnumerator(apigatewayRepository, factory))
	remoteLibrary.AddEnumerator(NewApiGatewayBasePathMappingEnumerator(apigatewayRepository, factory))
	remoteLibrary.AddEnumerator(NewApiGatewayMethodEnumerator(apigatewayRepository, factory))
	remoteLibrary.AddEnumerator(NewApiGatewayModelEnumerator(apigatewayRepository, factory))
	remoteLibrary.AddEnumerator(NewApiGatewayMethodResponseEnumerator(apigatewayRepository, factory))
	remoteLibrary.AddEnumerator(NewApiGatewayGatewayResponseEnumerator(apigatewayRepository, factory))
	remoteLibrary.AddEnumerator(NewApiGatewayMethodSettingsEnumerator(apigatewayRepository, factory))
	remoteLibrary.AddEnumerator(NewApiGatewayIntegrationEnumerator(apigatewayRepository, factory))
	remoteLibrary.AddEnumerator(NewApiGatewayIntegrationResponseEnumerator(apigatewayRepository, factory))

	remoteLibrary.AddEnumerator(NewApiGatewayV2ApiEnumerator(apigatewayv2Repository, factory))
	remoteLibrary.AddEnumerator(NewApiGatewayV2RouteEnumerator(apigatewayv2Repository, factory))
	remoteLibrary.AddEnumerator(NewApiGatewayV2DeploymentEnumerator(apigatewayv2Repository, factory))
	remoteLibrary.AddEnumerator(NewApiGatewayV2VpcLinkEnumerator(apigatewayv2Repository, factory))
	remoteLibrary.AddEnumerator(NewApiGatewayV2AuthorizerEnumerator(apigatewayv2Repository, factory))
	remoteLibrary.AddEnumerator(NewApiGatewayV2IntegrationEnumerator(apigatewayv2Repository, factory))
	remoteLibrary.AddEnumerator(NewApiGatewayV2ModelEnumerator(apigatewayv2Repository, factory))
	remoteLibrary.AddEnumerator(NewApiGatewayV2StageEnumerator(apigatewayv2Repository, factory))
	remoteLibrary.AddEnumerator(NewApiGatewayV2RouteResponseEnumerator(apigatewayv2Repository, factory))
	remoteLibrary.AddEnumerator(NewApiGatewayV2MappingEnumerator(apigatewayv2Repository, apigatewayRepository, factory))
	remoteLibrary.AddEnumerator(NewApiGatewayV2DomainNameEnumerator(apigatewayRepository, factory))
	remoteLibrary.AddEnumerator(NewApiGatewayV2IntegrationResponseEnumerator(apigatewayv2Repository, factory))

	remoteLibrary.AddEnumerator(NewAppAutoscalingTargetEnumerator(appAutoScalingRepository, factory))

	remoteLibrary.AddEnumerator(NewAppAutoscalingPolicyEnumerator(appAutoScalingRepository, factory))

	remoteLibrary.AddEnumerator(NewAppAutoscalingScheduledActionEnumerator(appAutoScalingRepository, factory))

	remoteLibrary.AddEnumerator(NewLaunchConfigurationEnumerator(autoscalingRepository, factory))

	remoteLibrary.AddEnumerator(NewLoadBalancerEnumerator(elbv2Repository, factory))
	remoteLibrary.AddEnumerator(NewLoadBalancerListenerEnumerator(elbv2Repository, factory))

	remoteLibrary.AddEnumerator(NewClassicLoadBalancerEnumerator(elbRepository, factory))

	remoteLibrary.AddEnumerator(NewElastiCacheClusterEnumerator(elasticacheRepository, factory))

	return nil
}
