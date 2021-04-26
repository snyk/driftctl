package aws

import "github.com/cloudskiff/driftctl/pkg/resource"

func InitResourcesMetadata(resourceSchemaRepository resource.SchemaRepositoryInterface) {
	initAwsAmiMetaData(resourceSchemaRepository)
	initAwsCloudfrontDistributionMetaData(resourceSchemaRepository)
	initAwsDbInstanceMetaData(resourceSchemaRepository)
	initAwsDbSubnetGroupMetaData(resourceSchemaRepository)
}
