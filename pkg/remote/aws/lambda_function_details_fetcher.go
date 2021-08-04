package aws

import (
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/sirupsen/logrus"
)

type LambdaFunctionDetailsFetcher struct {
	reader       terraform.ResourceReader
	deserializer *resource.Deserializer
}

func NewLambdaFunctionDetailsFetcher(provider terraform.ResourceReader, deserializer *resource.Deserializer) *LambdaFunctionDetailsFetcher {
	return &LambdaFunctionDetailsFetcher{
		reader:       provider,
		deserializer: deserializer,
	}
}

func (r *LambdaFunctionDetailsFetcher) ReadDetails(res resource.Resource) (resource.Resource, error) {
	val, err := r.reader.ReadResource(terraform.ReadResourceArgs{
		ID: res.TerraformId(),
		Ty: resourceaws.AwsLambdaFunctionResourceType,
		Attributes: map[string]string{
			"function_name": res.TerraformId(),
		},
	})
	if err != nil {
		logrus.Error(err)
		return nil, remoteerror.NewResourceScanningError(err, resourceaws.AwsLambdaFunctionResourceType, res.TerraformId())
	}
	return r.deserializer.DeserializeOne(resourceaws.AwsLambdaFunctionResourceType, *val)
}
