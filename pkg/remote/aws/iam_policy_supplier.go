package aws

import (
	"github.com/aws/aws-sdk-go/aws"
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

type IamPolicySupplier struct {
	reader       terraform.ResourceReader
	deserializer deserializer.CTYDeserializer
	client       iamiface.IAMAPI
	runner       *terraform.ParallelResourceReader
}

func NewIamPolicySupplier(provider *TerraformProvider) *IamPolicySupplier {
	return &IamPolicySupplier{
		provider,
		awsdeserializer.NewIamPolicyDeserializer(),
		iam.New(provider.session),
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
	}
}

func (s IamPolicySupplier) Resources() ([]resource.Resource, error) {
	policies, err := listIamPolicies(s.client)
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationError(err, resourceaws.AwsIamPolicyResourceType)
	}
	results := make([]cty.Value, 0)
	if len(policies) > 0 {
		for _, policy := range policies {
			u := *policy
			s.runner.Run(func() (cty.Value, error) {
				return s.readRes(&u)
			})
		}
		results, err = s.runner.Wait()
		if err != nil {
			return nil, err
		}
	}
	return s.deserializer.Deserialize(results)
}

func (s IamPolicySupplier) readRes(resource *iam.Policy) (cty.Value, error) {
	res, err := s.reader.ReadResource(
		terraform.ReadResourceArgs{
			Ty: resourceaws.AwsIamPolicyResourceType,
			ID: *resource.Arn,
		},
	)
	if err != nil {
		logrus.Warnf("Error reading iam policy %s[%s]: %+v", *resource.Arn, resourceaws.AwsIamPolicyResourceType, err)
		return cty.NilVal, err
	}

	return *res, nil
}

func listIamPolicies(client iamiface.IAMAPI) ([]*iam.Policy, error) {
	var resources []*iam.Policy
	input := &iam.ListPoliciesInput{
		Scope: aws.String(iam.PolicyScopeTypeLocal),
	}
	err := client.ListPoliciesPages(input, func(res *iam.ListPoliciesOutput, lastPage bool) bool {
		resources = append(resources, res.Policies...)
		return !lastPage
	})
	if err != nil {
		return nil, err
	}
	return resources, nil
}
