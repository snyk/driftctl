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
	ListAllInternetGateways() ([]*ec2.InternetGateway, error)
	ListAllSubnets() ([]*ec2.Subnet, []*ec2.Subnet, error)
	ListAllNatGateways() ([]*ec2.NatGateway, error)
	ListAllRouteTables() ([]*ec2.RouteTable, error)
	ListAllVPCs() ([]*ec2.Vpc, []*ec2.Vpc, error)
	ListAllSecurityGroups() ([]*ec2.SecurityGroup, []*ec2.SecurityGroup, error)
}

type EC2Client interface {
	ec2iface.EC2API
}

type ec2Repository struct {
	client ec2iface.EC2API
	cache  cache.Cache
}

func NewEC2Repository(session *session.Session, c cache.Cache) *ec2Repository {
	return &ec2Repository{
		ec2.New(session),
		c,
	}
}

func (r *ec2Repository) ListAllImages() ([]*ec2.Image, error) {
	if v := r.cache.Get("ec2ListAllImages"); v != nil {
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
	r.cache.Put("ec2ListAllImages", images.Images)
	return images.Images, err
}

func (r *ec2Repository) ListAllSnapshots() ([]*ec2.Snapshot, error) {
	if v := r.cache.Get("ec2ListAllSnapshots"); v != nil {
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
	r.cache.Put("ec2ListAllSnapshots", snapshots)
	return snapshots, err
}

func (r *ec2Repository) ListAllVolumes() ([]*ec2.Volume, error) {
	if v := r.cache.Get("ec2ListAllVolumes"); v != nil {
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
	r.cache.Put("ec2ListAllVolumes", volumes)
	return volumes, nil
}

func (r *ec2Repository) ListAllAddresses() ([]*ec2.Address, error) {
	if v := r.cache.Get("ec2ListAllAddresses"); v != nil {
		return v.([]*ec2.Address), nil
	}

	input := &ec2.DescribeAddressesInput{}
	response, err := r.client.DescribeAddresses(input)
	if err != nil {
		return nil, err
	}
	r.cache.Put("ec2ListAllAddresses", response.Addresses)
	return response.Addresses, nil
}

func (r *ec2Repository) ListAllAddressesAssociation() ([]string, error) {
	if v := r.cache.Get("ec2ListAllAddressesAssociation"); v != nil {
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
	r.cache.Put("ec2ListAllAddressesAssociation", results)
	return results, nil
}

func (r *ec2Repository) ListAllInstances() ([]*ec2.Instance, error) {
	if v := r.cache.Get("ec2ListAllInstances"); v != nil {
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
	r.cache.Put("ec2ListAllInstances", instances)
	return instances, nil
}

func (r *ec2Repository) ListAllKeyPairs() ([]*ec2.KeyPairInfo, error) {
	if v := r.cache.Get("ec2ListAllKeyPairs"); v != nil {
		return v.([]*ec2.KeyPairInfo), nil
	}

	input := &ec2.DescribeKeyPairsInput{}
	pairs, err := r.client.DescribeKeyPairs(input)
	if err != nil {
		return nil, err
	}
	r.cache.Put("ec2ListAllKeyPairs", pairs.KeyPairs)
	return pairs.KeyPairs, err
}

func (r *ec2Repository) ListAllInternetGateways() ([]*ec2.InternetGateway, error) {
	if v := r.cache.Get("ec2ListAllInternetGateways"); v != nil {
		return v.([]*ec2.InternetGateway), nil
	}

	var internetGateways []*ec2.InternetGateway
	input := ec2.DescribeInternetGatewaysInput{}
	err := r.client.DescribeInternetGatewaysPages(&input,
		func(resp *ec2.DescribeInternetGatewaysOutput, lastPage bool) bool {
			internetGateways = append(internetGateways, resp.InternetGateways...)
			return !lastPage
		},
	)
	if err != nil {
		return nil, err
	}
	r.cache.Put("ec2ListAllInternetGateways", internetGateways)
	return internetGateways, nil
}

func (r *ec2Repository) ListAllSubnets() ([]*ec2.Subnet, []*ec2.Subnet, error) {
	cacheSubnets := r.cache.Get("ec2ListAllSubnets")
	cacheDefaultSubnets := r.cache.Get("ec2ListAllDefaultSubnets")
	if cacheSubnets != nil && cacheDefaultSubnets != nil {
		return cacheSubnets.([]*ec2.Subnet), cacheDefaultSubnets.([]*ec2.Subnet), nil
	}

	input := ec2.DescribeSubnetsInput{}
	var subnets []*ec2.Subnet
	var defaultSubnets []*ec2.Subnet
	err := r.client.DescribeSubnetsPages(&input,
		func(resp *ec2.DescribeSubnetsOutput, lastPage bool) bool {
			for _, subnet := range resp.Subnets {
				if subnet.DefaultForAz != nil && *subnet.DefaultForAz {
					defaultSubnets = append(defaultSubnets, subnet)
					continue
				}
				subnets = append(subnets, subnet)
			}
			return !lastPage
		})
	if err != nil {
		return nil, nil, err
	}
	r.cache.Put("ec2ListAllSubnets", subnets)
	r.cache.Put("ec2ListAllDefaultSubnets", defaultSubnets)
	return subnets, defaultSubnets, nil
}

func (r *ec2Repository) ListAllNatGateways() ([]*ec2.NatGateway, error) {
	if v := r.cache.Get("ec2ListAllNatGateways"); v != nil {
		return v.([]*ec2.NatGateway), nil
	}

	var result []*ec2.NatGateway
	input := ec2.DescribeNatGatewaysInput{}
	err := r.client.DescribeNatGatewaysPages(&input,
		func(resp *ec2.DescribeNatGatewaysOutput, lastPage bool) bool {
			result = append(result, resp.NatGateways...)
			return !lastPage
		},
	)

	if err != nil {
		return nil, err
	}

	r.cache.Put("ec2ListAllNatGateways", result)
	return result, nil
}

func (r *ec2Repository) ListAllRouteTables() ([]*ec2.RouteTable, error) {
	if v := r.cache.Get("ec2ListAllRouteTables"); v != nil {
		return v.([]*ec2.RouteTable), nil
	}

	var routeTables []*ec2.RouteTable
	input := ec2.DescribeRouteTablesInput{}
	err := r.client.DescribeRouteTablesPages(&input,
		func(resp *ec2.DescribeRouteTablesOutput, lastPage bool) bool {
			routeTables = append(routeTables, resp.RouteTables...)
			return !lastPage
		},
	)

	if err != nil {
		return nil, err
	}

	r.cache.Put("ec2ListAllRouteTables", routeTables)
	return routeTables, nil
}

func (r *ec2Repository) ListAllVPCs() ([]*ec2.Vpc, []*ec2.Vpc, error) {
	cacheVPCs := r.cache.Get("ec2ListAllVPCs")
	cacheDefaultVPCs := r.cache.Get("ec2ListAllDefaultVPCs")
	if cacheVPCs != nil && cacheDefaultVPCs != nil {
		return cacheVPCs.([]*ec2.Vpc), cacheDefaultVPCs.([]*ec2.Vpc), nil
	}

	input := ec2.DescribeVpcsInput{}
	var VPCs []*ec2.Vpc
	var defaultVPCs []*ec2.Vpc
	err := r.client.DescribeVpcsPages(&input,
		func(resp *ec2.DescribeVpcsOutput, lastPage bool) bool {
			for _, vpc := range resp.Vpcs {
				if vpc.IsDefault != nil && *vpc.IsDefault {
					defaultVPCs = append(defaultVPCs, vpc)
					continue
				}
				VPCs = append(VPCs, vpc)
			}
			return !lastPage
		},
	)
	if err != nil {
		return nil, nil, err
	}

	r.cache.Put("ec2ListAllVPCs", VPCs)
	r.cache.Put("ec2ListAllDefaultVPCs", defaultVPCs)
	return VPCs, defaultVPCs, nil
}

func (r *ec2Repository) ListAllSecurityGroups() ([]*ec2.SecurityGroup, []*ec2.SecurityGroup, error) {
	cacheSecurityGroups := r.cache.Get("ec2ListAllSecurityGroups")
	cacheDefaultSecurityGroups := r.cache.Get("ec2ListAllDefaultSecurityGroups")
	if cacheSecurityGroups != nil && cacheDefaultSecurityGroups != nil {
		return cacheSecurityGroups.([]*ec2.SecurityGroup), cacheDefaultSecurityGroups.([]*ec2.SecurityGroup), nil
	}

	var securityGroups []*ec2.SecurityGroup
	var defaultSecurityGroups []*ec2.SecurityGroup
	input := &ec2.DescribeSecurityGroupsInput{}
	err := r.client.DescribeSecurityGroupsPages(input, func(res *ec2.DescribeSecurityGroupsOutput, lastPage bool) bool {
		for _, securityGroup := range res.SecurityGroups {
			if securityGroup.GroupName != nil && *securityGroup.GroupName == "default" {
				defaultSecurityGroups = append(defaultSecurityGroups, securityGroup)
				continue
			}
			securityGroups = append(securityGroups, securityGroup)
		}
		return !lastPage
	})
	if err != nil {
		return nil, nil, err
	}

	r.cache.Put("ec2ListAllSecurityGroups", securityGroups)
	r.cache.Put("ec2ListAllDefaultSecurityGroups", defaultSecurityGroups)
	return securityGroups, defaultSecurityGroups, nil
}
