package repository

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/cloudskiff/driftctl/pkg/remote/cache"
)

type EC2Repository interface {
	ListAllImages() ([]*ec2.Image, error)
	ListAllSnapshots() ([]*ec2.Snapshot, error)
	ListAllVolumes() ([]*ec2.Volume, error)
	ListAllAddresses() ([]*ec2.Address, error)
	ListAllAddressesAssociation() ([]string, error)
	ListAllInstances() ([]*ec2.Instance, error)
	ListAllKeyPairs() ([]*ec2.KeyPairInfo, error)
}

type EC2Client interface {
	ec2iface.EC2API
}

type ec2Repository struct {
	client ec2iface.EC2API
	cache  *cache.Cache
}

func NewEC2Repository(session *session.Session) *ec2Repository {
	return &ec2Repository{
		ec2.New(session),
		cache.New(),
	}
}

func (r *ec2Repository) ListAllImages() ([]*ec2.Image, error) {
	if v := r.cache.Get("ec2AllImages"); v != nil {
		return v.([]*ec2.Image), nil
	}

	input := &ec2.DescribeImagesInput{
		Owners: []*string{
			aws.String("self"),
		},
	}
	images, err := r.client.DescribeImages(input)
	if err != nil {
		return nil, err
	}
	r.cache.Set("ec2AllImages", images.Images)
	return images.Images, err
}

func (r *ec2Repository) ListAllSnapshots() ([]*ec2.Snapshot, error) {
	if v := r.cache.Get("ec2AllSnapshots"); v != nil {
		return v.([]*ec2.Snapshot), nil
	}

	var snapshots []*ec2.Snapshot
	input := &ec2.DescribeSnapshotsInput{
		OwnerIds: []*string{
			aws.String("self"),
		},
	}
	err := r.client.DescribeSnapshotsPages(input, func(res *ec2.DescribeSnapshotsOutput, lastPage bool) bool {
		snapshots = append(snapshots, res.Snapshots...)
		return !lastPage
	})
	if err != nil {
		return nil, err
	}
	r.cache.Set("ec2AllSnapshots", snapshots)
	return snapshots, err
}

func (r *ec2Repository) ListAllVolumes() ([]*ec2.Volume, error) {
	if v := r.cache.Get("ec2AllVolumes"); v != nil {
		return v.([]*ec2.Volume), nil
	}

	var volumes []*ec2.Volume
	input := &ec2.DescribeVolumesInput{}
	err := r.client.DescribeVolumesPages(input, func(res *ec2.DescribeVolumesOutput, lastPage bool) bool {
		volumes = append(volumes, res.Volumes...)
		return !lastPage
	})
	if err != nil {
		return nil, err
	}
	r.cache.Set("ec2AllVolumes", volumes)
	return volumes, nil
}

func (r *ec2Repository) ListAllAddresses() ([]*ec2.Address, error) {
	if v := r.cache.Get("ec2AllAddresses"); v != nil {
		return v.([]*ec2.Address), nil
	}

	input := &ec2.DescribeAddressesInput{}
	response, err := r.client.DescribeAddresses(input)
	if err != nil {
		return nil, err
	}
	r.cache.Set("ec2AllAddresses", response.Addresses)
	return response.Addresses, nil
}

func (r *ec2Repository) ListAllAddressesAssociation() ([]string, error) {
	if v := r.cache.Get("ec2AddressesAssociation"); v != nil {
		return v.([]string), nil
	}

	results := make([]string, 0)
	addresses, err := r.ListAllAddresses()
	if err != nil {
		return nil, err
	}
	for _, address := range addresses {
		if address.AssociationId != nil {
			results = append(results, aws.StringValue(address.AssociationId))
		}
	}
	r.cache.Set("ec2AddressesAssociation", results)
	return results, nil
}

func (r *ec2Repository) ListAllInstances() ([]*ec2.Instance, error) {
	if v := r.cache.Get("ec2AllInstances"); v != nil {
		return v.([]*ec2.Instance), nil
	}

	var instances []*ec2.Instance
	input := &ec2.DescribeInstancesInput{}
	err := r.client.DescribeInstancesPages(input, func(res *ec2.DescribeInstancesOutput, lastPage bool) bool {
		for _, reservation := range res.Reservations {
			instances = append(instances, reservation.Instances...)
		}
		return !lastPage
	})
	if err != nil {
		return nil, err
	}
	r.cache.Set("ec2AllInstances", instances)
	return instances, nil
}

func (r *ec2Repository) ListAllKeyPairs() ([]*ec2.KeyPairInfo, error) {
	if v := r.cache.Get("ec2AllKeyPairs"); v != nil {
		return v.([]*ec2.KeyPairInfo), nil
	}

	input := &ec2.DescribeKeyPairsInput{}
	pairs, err := r.client.DescribeKeyPairs(input)
	if err != nil {
		return nil, err
	}
	r.cache.Set("ec2AllKeyPairs", pairs.KeyPairs)
	return pairs.KeyPairs, err
}
