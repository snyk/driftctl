package aws

import (
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/cloudskiff/driftctl/pkg/alerter"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/terraform"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/aws/aws-sdk-go/service/route53"
)

const RemoteAWSTerraform = "aws+tf"

/**
 * Initialize remote (configure credentials, launch tf providers and start gRPC clients)
 * Required to use Scanner
 */
func Init(alerter *alerter.Alerter) error {
	provider, err := NewTerraFormProvider()
	if err != nil {
		return err
	}

	factory := AwsClientFactory{config: provider.session}

	terraform.AddProvider(terraform.AWS, provider)
	resource.AddSupplier(NewS3BucketSupplier(provider.Runner().SubRunner(), factory))
	resource.AddSupplier(NewS3BucketAnalyticSupplier(provider.Runner().SubRunner(), factory))
	resource.AddSupplier(NewS3BucketInventorySupplier(provider.Runner().SubRunner(), factory))
	resource.AddSupplier(NewS3BucketMetricSupplier(provider.Runner().SubRunner(), factory))
	resource.AddSupplier(NewS3BucketNotificationSupplier(provider.Runner().SubRunner(), factory))
	resource.AddSupplier(NewS3BucketPolicySupplier(provider.Runner().SubRunner(), factory))
	resource.AddSupplier(NewEC2EipSupplier(provider.Runner().SubRunner(), ec2.New(provider.session)))
	resource.AddSupplier(NewEC2EipAssociationSupplier(provider.Runner().SubRunner(), ec2.New(provider.session)))
	resource.AddSupplier(NewEC2EbsVolumeSupplier(provider.Runner().SubRunner(), ec2.New(provider.session)))
	resource.AddSupplier(NewEC2EbsSnapshotSupplier(provider.Runner().SubRunner(), ec2.New(provider.session)))
	resource.AddSupplier(NewRoute53ZoneSupplier(provider.Runner().SubRunner(), route53.New(provider.session)))
	resource.AddSupplier(NewRoute53RecordSupplier(provider.Runner().SubRunner(), route53.New(provider.session)))
	resource.AddSupplier(NewEC2InstanceSupplier(provider.Runner().SubRunner(), ec2.New(provider.session)))
	resource.AddSupplier(NewEC2AmiSupplier(provider.Runner().SubRunner(), ec2.New(provider.session)))
	resource.AddSupplier(NewEC2KeyPairSupplier(provider.Runner().SubRunner(), ec2.New(provider.session)))
	resource.AddSupplier(NewLambdaFunctionSupplier(provider.Runner().SubRunner(), lambda.New(provider.session)))
	resource.AddSupplier(NewDBSubnetGroupSupplier(provider.Runner().SubRunner(), rds.New(provider.session)))
	resource.AddSupplier(NewDBInstanceSupplier(provider.Runner().SubRunner(), rds.New(provider.session)))
	resource.AddSupplier(NewVPCSecurityGroupSupplier(provider.Runner(), ec2.New(provider.session)))
	resource.AddSupplier(NewIamUserSupplier(provider.Runner().SubRunner(), iam.New(provider.session)))
	resource.AddSupplier(NewIamUserPolicySupplier(provider.Runner().SubRunner(), iam.New(provider.session)))
	resource.AddSupplier(NewIamUserPolicyAttachmentSupplier(provider.Runner().SubRunner(), iam.New(provider.session)))
	resource.AddSupplier(NewIamAccessKeySupplier(provider.Runner().SubRunner(), iam.New(provider.session)))
	resource.AddSupplier(NewIamRoleSupplier(provider.Runner().SubRunner(), iam.New(provider.session)))
	resource.AddSupplier(NewIamPolicySupplier(provider.Runner().SubRunner(), iam.New(provider.session)))
	resource.AddSupplier(NewIamRolePolicySupplier(provider.Runner().SubRunner(), iam.New(provider.session)))
	resource.AddSupplier(NewIamRolePolicyAttachmentSupplier(provider.Runner().SubRunner(), iam.New(provider.session)))
	resource.AddSupplier(NewVPCSecurityGroupRuleSupplier(provider.Runner().SubRunner(), ec2.New(provider.session)))
	resource.AddSupplier(NewVPCSupplier(provider.Runner(), ec2.New(provider.session)))
	resource.AddSupplier(NewSubnetSupplier(provider.Runner(), ec2.New(provider.session)))
	resource.AddSupplier(NewRouteTableSupplier(provider.Runner(), ec2.New(provider.session)))
	resource.AddSupplier(NewRouteSupplier(provider.Runner(), ec2.New(provider.session)))
	resource.AddSupplier(NewRouteTableAssociationSupplier(provider.Runner(), ec2.New(provider.session)))
	resource.AddSupplier(NewNatGatewaySupplier(provider.Runner(), ec2.New(provider.session)))
	resource.AddSupplier(NewInternetGatewaySupplier(provider.Runner().SubRunner(), ec2.New(provider.session)))

	return nil
}
