package aws

import (
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	"github.com/cloudskiff/driftctl/pkg/remote/deserializer"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"
	awsdeserializer "github.com/cloudskiff/driftctl/pkg/resource/aws/deserializer"
	"github.com/cloudskiff/driftctl/pkg/terraform"

	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

type IamUserPolicyAttachmentSupplier struct {
	reader       terraform.ResourceReader
	deserializer deserializer.CTYDeserializer
	repo         repository.IAMRepository
	runner       *terraform.ParallelResourceReader
}

func NewIamUserPolicyAttachmentSupplier(provider *AWSTerraformProvider) *IamUserPolicyAttachmentSupplier {
	return &IamUserPolicyAttachmentSupplier{
		provider,
		awsdeserializer.NewIamUserPolicyAttachmentDeserializer(),
		repository.NewIAMRepository(provider.session),
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
	}
}

func (s *IamUserPolicyAttachmentSupplier) Resources() ([]resource.Resource, error) {
	users, err := s.repo.ListAllUsers()
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationErrorWithType(err, resourceaws.AwsIamUserPolicyAttachmentResourceType, resourceaws.AwsIamUserResourceType)
	}
	policyAttachments, err := s.repo.ListAllUserPolicyAttachments(users)
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationError(err, resourceaws.AwsIamUserPolicyAttachmentResourceType)
	}
	results := make([]cty.Value, 0)
	if len(policyAttachments) > 0 {
		for _, attachedPolicy := range policyAttachments {
			attached := *attachedPolicy
			s.runner.Run(func() (cty.Value, error) {
				return s.readUserPolicyAttachment(&attached)
			})
		}
		results, err = s.runner.Wait()
		if err != nil {
			return nil, err
		}
	}

	return s.deserializer.Deserialize(results)
}

func (s *IamUserPolicyAttachmentSupplier) readUserPolicyAttachment(attachedPol *repository.AttachedUserPolicy) (cty.Value, error) {
	res, err := s.reader.ReadResource(
		terraform.ReadResourceArgs{
			Ty: resourceaws.AwsIamUserPolicyAttachmentResourceType,
			ID: *attachedPol.PolicyName,
			Attributes: map[string]string{
				"user":       attachedPol.UserName,
				"policy_arn": *attachedPol.PolicyArn,
			},
		},
	)

	if err != nil {
		logrus.Warnf("Error reading iam user policy attachment %s[%s]: %+v", attachedPol, resourceaws.AwsIamUserPolicyAttachmentResourceType, err)
		return cty.NilVal, err
	}
	return *res, nil
}
