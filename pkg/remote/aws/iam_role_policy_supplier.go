package aws

import (
	"fmt"

	"github.com/cloudskiff/driftctl/pkg/parallel"
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

func NewIamRolePolicySupplier(runner *parallel.ParallelRunner, client iamiface.IAMAPI) *IamRolePolicySupplier {
	return &IamRolePolicySupplier{terraform.Provider(terraform.AWS), awsdeserializer.NewIamRolePolicyDeserializer(), client, terraform.NewParallelResourceReader(runner)}
}

func (s IamRolePolicySupplier) Resources() ([]resource.Resource, error) {
	policies, err := listIamRolePolicies(s.client)
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

func listIamRolePolicies(client iamiface.IAMAPI) ([]string, error) {
	roles, err := listIamRoles(client)
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
			return nil, err
		}
	}

	return resources, nil
}
