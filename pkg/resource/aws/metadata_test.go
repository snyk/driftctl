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
		aws.AwsAmiResourceType:                             {resource.FlagDeepMode},
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
		aws.AwsAppAutoscalingPolicyResourceType:            {resource.FlagDeepMode},
		aws.AwsAppAutoscalingScheduledActionResourceType:   {},
		aws.AwsAppAutoscalingTargetResourceType:            {resource.FlagDeepMode},
		aws.AwsCloudformationStackResourceType:             {resource.FlagDeepMode},
		aws.AwsCloudfrontDistributionResourceType:          {resource.FlagDeepMode},
		aws.AwsDbInstanceResourceType:                      {resource.FlagDeepMode},
		aws.AwsDbSubnetGroupResourceType:                   {resource.FlagDeepMode},
		aws.AwsDefaultNetworkACLResourceType:               {resource.FlagDeepMode},
		aws.AwsDefaultRouteTableResourceType:               {resource.FlagDeepMode},
		aws.AwsDefaultSecurityGroupResourceType:            {resource.FlagDeepMode},
		aws.AwsDefaultSubnetResourceType:                   {resource.FlagDeepMode},
		aws.AwsDefaultVpcResourceType:                      {resource.FlagDeepMode},
		aws.AwsDynamodbTableResourceType:                   {resource.FlagDeepMode},
		aws.AwsEbsEncryptionByDefaultResourceType:          {resource.FlagDeepMode},
		aws.AwsEbsSnapshotResourceType:                     {resource.FlagDeepMode},
		aws.AwsEbsVolumeResourceType:                       {resource.FlagDeepMode},
		aws.AwsEcrRepositoryResourceType:                   {resource.FlagDeepMode},
		aws.AwsEipResourceType:                             {resource.FlagDeepMode},
		aws.AwsEipAssociationResourceType:                  {resource.FlagDeepMode},
		aws.AwsElastiCacheClusterResourceType:              {},
		aws.AwsIamAccessKeyResourceType:                    {resource.FlagDeepMode},
		aws.AwsIamPolicyResourceType:                       {resource.FlagDeepMode},
		aws.AwsIamPolicyAttachmentResourceType:             {resource.FlagDeepMode},
		aws.AwsIamRoleResourceType:                         {resource.FlagDeepMode},
		aws.AwsIamRolePolicyResourceType:                   {resource.FlagDeepMode},
		aws.AwsIamRolePolicyAttachmentResourceType:         {resource.FlagDeepMode},
		aws.AwsIamUserResourceType:                         {resource.FlagDeepMode},
		aws.AwsIamUserPolicyResourceType:                   {resource.FlagDeepMode},
		aws.AwsIamUserPolicyAttachmentResourceType:         {resource.FlagDeepMode},
		aws.AwsIamGroupPolicyResourceType:                  {},
		aws.AwsIamGroupPolicyAttachmentResourceType:        {},
		aws.AwsInstanceResourceType:                        {resource.FlagDeepMode},
		aws.AwsInternetGatewayResourceType:                 {resource.FlagDeepMode},
		aws.AwsKeyPairResourceType:                         {resource.FlagDeepMode},
		aws.AwsKmsAliasResourceType:                        {resource.FlagDeepMode},
		aws.AwsKmsKeyResourceType:                          {resource.FlagDeepMode},
		aws.AwsLambdaEventSourceMappingResourceType:        {resource.FlagDeepMode},
		aws.AwsLambdaFunctionResourceType:                  {resource.FlagDeepMode},
		aws.AwsNatGatewayResourceType:                      {resource.FlagDeepMode},
		aws.AwsNetworkACLResourceType:                      {resource.FlagDeepMode},
		aws.AwsRDSClusterResourceType:                      {resource.FlagDeepMode},
		aws.AwsRDSClusterInstanceResourceType:              {},
		aws.AwsRouteResourceType:                           {resource.FlagDeepMode},
		aws.AwsRoute53HealthCheckResourceType:              {resource.FlagDeepMode},
		aws.AwsRoute53RecordResourceType:                   {resource.FlagDeepMode},
		aws.AwsRoute53ZoneResourceType:                     {resource.FlagDeepMode},
		aws.AwsRouteTableResourceType:                      {resource.FlagDeepMode},
		aws.AwsRouteTableAssociationResourceType:           {resource.FlagDeepMode},
		aws.AwsS3BucketResourceType:                        {resource.FlagDeepMode},
		aws.AwsS3BucketAnalyticsConfigurationResourceType:  {resource.FlagDeepMode},
		aws.AwsS3BucketInventoryResourceType:               {resource.FlagDeepMode},
		aws.AwsS3BucketMetricResourceType:                  {resource.FlagDeepMode},
		aws.AwsS3BucketNotificationResourceType:            {resource.FlagDeepMode},
		aws.AwsS3BucketPolicyResourceType:                  {resource.FlagDeepMode},
		aws.AwsS3BucketPublicAccessBlockResourceType:       {},
		aws.AwsSecurityGroupResourceType:                   {resource.FlagDeepMode},
		aws.AwsSnsTopicResourceType:                        {resource.FlagDeepMode},
		aws.AwsSnsTopicPolicyResourceType:                  {resource.FlagDeepMode},
		aws.AwsSnsTopicSubscriptionResourceType:            {resource.FlagDeepMode},
		aws.AwsSqsQueueResourceType:                        {resource.FlagDeepMode},
		aws.AwsSqsQueuePolicyResourceType:                  {resource.FlagDeepMode},
		aws.AwsSubnetResourceType:                          {resource.FlagDeepMode},
		aws.AwsVpcResourceType:                             {resource.FlagDeepMode},
		aws.AwsSecurityGroupRuleResourceType:               {resource.FlagDeepMode},
		aws.AwsNetworkACLRuleResourceType:                  {resource.FlagDeepMode},
		aws.AwsLaunchTemplateResourceType:                  {resource.FlagDeepMode},
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
