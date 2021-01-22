package aws

import (
	"github.com/cloudskiff/driftctl/pkg/alerter"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/terraform"
)

const RemoteAWSTerraform = "aws+tf"

/**
 * Initialize remote (configure credentials, launch tf providers and start gRPC clients)
 * Required to use Scanner
 */
func Init(alerter *alerter.Alerter, providerLibrary *terraform.ProviderLibrary, supplierLibrary *resource.SupplierLibrary) error {
	provider, err := NewTerraFormProvider()
	if err != nil {
		return err
	}

	factory := AwsClientFactory{config: provider.session}

	providerLibrary.AddProvider(terraform.AWS, provider)

	supplierLibrary.AddSupplier(NewS3BucketSupplier(provider, factory))
	supplierLibrary.AddSupplier(NewS3BucketAnalyticSupplier(provider, factory))
	supplierLibrary.AddSupplier(NewS3BucketInventorySupplier(provider, factory))
	supplierLibrary.AddSupplier(NewS3BucketMetricSupplier(provider, factory))
	supplierLibrary.AddSupplier(NewS3BucketNotificationSupplier(provider, factory))
	supplierLibrary.AddSupplier(NewS3BucketPolicySupplier(provider, factory))
	supplierLibrary.AddSupplier(NewEC2EipSupplier(provider))
	supplierLibrary.AddSupplier(NewEC2EipAssociationSupplier(provider))
	supplierLibrary.AddSupplier(NewEC2EbsVolumeSupplier(provider))
	supplierLibrary.AddSupplier(NewEC2EbsSnapshotSupplier(provider))
	supplierLibrary.AddSupplier(NewRoute53ZoneSupplier(provider))
	supplierLibrary.AddSupplier(NewRoute53RecordSupplier(provider))
	supplierLibrary.AddSupplier(NewEC2InstanceSupplier(provider))
	supplierLibrary.AddSupplier(NewEC2AmiSupplier(provider))
	supplierLibrary.AddSupplier(NewEC2KeyPairSupplier(provider))
	supplierLibrary.AddSupplier(NewLambdaFunctionSupplier(provider))
	supplierLibrary.AddSupplier(NewDBSubnetGroupSupplier(provider))
	supplierLibrary.AddSupplier(NewDBInstanceSupplier(provider))
	supplierLibrary.AddSupplier(NewVPCSecurityGroupSupplier(provider))
	supplierLibrary.AddSupplier(NewIamUserSupplier(provider))
	supplierLibrary.AddSupplier(NewIamUserPolicySupplier(provider))
	supplierLibrary.AddSupplier(NewIamUserPolicyAttachmentSupplier(provider))
	supplierLibrary.AddSupplier(NewIamAccessKeySupplier(provider))
	supplierLibrary.AddSupplier(NewIamRoleSupplier(provider))
	supplierLibrary.AddSupplier(NewIamPolicySupplier(provider))
	supplierLibrary.AddSupplier(NewIamRolePolicySupplier(provider))
	supplierLibrary.AddSupplier(NewIamRolePolicyAttachmentSupplier(provider))
	supplierLibrary.AddSupplier(NewVPCSecurityGroupRuleSupplier(provider))
	supplierLibrary.AddSupplier(NewVPCSupplier(provider))
	supplierLibrary.AddSupplier(NewSubnetSupplier(provider))
	supplierLibrary.AddSupplier(NewRouteTableSupplier(provider))
	supplierLibrary.AddSupplier(NewRouteSupplier(provider))
	supplierLibrary.AddSupplier(NewRouteTableAssociationSupplier(provider))
	supplierLibrary.AddSupplier(NewNatGatewaySupplier(provider))
	supplierLibrary.AddSupplier(NewInternetGatewaySupplier(provider))

	return nil
}
