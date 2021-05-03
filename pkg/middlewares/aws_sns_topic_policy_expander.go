package middlewares

import (
	"github.com/sirupsen/logrus"

	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
)

// Explodes policy found in aws_sns_topic from state resources to aws_sns_topic_policy resources
type AwsSNSTopicPolicyExpander struct {
	resourceFactory          resource.ResourceFactory
	resourceSchemaRepository resource.SchemaRepositoryInterface
}

func NewAwsSNSTopicPolicyExpander(resourceFactory resource.ResourceFactory, resourceSchemaRepository resource.SchemaRepositoryInterface) AwsSNSTopicPolicyExpander {
	return AwsSNSTopicPolicyExpander{
		resourceFactory,
		resourceSchemaRepository,
	}
}

func (m AwsSNSTopicPolicyExpander) Execute(_, resourcesFromState *[]resource.Resource) error {
	newList := make([]resource.Resource, 0)
	for _, res := range *resourcesFromState {
		// Ignore all resources other than sns_topic
		if res.TerraformType() != aws.AwsSnsTopicResourceType {
			newList = append(newList, res)
			continue
		}

		topic, _ := res.(*aws.AwsSnsTopic)
		newList = append(newList, res)

		if m.hasPolicyAttached(topic, resourcesFromState) {
			topic.Policy = nil
			continue
		}

		err := m.splitPolicy(topic, &newList)
		if err != nil {
			return err
		}
	}
	*resourcesFromState = newList
	return nil
}

func (m *AwsSNSTopicPolicyExpander) splitPolicy(topic *aws.AwsSnsTopic, results *[]resource.Resource) error {
	if topic.Policy == nil || *topic.Policy == "" {
		return nil
	}

	data := map[string]interface{}{
		"arn":    topic.Arn,
		"id":     topic.Id,
		"policy": topic.Policy,
	}
	ctyVal, err := m.resourceFactory.CreateResource(data, "aws_sns_topic_policy")
	if err != nil {
		return err
	}

	schema, exist := m.resourceSchemaRepository.GetSchema("aws_sns_topic_policy")
	ctyAttr := resource.ToResourceAttributes(ctyVal)
	ctyAttr.SanitizeDefaultsV3()
	if exist && schema.NormalizeFunc != nil {
		schema.NormalizeFunc(ctyAttr)
	}

	newPolicy := &resource.AbstractResource{
		Id:    topic.Id,
		Type:  aws.AwsSnsTopicPolicyResourceType,
		Attrs: ctyAttr,
	}

	*results = append(*results, newPolicy)
	logrus.WithFields(logrus.Fields{
		"id": newPolicy.TerraformId(),
	}).Debug("Created new policy from sns_topic")

	topic.Policy = nil
	return nil
}

func (m *AwsSNSTopicPolicyExpander) hasPolicyAttached(topic *aws.AwsSnsTopic, resourcesFromState *[]resource.Resource) bool {
	for _, res := range *resourcesFromState {
		if res.TerraformType() == aws.AwsSnsTopicPolicyResourceType &&
			res.TerraformId() == topic.Id {
			return true
		}
	}
	return false
}
