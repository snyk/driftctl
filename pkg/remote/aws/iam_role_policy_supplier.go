package aws

import (
	"fmt"

	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	awsdeserializer "github.com/cloudskiff/driftctl/pkg/resource/aws/deserializer"

	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
	"github.com/cloudskiff/driftctl/pkg/remote/deserializer"
	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/cloudskiff/driftctl/pkg/terraform"

	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

type IamRolePolicySupplier struct {
	reader       terraform.ResourceReader
	deserializer deserializer.CTYDeserializer
	client       iamiface.IAMAPI
	runner       *terraform.ParallelResourceReader
}

func NewIamRolePolicySupplier(provider *TerraformProvider) *IamRolePolicySupplier {
	return &IamRolePolicySupplier{
		provider,
		awsdeserializer.NewIamRolePolicyDeserializer(),
		iam.New(provider.session),
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
	}
}

func (s IamRolePolicySupplier) Resources() ([]resource.Resource, error) {
	policies, err := listIamRolePolicies(s.client, resourceaws.AwsIamRolePolicyResourceType)
	if err != nil {
		return nil, err
	}
	for _, policyName := range policies {
		name := policyName
		s.runner.Run(func() (cty.Value, error) {
			return s.readRes(name)
		})
	}
	results, err := s.runner.Wait()
	if err != nil {
		return nil, err
	}

	return s.deserializer.Deserialize(results)
}

func (s IamRolePolicySupplier) readRes(name string) (cty.Value, error) {
	res, err := s.reader.ReadResource(
		terraform.ReadResourceArgs{
			Ty: resourceaws.AwsIamRolePolicyResourceType,
			ID: name,
		},
	)
	if err != nil {
		logrus.Warnf("Error reading iam role policy %s[%s]: %+v", name, resourceaws.AwsIamRolePolicyResourceType, err)
		return cty.NilVal, err
	}

	return *res, nil
}

func listIamRolePolicies(client iamiface.IAMAPI, supplierType string) ([]string, error) {
	roles, err := listIamRoles(client, supplierType)
	if err != nil {
		return nil, err
	}

	var resources []string
	for _, role := range roles {
		input := &iam.ListRolePoliciesInput{
			RoleName: role.RoleName,
		}

		err := client.ListRolePoliciesPages(input, func(res *iam.ListRolePoliciesOutput, lastPage bool) bool {
			for _, policy := range res.PolicyNames {
				resources = append(
					resources,
					fmt.Sprintf(
						"%s:%s",
						*role.RoleName,
						*policy,
					),
				)
			}
			return !lastPage
		})
		if err != nil {
			return nil, remoteerror.NewResourceEnumerationErrorWithType(err, supplierType, resourceaws.AwsIamRoleResourceType)
		}
	}

	return resources, nil
}
