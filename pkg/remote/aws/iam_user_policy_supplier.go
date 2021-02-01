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

type IamUserPolicySupplier struct {
	reader       terraform.ResourceReader
	deserializer deserializer.CTYDeserializer
	client       iamiface.IAMAPI
	runner       *terraform.ParallelResourceReader
}

func NewIamUserPolicySupplier(provider *TerraformProvider) *IamUserPolicySupplier {
	return &IamUserPolicySupplier{
		provider,
		awsdeserializer.NewIamUserPolicyDeserializer(),
		iam.New(provider.session),
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
	}
}

func (s IamUserPolicySupplier) Resources() ([]resource.Resource, error) {
	users, err := listIamUsers(s.client, resourceaws.AwsIamUserPolicyResourceType)
	if err != nil {
		return nil, err
	}
	results := make([]cty.Value, 0)
	if len(users) > 0 {
		policies := make([]string, 0)
		for _, user := range users {
			userName := *user.UserName
			policyList, err := listIamUserPolicies(userName, s.client)
			if err != nil {
				return nil, remoteerror.NewResourceEnumerationError(err, resourceaws.AwsIamUserPolicyResourceType)
			}
			for _, polName := range policyList {
				policies = append(policies, fmt.Sprintf("%s:%s", userName, *polName))
			}
		}

		for _, policy := range policies {
			polName := policy
			s.runner.Run(func() (cty.Value, error) {
				return s.readRes(polName)
			})
		}
		results, err = s.runner.Wait()
		if err != nil {
			return nil, err
		}
	}
	return s.deserializer.Deserialize(results)
}

func (s IamUserPolicySupplier) readRes(policyName string) (cty.Value, error) {
	res, err := s.reader.ReadResource(
		terraform.ReadResourceArgs{
			Ty: resourceaws.AwsIamUserPolicyResourceType,
			ID: policyName,
		},
	)
	if err != nil {
		logrus.Warnf("Error reading iam user policy %s[%s]: %+v", policyName, resourceaws.AwsIamUserResourceType, err)
		return cty.NilVal, err
	}

	return *res, nil
}

func listIamUserPolicies(username string, client iamiface.IAMAPI) ([]*string, error) {
	var policyNames []*string
	input := &iam.ListUserPoliciesInput{
		UserName: &username,
	}
	err := client.ListUserPoliciesPages(input, func(res *iam.ListUserPoliciesOutput, lastPage bool) bool {
		policyNames = append(policyNames, res.PolicyNames...)
		return !lastPage
	})
	if err != nil {
		return nil, err
	}
	return policyNames, nil
}
