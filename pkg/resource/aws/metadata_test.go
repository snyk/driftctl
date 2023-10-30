package aws_test

import (
	"testing"

	"github.com/snyk/driftctl/enumeration/resource"
	"github.com/snyk/driftctl/pkg/resource/aws"
	testresource "github.com/snyk/driftctl/test/resource"
	"github.com/stretchr/testify/assert"
)

func TestAWS_Metadata_Flags(t *testing.T) {
	testcases := map[string][]resource.Flags{
		aws.AwsAmiResourceType:                             {},
		aws.AwsApiGatewayAccountResourceType:               {},
		aws.AwsApiGatewayApiKeyResourceType:                {},
		aws.AwsApiGatewayAuthorizerResourceType:            {},
		aws.AwsApiGatewayBasePathMappingResourceType:       {},
		aws.AwsApiGatewayDeploymentResourceType:            {},
		aws.AwsApiGatewayDomainNameResourceType:            {},
		aws.AwsApiGatewayGatewayResponseResourceType:       {},
		aws.AwsApiGatewayIntegrationResourceType:           {},
		aws.AwsApiGatewayIntegrationResponseResourceType:   {},
		aws.AwsApiGatewayMethodResourceType:                {},
		aws.AwsApiGatewayMethodResponseResourceType:        {},
		aws.AwsApiGatewayMethodSettingsResourceType:        {},
		aws.AwsApiGatewayModelResourceType:                 {},
		aws.AwsApiGatewayRequestValidatorResourceType:      {},
		aws.AwsApiGatewayResourceResourceType:              {},
		aws.AwsApiGatewayRestApiResourceType:               {},
		aws.AwsApiGatewayRestApiPolicyResourceType:         {},
		aws.AwsApiGatewayStageResourceType:                 {},
		aws.AwsApiGatewayVpcLinkResourceType:               {},
		aws.AwsApiGatewayV2ApiResourceType:                 {},
		aws.AwsApiGatewayV2RouteResourceType:               {},
		aws.AwsApiGatewayV2DeploymentResourceType:          {},
		aws.AwsApiGatewayV2VpcLinkResourceType:             {},
		aws.AwsApiGatewayV2AuthorizerResourceType:          {},
		aws.AwsApiGatewayV2RouteResponseResourceType:       {},
		aws.AwsApiGatewayV2DomainNameResourceType:          {},
		aws.AwsApiGatewayV2ModelResourceType:               {},
		aws.AwsApiGatewayV2StageResourceType:               {},
		aws.AwsApiGatewayV2MappingResourceType:             {},
		aws.AwsApiGatewayV2IntegrationResourceType:         {},
		aws.AwsApiGatewayV2IntegrationResponseResourceType: {},
		aws.AwsAppAutoscalingPolicyResourceType:            {},
		aws.AwsAppAutoscalingScheduledActionResourceType:   {},
		aws.AwsAppAutoscalingTargetResourceType:            {},
		aws.AwsCloudformationStackResourceType:             {},
		aws.AwsCloudfrontDistributionResourceType:          {},
		aws.AwsDbInstanceResourceType:                      {},
		aws.AwsDbSubnetGroupResourceType:                   {},
		aws.AwsDefaultNetworkACLResourceType:               {},
		aws.AwsDefaultRouteTableResourceType:               {},
		aws.AwsDefaultSecurityGroupResourceType:            {},
		aws.AwsDefaultSubnetResourceType:                   {},
		aws.AwsDefaultVpcResourceType:                      {},
		aws.AwsDynamodbTableResourceType:                   {},
		aws.AwsEbsEncryptionByDefaultResourceType:          {},
		aws.AwsEbsSnapshotResourceType:                     {},
		aws.AwsEbsVolumeResourceType:                       {},
		aws.AwsEcrRepositoryResourceType:                   {},
		aws.AwsEipResourceType:                             {},
		aws.AwsEipAssociationResourceType:                  {},
		aws.AwsElastiCacheClusterResourceType:              {},
		aws.AwsIamAccessKeyResourceType:                    {},
		aws.AwsIamPolicyResourceType:                       {},
		aws.AwsIamPolicyAttachmentResourceType:             {},
		aws.AwsIamRoleResourceType:                         {},
		aws.AwsIamRolePolicyResourceType:                   {},
		aws.AwsIamRolePolicyAttachmentResourceType:         {},
		aws.AwsIamUserResourceType:                         {},
		aws.AwsIamUserPolicyResourceType:                   {},
		aws.AwsIamUserPolicyAttachmentResourceType:         {},
		aws.AwsIamGroupPolicyResourceType:                  {},
		aws.AwsIamGroupPolicyAttachmentResourceType:        {},
		aws.AwsInstanceResourceType:                        {},
		aws.AwsInternetGatewayResourceType:                 {},
		aws.AwsKeyPairResourceType:                         {},
		aws.AwsKmsAliasResourceType:                        {},
		aws.AwsKmsKeyResourceType:                          {},
		aws.AwsLambdaEventSourceMappingResourceType:        {},
		aws.AwsLambdaFunctionResourceType:                  {},
		aws.AwsNatGatewayResourceType:                      {},
		aws.AwsNetworkACLResourceType:                      {},
		aws.AwsRDSClusterResourceType:                      {},
		aws.AwsRDSClusterInstanceResourceType:              {},
		aws.AwsRouteResourceType:                           {},
		aws.AwsRoute53HealthCheckResourceType:              {},
		aws.AwsRoute53RecordResourceType:                   {},
		aws.AwsRoute53ZoneResourceType:                     {},
		aws.AwsRouteTableResourceType:                      {},
		aws.AwsRouteTableAssociationResourceType:           {},
		aws.AwsS3BucketResourceType:                        {},
		aws.AwsS3BucketAnalyticsConfigurationResourceType:  {},
		aws.AwsS3BucketInventoryResourceType:               {},
		aws.AwsS3BucketMetricResourceType:                  {},
		aws.AwsS3BucketNotificationResourceType:            {},
		aws.AwsS3BucketPolicyResourceType:                  {},
		aws.AwsS3BucketPublicAccessBlockResourceType:       {},
		aws.AwsS3AccountPublicAccessBlockResourceType:      {},
		aws.AwsSecurityGroupResourceType:                   {},
		aws.AwsSnsTopicResourceType:                        {},
		aws.AwsSnsTopicPolicyResourceType:                  {},
		aws.AwsSnsTopicSubscriptionResourceType:            {},
		aws.AwsSqsQueueResourceType:                        {},
		aws.AwsSqsQueuePolicyResourceType:                  {},
		aws.AwsSubnetResourceType:                          {},
		aws.AwsVpcResourceType:                             {},
		aws.AwsSecurityGroupRuleResourceType:               {},
		aws.AwsNetworkACLRuleResourceType:                  {},
		aws.AwsLaunchTemplateResourceType:                  {},
		aws.AwsLaunchConfigurationResourceType:             {},
		aws.AwsLoadBalancerResourceType:                    {},
		aws.AwsApplicationLoadBalancerResourceType:         {},
		aws.AwsClassicLoadBalancerResourceType:             {},
		aws.AwsLoadBalancerListenerResourceType:            {},
		aws.AwsApplicationLoadBalancerListenerResourceType: {},
		aws.AwsIamGroupResourceType:                        {},
		aws.AwsEcrRepositoryPolicyResourceType:             {},
	}

	schemaRepository := testresource.InitFakeSchemaRepository("aws", "3.19.0")
	aws.InitResourcesMetadata(schemaRepository)

	for ty, flags := range testcases {
		t.Run(ty, func(tt *testing.T) {
			sch, exist := schemaRepository.GetSchema(ty)
			assert.True(tt, exist)

			if len(flags) == 0 {
				assert.Equal(tt, resource.Flags(0x0), sch.Flags, "should not have any flag")
				return
			}

			for _, flag := range flags {
				assert.Truef(tt, sch.Flags.HasFlag(flag), "should have given flag %d", flag)
			}
		})
	}
}
