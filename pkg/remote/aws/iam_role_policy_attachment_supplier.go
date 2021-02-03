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

type IamRolePolicyAttachmentSupplier struct {
	reader       terraform.ResourceReader
	deserializer deserializer.CTYDeserializer
	client       iamiface.IAMAPI
	runner       *terraform.ParallelResourceReader
}

func NewIamRolePolicyAttachmentSupplier(provider *TerraformProvider) *IamRolePolicyAttachmentSupplier {
	return &IamRolePolicyAttachmentSupplier{
		provider,
		awsdeserializer.NewIamRolePolicyAttachmentDeserializer(),
		iam.New(provider.session),
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
	}
}

func (s IamRolePolicyAttachmentSupplier) Resources() ([]resource.Resource, error) {
	roles, err := listIamRoles(s.client, resourceaws.AwsIamRolePolicyAttachmentResourceType)
	if err != nil {
		return nil, err
	}
	results := make([]cty.Value, 0)
	if len(roles) > 0 {
		attachedPolicies := make([]*attachedRolePolicy, 0)
		for _, role := range roles {
			roleName := *role.RoleName
			if awsIamRoleShouldBeIgnored(roleName) {
				continue
			}
			roleAttachmentList, err := listIamRolePoliciesAttachment(roleName, s.client)
			if err != nil {
				return nil, err
			}
			attachedPolicies = append(attachedPolicies, roleAttachmentList...)
		}

		for _, attachedPolicy := range attachedPolicies {
			attached := *attachedPolicy
			s.runner.Run(func() (cty.Value, error) {
				return s.readRes(attached)
			})
		}
		results, err = s.runner.Wait()
		if err != nil {
			return nil, err
		}
	}

	return s.deserializer.Deserialize(results)
}

func (s IamRolePolicyAttachmentSupplier) readRes(attachedPol attachedRolePolicy) (cty.Value, error) {
	res, err := s.reader.ReadResource(
		terraform.ReadResourceArgs{
			Ty: resourceaws.AwsIamRolePolicyAttachmentResourceType,
			ID: *attachedPol.PolicyName,
			Attributes: map[string]string{
				"role":       attachedPol.RoleName,
				"policy_arn": *attachedPol.PolicyArn,
			},
		},
	)

	if err != nil {
		logrus.Warnf("Error reading iam role policy attachment %s[%s]: %+v", attachedPol, resourceaws.AwsIamRolePolicyAttachmentResourceType, err)
		return cty.NilVal, err
	}
	return *res, nil
}

func listIamRolePoliciesAttachment(roleName string, client iamiface.IAMAPI) ([]*attachedRolePolicy, error) {
	var attachedRolePolicies []*attachedRolePolicy
	input := &iam.ListAttachedRolePoliciesInput{
		RoleName: &roleName,
	}
	err := client.ListAttachedRolePoliciesPages(input, func(res *iam.ListAttachedRolePoliciesOutput, lastPage bool) bool {
		for _, policy := range res.AttachedPolicies {
			attachedRolePolicies = append(attachedRolePolicies, &attachedRolePolicy{
				AttachedPolicy: *policy,
				RoleName:       roleName,
			})
		}
		return !lastPage
	})
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationErrorWithType(err, resourceaws.AwsIamRolePolicyAttachmentResourceType, resourceaws.AwsIamRolePolicyResourceType)
	}
	return attachedRolePolicies, nil
}

type attachedRolePolicy struct {
	iam.AttachedPolicy
	RoleName string
}
