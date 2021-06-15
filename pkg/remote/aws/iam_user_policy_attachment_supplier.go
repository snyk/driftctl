package aws

import (
	"fmt"

	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"

	"github.com/cloudskiff/driftctl/pkg/resource"
	resourceaws "github.com/cloudskiff/driftctl/pkg/resource/aws"

	"github.com/cloudskiff/driftctl/pkg/terraform"

	"github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

type IamUserPolicyAttachmentSupplier struct {
	reader       terraform.ResourceReader
	deserializer *resource.Deserializer
	repo         repository.IAMRepository
	runner       *terraform.ParallelResourceReader
}

func NewIamUserPolicyAttachmentSupplier(provider *AWSTerraformProvider, deserializer *resource.Deserializer, repo repository.IAMRepository) *IamUserPolicyAttachmentSupplier {
	return &IamUserPolicyAttachmentSupplier{
		provider,
		deserializer,
		repo,
		terraform.NewParallelResourceReader(provider.Runner().SubRunner()),
	}
}

func (s *IamUserPolicyAttachmentSupplier) SuppliedType() resource.ResourceType {
	return resourceaws.AwsIamUserPolicyAttachmentResourceType
}

func (s *IamUserPolicyAttachmentSupplier) Resources() ([]resource.Resource, error) {
	users, err := s.repo.ListAllUsers()
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationErrorWithType(err, s.SuppliedType(), resourceaws.AwsIamUserResourceType)
	}
	policyAttachments, err := s.repo.ListAllUserPolicyAttachments(users)
	if err != nil {
		return nil, remoteerror.NewResourceEnumerationError(err, s.SuppliedType())
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

	return s.deserializer.Deserialize(s.SuppliedType(), results)
}

func (s *IamUserPolicyAttachmentSupplier) readUserPolicyAttachment(attachedPol *repository.AttachedUserPolicy) (cty.Value, error) {
	res, err := s.reader.ReadResource(
		terraform.ReadResourceArgs{
			Ty: s.SuppliedType(),
			ID: fmt.Sprintf("%s-%s", *attachedPol.PolicyName, attachedPol.UserName),
			Attributes: map[string]string{
				"user":       attachedPol.UserName,
				"policy_arn": *attachedPol.PolicyArn,
			},
		},
	)

	if err != nil {
		logrus.Warnf("Error reading iam user policy attachment %s[%s]: %+v", attachedPol, s.SuppliedType(), err)
		return cty.NilVal, err
	}
	return *res, nil
}
