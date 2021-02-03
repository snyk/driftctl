package aws

import (
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"

	"github.com/cloudskiff/driftctl/pkg/remote/deserializer"
	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"
	awsdeserializer "github.com/cloudskiff/driftctl/pkg/resource/aws/deserializer"
	"github.com/cloudskiff/driftctl/pkg/terraform"

	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

type IamAccessKeySupplier struct {
	reader       terraform.ResourceReader
	deserializer deserializer.CTYDeserializer
	client       iamiface.IAMAPI
	runner       *terraform.ParallelResourceReader
}

func NewIamAccessKeySupplier(provider *TerraformProvider) *IamAccessKeySupplier {
	return &IamAccessKeySupplier{
		provider,
		awsdeserializer.NewIamAccessKeyDeserializer(),
		iam.New(provider.session),
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
	}
}

func (s IamAccessKeySupplier) Resources() ([]resource.Resource, error) {
	keys, err := listIamAccessKeys(s.client)
	if err != nil {
		return nil, err
	}
	results := make([]cty.Value, 0)
	if len(keys) > 0 {
		for _, key := range keys {
			k := *key
			s.runner.Run(func() (cty.Value, error) {
				return s.readRes(&k)
			})
		}
		results, err = s.runner.Wait()
		if err != nil {
			return nil, err
		}
	}
	return s.deserializer.Deserialize(results)
}

func (s IamAccessKeySupplier) readRes(key *iam.AccessKeyMetadata) (cty.Value, error) {
	res, err := s.reader.ReadResource(
		terraform.ReadResourceArgs{
			Ty: resourceaws.AwsIamAccessKeyResourceType,
			ID: *key.AccessKeyId,
			Attributes: map[string]string{
				"user": *key.UserName,
			},
		},
	)
	if err != nil {
		logrus.Warnf("Error reading iam access key %s[%s]: %+v", *key.AccessKeyId, resourceaws.AwsIamAccessKeyResourceType, err)
		return cty.NilVal, err
	}

	return *res, nil
}

func listIamAccessKeys(client iamiface.IAMAPI) ([]*iam.AccessKeyMetadata, error) {
	users, err := listIamUsers(client, resourceaws.AwsIamAccessKeyResourceType)
	if err != nil {
		return nil, err
	}
	var resources []*iam.AccessKeyMetadata
	for _, user := range users {
		input := &iam.ListAccessKeysInput{
			UserName: user.UserName,
		}
		err := client.ListAccessKeysPages(input, func(res *iam.ListAccessKeysOutput, lastPage bool) bool {
			resources = append(resources, res.AccessKeyMetadata...)
			return !lastPage
		})
		if err != nil {
			return nil, remoteerror.NewResourceEnumerationError(err, resourceaws.AwsIamAccessKeyResourceType)
		}
	}

	return resources, nil
}
