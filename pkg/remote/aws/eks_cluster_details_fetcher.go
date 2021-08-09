package aws

import (
	"github.com/cloudskiff/driftctl/pkg/remote/aws/repository"
	remoteerror "github.com/cloudskiff/driftctl/pkg/remote/error"
	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/cloudskiff/driftctl/pkg/terraform"
	"github.com/hashicorp/terraform/flatmap"
)

type EKSClusterDetailsFetcher struct {
	reader       terraform.ResourceReader
	deserializer *resource.Deserializer
	repo         repository.EKSRepository
}

func NewEKSClusterDetailsFetcher(provider terraform.ResourceReader, deserializer *resource.Deserializer, repo repository.EKSRepository) *EKSClusterDetailsFetcher {
	return &EKSClusterDetailsFetcher{
		reader:       provider,
		deserializer: deserializer,
		repo:         repo,
	}
}

func (r *EKSClusterDetailsFetcher) ReadDetails(res resource.Resource) (resource.Resource, error) {
	info, err := r.repo.DescribeCluster(res.TerraformId())
	if err != nil {
		return nil, remoteerror.NewResourceScanningError(err, res.TerraformType(), res.TerraformId())
	}

	var attrs = map[string]interface{}{
		"tags": map[string]string{},
	}
	for k, v := range info.Tags {
		(attrs["tags"].(map[string]string))[k] = *v
	}
	if val := info.Endpoint; val != nil {
		attrs["endpoint"] = *val
	}
	if val := info.PlatformVersion; val != nil {
		attrs["platform_version"] = *val
	}
	if val := info.Status; val != nil {
		attrs["status"] = *val
	}
	if val := info.Version; val != nil {
		attrs["version"] = *val
	}
	if val := info.KubernetesNetworkConfig; val != nil {
		attrs["kubernetes_network_config"] = map[string]string{
			"service_ipv4_cidr:": *val.ServiceIpv4Cidr,
		}
	}
	if val := info.RoleArn; val != nil {
		attrs["role_arn"] = *val
	}
	if val := info.Arn; val != nil {
		attrs["arn"] = *val
	}

	ctyVal, err := r.reader.ReadResource(terraform.ReadResourceArgs{
		Ty:         aws.AwsEKSClusterResourceType,
		ID:         res.TerraformId(),
		Attributes: flatmap.Flatten(attrs),
	})
	if err != nil {
		return nil, remoteerror.NewResourceScanningError(err, res.TerraformType(), res.TerraformId())
	}
	deserializedRes, err := r.deserializer.DeserializeOne(aws.AwsEKSClusterResourceType, *ctyVal)
	if err != nil {
		return nil, err
	}

	return deserializedRes, nil
}
