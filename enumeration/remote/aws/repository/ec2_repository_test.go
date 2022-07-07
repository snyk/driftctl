package repository

import (
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/snyk/driftctl/enumeration/remote/cache"
	awstest "github.com/snyk/driftctl/test/aws"
	"github.com/stretchr/testify/mock"

	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/r3labs/diff/v2"
	"github.com/stretchr/testify/assert"
)

func Test_ec2Repository_ListAllImages(t *testing.T) {

	tests := []struct {
		name    string
		mocks   func(client *awstest.MockFakeEC2)
		want    []*ec2.Image
		wantErr error
	}{
		{
			name: "List all images",
			mocks: func(client *awstest.MockFakeEC2) {
				client.On("DescribeImages",
					&ec2.DescribeImagesInput{
						Owners: []*string{
							aws.String("self"),
						},
					}).Return(&ec2.DescribeImagesOutput{
					Images: []*ec2.Image{
						{ImageId: aws.String("1")},
						{ImageId: aws.String("2")},
						{ImageId: aws.String("3")},
						{ImageId: aws.String("4")},
					},
				}, nil).Once()
			},
			want: []*ec2.Image{
				{ImageId: aws.String("1")},
				{ImageId: aws.String("2")},
				{ImageId: aws.String("3")},
				{ImageId: aws.String("4")},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := cache.New(1)
			client := &awstest.MockFakeEC2{}
			tt.mocks(client)
			r := &ec2Repository{
				client: client,
				cache:  store,
			}
			got, err := r.ListAllImages()
			assert.Equal(t, tt.wantErr, err)

			if err == nil {
				// Check that results were cached
				cachedData, err := r.ListAllImages()
				assert.NoError(t, err)
				assert.Equal(t, got, cachedData)
				assert.IsType(t, []*ec2.Image{}, store.Get("ec2ListAllImages"))
			}

			changelog, err := diff.Diff(got, tt.want)
			assert.Nil(t, err)
			if len(changelog) > 0 {
				for _, change := range changelog {
					t.Errorf("%s: %s -> %s", strings.Join(change.Path, "."), change.From, change.To)
				}
				t.Fail()
			}
		})
	}
}

func Test_ec2Repository_ListAllSnapshots(t *testing.T) {
	tests := []struct {
		name    string
		mocks   func(client *awstest.MockFakeEC2)
		want    []*ec2.Snapshot
		wantErr error
	}{
		{name: "List with 2 pages",
			mocks: func(client *awstest.MockFakeEC2) {
				client.On("DescribeSnapshotsPages",
					&ec2.DescribeSnapshotsInput{
						OwnerIds: []*string{
							aws.String("self"),
						},
					},
					mock.MatchedBy(func(callback func(res *ec2.DescribeSnapshotsOutput, lastPage bool) bool) bool {
						callback(&ec2.DescribeSnapshotsOutput{
							Snapshots: []*ec2.Snapshot{
								{VolumeId: aws.String("1")},
								{VolumeId: aws.String("2")},
								{VolumeId: aws.String("3")},
								{VolumeId: aws.String("4")},
							},
						}, false)
						callback(&ec2.DescribeSnapshotsOutput{
							Snapshots: []*ec2.Snapshot{
								{VolumeId: aws.String("5")},
								{VolumeId: aws.String("6")},
								{VolumeId: aws.String("7")},
								{VolumeId: aws.String("8")},
							},
						}, true)
						return true
					})).Return(nil).Once()
			},
			want: []*ec2.Snapshot{
				{VolumeId: aws.String("1")},
				{VolumeId: aws.String("2")},
				{VolumeId: aws.String("3")},
				{VolumeId: aws.String("4")},
				{VolumeId: aws.String("5")},
				{VolumeId: aws.String("6")},
				{VolumeId: aws.String("7")},
				{VolumeId: aws.String("8")},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := cache.New(1)
			client := &awstest.MockFakeEC2{}
			tt.mocks(client)
			r := &ec2Repository{
				client: client,
				cache:  store,
			}
			got, err := r.ListAllSnapshots()
			assert.Equal(t, tt.wantErr, err)

			if err == nil {
				// Check that results were cached
				cachedData, err := r.ListAllSnapshots()
				assert.NoError(t, err)
				assert.Equal(t, got, cachedData)
				assert.IsType(t, []*ec2.Snapshot{}, store.Get("ec2ListAllSnapshots"))
			}

			changelog, err := diff.Diff(got, tt.want)
			assert.Nil(t, err)
			if len(changelog) > 0 {
				for _, change := range changelog {
					t.Errorf("%s: %s -> %s", strings.Join(change.Path, "."), change.From, change.To)
				}
				t.Fail()
			}
		})
	}
}

func Test_ec2Repository_ListAllVolumes(t *testing.T) {
	tests := []struct {
		name    string
		mocks   func(client *awstest.MockFakeEC2)
		want    []*ec2.Volume
		wantErr error
	}{
		{name: "List with 2 pages",
			mocks: func(client *awstest.MockFakeEC2) {
				client.On("DescribeVolumesPages",
					&ec2.DescribeVolumesInput{},
					mock.MatchedBy(func(callback func(res *ec2.DescribeVolumesOutput, lastPage bool) bool) bool {
						callback(&ec2.DescribeVolumesOutput{
							Volumes: []*ec2.Volume{
								{VolumeId: aws.String("1")},
								{VolumeId: aws.String("2")},
								{VolumeId: aws.String("3")},
								{VolumeId: aws.String("4")},
							},
						}, false)
						callback(&ec2.DescribeVolumesOutput{
							Volumes: []*ec2.Volume{
								{VolumeId: aws.String("5")},
								{VolumeId: aws.String("6")},
								{VolumeId: aws.String("7")},
								{VolumeId: aws.String("8")},
							},
						}, true)
						return true
					})).Return(nil).Once()
			},
			want: []*ec2.Volume{
				{VolumeId: aws.String("1")},
				{VolumeId: aws.String("2")},
				{VolumeId: aws.String("3")},
				{VolumeId: aws.String("4")},
				{VolumeId: aws.String("5")},
				{VolumeId: aws.String("6")},
				{VolumeId: aws.String("7")},
				{VolumeId: aws.String("8")},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := cache.New(1)
			client := &awstest.MockFakeEC2{}
			tt.mocks(client)
			r := &ec2Repository{
				client: client,
				cache:  store,
			}
			got, err := r.ListAllVolumes()
			assert.Equal(t, tt.wantErr, err)

			if err == nil {
				// Check that results were cached
				cachedData, err := r.ListAllVolumes()
				assert.NoError(t, err)
				assert.Equal(t, got, cachedData)
				assert.IsType(t, []*ec2.Volume{}, store.Get("ec2ListAllVolumes"))
			}

			changelog, err := diff.Diff(got, tt.want)
			assert.Nil(t, err)
			if len(changelog) > 0 {
				for _, change := range changelog {
					t.Errorf("%s: %s -> %s", strings.Join(change.Path, "."), change.From, change.To)
				}
				t.Fail()
			}
		})
	}
}

func Test_ec2Repository_ListAllAddresses(t *testing.T) {
	tests := []struct {
		name    string
		mocks   func(client *awstest.MockFakeEC2)
		want    []*ec2.Address
		wantErr error
	}{
		{
			name: "List address",
			mocks: func(client *awstest.MockFakeEC2) {
				client.On("DescribeAddresses", &ec2.DescribeAddressesInput{}).
					Return(&ec2.DescribeAddressesOutput{
						Addresses: []*ec2.Address{
							{AssociationId: aws.String("1")},
							{AssociationId: aws.String("2")},
							{AssociationId: aws.String("3")},
							{AssociationId: aws.String("4")},
						},
					}, nil).Once()
			},
			want: []*ec2.Address{
				{AssociationId: aws.String("1")},
				{AssociationId: aws.String("2")},
				{AssociationId: aws.String("3")},
				{AssociationId: aws.String("4")},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := cache.New(1)
			client := &awstest.MockFakeEC2{}
			tt.mocks(client)
			r := &ec2Repository{
				client: client,
				cache:  store,
			}
			got, err := r.ListAllAddresses()
			assert.Equal(t, tt.wantErr, err)

			if err == nil {
				// Check that results were cached
				cachedData, err := r.ListAllAddresses()
				assert.NoError(t, err)
				assert.Equal(t, got, cachedData)
				assert.IsType(t, []*ec2.Address{}, store.Get("ec2ListAllAddresses"))
			}

			changelog, err := diff.Diff(got, tt.want)
			assert.Nil(t, err)
			if len(changelog) > 0 {
				for _, change := range changelog {
					t.Errorf("%s: %s -> %s", strings.Join(change.Path, "."), change.From, change.To)
				}
				t.Fail()
			}
		})
	}
}

func Test_ec2Repository_ListAllAddressesAssociation(t *testing.T) {
	tests := []struct {
		name    string
		mocks   func(client *awstest.MockFakeEC2)
		want    []*ec2.Address
		wantErr error
	}{
		{
			name: "List address",
			mocks: func(client *awstest.MockFakeEC2) {
				client.On("DescribeAddresses", &ec2.DescribeAddressesInput{}).
					Return(&ec2.DescribeAddressesOutput{
						Addresses: []*ec2.Address{
							{AssociationId: aws.String("1")},
							{AssociationId: aws.String("2")},
							{AssociationId: aws.String("3")},
							{AssociationId: aws.String("4")},
						},
					}, nil).Once()
			},
			want: []*ec2.Address{
				{AssociationId: aws.String("1")},
				{AssociationId: aws.String("2")},
				{AssociationId: aws.String("3")},
				{AssociationId: aws.String("4")},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := cache.New(1)
			client := &awstest.MockFakeEC2{}
			tt.mocks(client)
			r := &ec2Repository{
				client: client,
				cache:  store,
			}
			got, err := r.ListAllAddressesAssociation()
			assert.Equal(t, tt.wantErr, err)

			if err == nil {
				// Check that results were cached
				cachedData, err := r.ListAllAddressesAssociation()
				assert.NoError(t, err)
				assert.Equal(t, got, cachedData)
				assert.IsType(t, []*ec2.Address{}, store.Get("ec2ListAllAddressesAssociation"))
			}

			changelog, err := diff.Diff(got, tt.want)
			assert.Nil(t, err)
			if len(changelog) > 0 {
				for _, change := range changelog {
					t.Errorf("%s: %s -> %s", strings.Join(change.Path, "."), change.From, change.To)
				}
				t.Fail()
			}
		})
	}
}

func Test_ec2Repository_ListAllInstances(t *testing.T) {
	tests := []struct {
		name    string
		mocks   func(client *awstest.MockFakeEC2)
		want    []*ec2.Instance
		wantErr error
	}{
		{name: "List with 2 pages",
			mocks: func(client *awstest.MockFakeEC2) {
				client.On("DescribeInstancesPages",
					&ec2.DescribeInstancesInput{
						Filters: []*ec2.Filter{
							{
								Name: aws.String("instance-state-name"),
								Values: aws.StringSlice([]string{
									"pending",
									"running",
									"stopping",
									"shutting-down",
									"stopped",
								}),
							},
						},
					},
					mock.MatchedBy(func(callback func(res *ec2.DescribeInstancesOutput, lastPage bool) bool) bool {
						callback(&ec2.DescribeInstancesOutput{
							Reservations: []*ec2.Reservation{
								{
									Instances: []*ec2.Instance{
										{ImageId: aws.String("1")},
										{ImageId: aws.String("2")},
										{ImageId: aws.String("3")},
									},
								},
								{
									Instances: []*ec2.Instance{
										{ImageId: aws.String("4")},
										{ImageId: aws.String("5")},
										{ImageId: aws.String("6")},
									},
								},
							},
						}, false)
						callback(&ec2.DescribeInstancesOutput{
							Reservations: []*ec2.Reservation{
								{
									Instances: []*ec2.Instance{
										{ImageId: aws.String("7")},
										{ImageId: aws.String("8")},
										{ImageId: aws.String("9")},
									},
								},
								{
									Instances: []*ec2.Instance{
										{ImageId: aws.String("10")},
										{ImageId: aws.String("11")},
										{ImageId: aws.String("12")},
									},
								},
							},
						}, true)
						return true
					})).Return(nil).Once()
			},
			want: []*ec2.Instance{
				{ImageId: aws.String("1")},
				{ImageId: aws.String("2")},
				{ImageId: aws.String("3")},
				{ImageId: aws.String("4")},
				{ImageId: aws.String("5")},
				{ImageId: aws.String("6")},
				{ImageId: aws.String("7")},
				{ImageId: aws.String("8")},
				{ImageId: aws.String("9")},
				{ImageId: aws.String("10")},
				{ImageId: aws.String("11")},
				{ImageId: aws.String("12")},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := cache.New(1)
			client := &awstest.MockFakeEC2{}
			tt.mocks(client)
			r := &ec2Repository{
				client: client,
				cache:  store,
			}
			got, err := r.ListAllInstances()
			assert.Equal(t, tt.wantErr, err)

			if err == nil {
				// Check that results were cached
				cachedData, err := r.ListAllInstances()
				assert.NoError(t, err)
				assert.Equal(t, got, cachedData)
				assert.IsType(t, []*ec2.Instance{}, store.Get("ec2ListAllInstances"))
			}
			changelog, err := diff.Diff(got, tt.want)
			assert.Nil(t, err)
			if len(changelog) > 0 {
				for _, change := range changelog {
					t.Errorf("%s: %s -> %s", strings.Join(change.Path, "."), change.From, change.To)
				}
				t.Fail()
			}
		})
	}
}

func Test_ec2Repository_ListAllKeyPairs(t *testing.T) {
	tests := []struct {
		name    string
		mocks   func(client *awstest.MockFakeEC2)
		want    []*ec2.KeyPairInfo
		wantErr error
	}{
		{
			name: "List address",
			mocks: func(client *awstest.MockFakeEC2) {
				client.On("DescribeKeyPairs", &ec2.DescribeKeyPairsInput{}).
					Return(&ec2.DescribeKeyPairsOutput{
						KeyPairs: []*ec2.KeyPairInfo{
							{KeyPairId: aws.String("1")},
							{KeyPairId: aws.String("2")},
							{KeyPairId: aws.String("3")},
							{KeyPairId: aws.String("4")},
						},
					}, nil).Once()
			},
			want: []*ec2.KeyPairInfo{
				{KeyPairId: aws.String("1")},
				{KeyPairId: aws.String("2")},
				{KeyPairId: aws.String("3")},
				{KeyPairId: aws.String("4")},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := cache.New(1)
			client := &awstest.MockFakeEC2{}
			tt.mocks(client)
			r := &ec2Repository{
				client: client,
				cache:  store,
			}
			got, err := r.ListAllKeyPairs()
			assert.Equal(t, tt.wantErr, err)

			if err == nil {
				// Check that results were cached
				cachedData, err := r.ListAllKeyPairs()
				assert.NoError(t, err)
				assert.Equal(t, got, cachedData)
				assert.IsType(t, []*ec2.KeyPairInfo{}, store.Get("ec2ListAllKeyPairs"))
			}

			changelog, err := diff.Diff(got, tt.want)
			assert.Nil(t, err)
			if len(changelog) > 0 {
				for _, change := range changelog {
					t.Errorf("%s: %s -> %s", strings.Join(change.Path, "."), change.From, change.To)
				}
				t.Fail()
			}
		})
	}
}

func Test_ec2Repository_ListAllInternetGateways(t *testing.T) {
	tests := []struct {
		name    string
		mocks   func(client *awstest.MockFakeEC2)
		want    []*ec2.InternetGateway
		wantErr error
	}{
		{
			name: "List only gateways with multiple pages",
			mocks: func(client *awstest.MockFakeEC2) {
				client.On("DescribeInternetGatewaysPages",
					&ec2.DescribeInternetGatewaysInput{},
					mock.MatchedBy(func(callback func(res *ec2.DescribeInternetGatewaysOutput, lastPage bool) bool) bool {
						callback(&ec2.DescribeInternetGatewaysOutput{
							InternetGateways: []*ec2.InternetGateway{
								{
									InternetGatewayId: aws.String("Internet-0"),
								},
								{
									InternetGatewayId: aws.String("Internet-1"),
								},
							},
						}, false)
						callback(&ec2.DescribeInternetGatewaysOutput{
							InternetGateways: []*ec2.InternetGateway{
								{
									InternetGatewayId: aws.String("Internet-2"),
								},
								{
									InternetGatewayId: aws.String("Internet-3"),
								},
							},
						}, true)
						return true
					})).Return(nil).Once()
			},
			want: []*ec2.InternetGateway{
				{
					InternetGatewayId: aws.String("Internet-0"),
				},
				{
					InternetGatewayId: aws.String("Internet-1"),
				},
				{
					InternetGatewayId: aws.String("Internet-2"),
				},
				{
					InternetGatewayId: aws.String("Internet-3"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := cache.New(1)
			client := &awstest.MockFakeEC2{}
			tt.mocks(client)
			r := &ec2Repository{
				client: client,
				cache:  store,
			}
			got, err := r.ListAllInternetGateways()
			assert.Equal(t, tt.wantErr, err)

			if err == nil {
				// Check that results were cached
				cachedData, err := r.ListAllInternetGateways()
				assert.NoError(t, err)
				assert.Equal(t, got, cachedData)
				assert.IsType(t, []*ec2.InternetGateway{}, store.Get("ec2ListAllInternetGateways"))
			}

			changelog, err := diff.Diff(got, tt.want)
			assert.Nil(t, err)
			if len(changelog) > 0 {
				for _, change := range changelog {
					t.Errorf("%s: %s -> %s", strings.Join(change.Path, "."), change.From, change.To)
				}
				t.Fail()
			}
		})
	}
}

func Test_ec2Repository_ListAllSubnets(t *testing.T) {
	tests := []struct {
		name              string
		mocks             func(client *awstest.MockFakeEC2)
		wantSubnet        []*ec2.Subnet
		wantDefaultSubnet []*ec2.Subnet
		wantErr           error
	}{
		{
			name: "List with 2 pages",
			mocks: func(client *awstest.MockFakeEC2) {
				client.On("DescribeSubnetsPages",
					&ec2.DescribeSubnetsInput{},
					mock.MatchedBy(func(callback func(res *ec2.DescribeSubnetsOutput, lastPage bool) bool) bool {
						callback(&ec2.DescribeSubnetsOutput{
							Subnets: []*ec2.Subnet{
								{
									SubnetId:     aws.String("subnet-0b13f1e0eacf67424"), // subnet2
									DefaultForAz: aws.Bool(false),
								},
								{
									SubnetId:     aws.String("subnet-0c9b78001fe186e22"), // subnet3
									DefaultForAz: aws.Bool(false),
								},
								{
									SubnetId:     aws.String("subnet-05810d3f933925f6d"), // subnet1
									DefaultForAz: aws.Bool(false),
								},
							},
						}, false)
						callback(&ec2.DescribeSubnetsOutput{
							Subnets: []*ec2.Subnet{
								{
									SubnetId:     aws.String("subnet-44fe0c65"), // us-east-1a
									DefaultForAz: aws.Bool(true),
								},
								{
									SubnetId:     aws.String("subnet-65e16628"), // us-east-1b
									DefaultForAz: aws.Bool(true),
								},
								{
									SubnetId:     aws.String("subnet-afa656f0"), // us-east-1c
									DefaultForAz: aws.Bool(true),
								},
							},
						}, true)
						return true
					})).Return(nil).Once()
			},
			wantSubnet: []*ec2.Subnet{
				{
					SubnetId:     aws.String("subnet-0b13f1e0eacf67424"), // subnet2
					DefaultForAz: aws.Bool(false),
				},
				{
					SubnetId:     aws.String("subnet-0c9b78001fe186e22"), // subnet3
					DefaultForAz: aws.Bool(false),
				},
				{
					SubnetId:     aws.String("subnet-05810d3f933925f6d"), // subnet1
					DefaultForAz: aws.Bool(false),
				},
			},
			wantDefaultSubnet: []*ec2.Subnet{
				{
					SubnetId:     aws.String("subnet-44fe0c65"), // us-east-1a
					DefaultForAz: aws.Bool(true),
				},
				{
					SubnetId:     aws.String("subnet-65e16628"), // us-east-1b
					DefaultForAz: aws.Bool(true),
				},
				{
					SubnetId:     aws.String("subnet-afa656f0"), // us-east-1c
					DefaultForAz: aws.Bool(true),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := cache.New(2)
			client := &awstest.MockFakeEC2{}
			tt.mocks(client)
			r := &ec2Repository{
				client: client,
				cache:  store,
			}
			gotSubnet, gotDefaultSubnet, err := r.ListAllSubnets()
			assert.Equal(t, tt.wantErr, err)

			if err == nil {
				// Check that results were cached
				cachedData, cachedDefaultData, err := r.ListAllSubnets()
				assert.NoError(t, err)
				assert.Equal(t, gotSubnet, cachedData)
				assert.Equal(t, gotDefaultSubnet, cachedDefaultData)
				assert.IsType(t, []*ec2.Subnet{}, store.Get("ec2ListAllSubnets"))
				assert.IsType(t, []*ec2.Subnet{}, store.Get("ec2ListAllDefaultSubnets"))
			}

			changelog, err := diff.Diff(gotSubnet, tt.wantSubnet)
			assert.Nil(t, err)
			if len(changelog) > 0 {
				for _, change := range changelog {
					t.Errorf("%s: %s -> %s", strings.Join(change.Path, "."), change.From, change.To)
				}
				t.Fail()
			}
			changelog, err = diff.Diff(gotDefaultSubnet, tt.wantDefaultSubnet)
			assert.Nil(t, err)
			if len(changelog) > 0 {
				for _, change := range changelog {
					t.Errorf("%s: %s -> %s", strings.Join(change.Path, "."), change.From, change.To)
				}
				t.Fail()
			}
		})
	}
}

func Test_ec2Repository_ListAllNatGateways(t *testing.T) {
	tests := []struct {
		name    string
		mocks   func(client *awstest.MockFakeEC2)
		want    []*ec2.NatGateway
		wantErr error
	}{
		{
			name: "List only gateways with multiple pages",
			mocks: func(client *awstest.MockFakeEC2) {
				client.On("DescribeNatGatewaysPages",
					&ec2.DescribeNatGatewaysInput{},
					mock.MatchedBy(func(callback func(res *ec2.DescribeNatGatewaysOutput, lastPage bool) bool) bool {
						callback(&ec2.DescribeNatGatewaysOutput{
							NatGateways: []*ec2.NatGateway{
								{
									NatGatewayId: aws.String("nat-0"),
								},
								{
									NatGatewayId: aws.String("nat-1"),
								},
							},
						}, false)
						callback(&ec2.DescribeNatGatewaysOutput{
							NatGateways: []*ec2.NatGateway{
								{
									NatGatewayId: aws.String("nat-2"),
								},
								{
									NatGatewayId: aws.String("nat-3"),
								},
							},
						}, true)
						return true
					})).Return(nil).Once()
			},
			want: []*ec2.NatGateway{
				{
					NatGatewayId: aws.String("nat-0"),
				},
				{
					NatGatewayId: aws.String("nat-1"),
				},
				{
					NatGatewayId: aws.String("nat-2"),
				},
				{
					NatGatewayId: aws.String("nat-3"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := cache.New(1)
			client := &awstest.MockFakeEC2{}
			tt.mocks(client)
			r := &ec2Repository{
				client: client,
				cache:  store,
			}
			got, err := r.ListAllNatGateways()
			assert.Equal(t, tt.wantErr, err)

			if err == nil {
				// Check that results were cached
				cachedData, err := r.ListAllNatGateways()
				assert.NoError(t, err)
				assert.Equal(t, got, cachedData)
				assert.IsType(t, []*ec2.NatGateway{}, store.Get("ec2ListAllNatGateways"))
			}

			changelog, err := diff.Diff(got, tt.want)
			assert.Nil(t, err)
			if len(changelog) > 0 {
				for _, change := range changelog {
					t.Errorf("%s: %s -> %s", strings.Join(change.Path, "."), change.From, change.To)
				}
				t.Fail()
			}
		})
	}
}

func Test_ec2Repository_ListAllRouteTables(t *testing.T) {
	tests := []struct {
		name    string
		mocks   func(client *awstest.MockFakeEC2)
		want    []*ec2.RouteTable
		wantErr error
	}{
		{
			name: "List only route with multiple pages",
			mocks: func(client *awstest.MockFakeEC2) {
				client.On("DescribeRouteTablesPages",
					&ec2.DescribeRouteTablesInput{},
					mock.MatchedBy(func(callback func(res *ec2.DescribeRouteTablesOutput, lastPage bool) bool) bool {
						callback(&ec2.DescribeRouteTablesOutput{
							RouteTables: []*ec2.RouteTable{
								{
									RouteTableId: aws.String("rtb-096bdfb69309c54c3"), // table1
									Routes: []*ec2.Route{
										{
											DestinationCidrBlock: aws.String("10.0.0.0/16"),
											Origin:               aws.String("CreateRouteTable"), // default route
										},
										{
											DestinationCidrBlock: aws.String("1.1.1.1/32"),
											GatewayId:            aws.String("igw-030e74f73bd67f21b"),
										},
										{
											DestinationIpv6CidrBlock: aws.String("::/0"),
											GatewayId:                aws.String("igw-030e74f73bd67f21b"),
										},
									},
								},
								{
									RouteTableId: aws.String("rtb-0169b0937fd963ddc"), // table2
									Routes: []*ec2.Route{
										{
											DestinationCidrBlock: aws.String("10.0.0.0/16"),
											Origin:               aws.String("CreateRouteTable"), // default route
										},
										{
											DestinationCidrBlock: aws.String("0.0.0.0/0"),
											GatewayId:            aws.String("igw-030e74f73bd67f21b"),
										},
										{
											DestinationIpv6CidrBlock: aws.String("::/0"),
											GatewayId:                aws.String("igw-030e74f73bd67f21b"),
										},
									},
								},
							},
						}, false)
						callback(&ec2.DescribeRouteTablesOutput{
							RouteTables: []*ec2.RouteTable{
								{
									RouteTableId: aws.String("rtb-02780c485f0be93c5"), // default_table
									VpcId:        aws.String("vpc-09fe5abc2309ba49d"),
									Associations: []*ec2.RouteTableAssociation{
										{
											Main: aws.Bool(true),
										},
									},
									Routes: []*ec2.Route{
										{
											DestinationCidrBlock: aws.String("10.0.0.0/16"),
											Origin:               aws.String("CreateRouteTable"), // default route
										},
										{
											DestinationCidrBlock: aws.String("10.1.1.0/24"),
											GatewayId:            aws.String("igw-030e74f73bd67f21b"),
										},
										{
											DestinationCidrBlock: aws.String("10.1.2.0/24"),
											GatewayId:            aws.String("igw-030e74f73bd67f21b"),
										},
									},
								},
								{
									RouteTableId: aws.String(""), // table3
									Routes: []*ec2.Route{
										{
											DestinationCidrBlock: aws.String("10.0.0.0/16"),
											Origin:               aws.String("CreateRouteTable"), // default route
										},
									},
								},
							},
						}, true)
						return true
					})).Return(nil).Once()
			},
			want: []*ec2.RouteTable{
				{
					RouteTableId: aws.String("rtb-096bdfb69309c54c3"), // table1
					Routes: []*ec2.Route{
						{
							DestinationCidrBlock: aws.String("10.0.0.0/16"),
							Origin:               aws.String("CreateRouteTable"), // default route
						},
						{
							DestinationCidrBlock: aws.String("1.1.1.1/32"),
							GatewayId:            aws.String("igw-030e74f73bd67f21b"),
						},
						{
							DestinationIpv6CidrBlock: aws.String("::/0"),
							GatewayId:                aws.String("igw-030e74f73bd67f21b"),
						},
					},
				},
				{
					RouteTableId: aws.String("rtb-0169b0937fd963ddc"), // table2
					Routes: []*ec2.Route{
						{
							DestinationCidrBlock: aws.String("10.0.0.0/16"),
							Origin:               aws.String("CreateRouteTable"), // default route
						},
						{
							DestinationCidrBlock: aws.String("0.0.0.0/0"),
							GatewayId:            aws.String("igw-030e74f73bd67f21b"),
						},
						{
							DestinationIpv6CidrBlock: aws.String("::/0"),
							GatewayId:                aws.String("igw-030e74f73bd67f21b"),
						},
					},
				},
				{
					RouteTableId: aws.String("rtb-02780c485f0be93c5"), // default_table
					VpcId:        aws.String("vpc-09fe5abc2309ba49d"),
					Associations: []*ec2.RouteTableAssociation{
						{
							Main: aws.Bool(true),
						},
					},
					Routes: []*ec2.Route{
						{
							DestinationCidrBlock: aws.String("10.0.0.0/16"),
							Origin:               aws.String("CreateRouteTable"), // default route
						},
						{
							DestinationCidrBlock: aws.String("10.1.1.0/24"),
							GatewayId:            aws.String("igw-030e74f73bd67f21b"),
						},
						{
							DestinationCidrBlock: aws.String("10.1.2.0/24"),
							GatewayId:            aws.String("igw-030e74f73bd67f21b"),
						},
					},
				},
				{
					RouteTableId: aws.String(""), // table3
					Routes: []*ec2.Route{
						{
							DestinationCidrBlock: aws.String("10.0.0.0/16"),
							Origin:               aws.String("CreateRouteTable"), // default route
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := cache.New(1)
			client := &awstest.MockFakeEC2{}
			tt.mocks(client)
			r := &ec2Repository{
				client: client,
				cache:  store,
			}
			got, err := r.ListAllRouteTables()
			assert.Equal(t, tt.wantErr, err)

			if err == nil {
				// Check that results were cached
				cachedData, err := r.ListAllRouteTables()
				assert.NoError(t, err)
				assert.Equal(t, got, cachedData)
				assert.IsType(t, []*ec2.RouteTable{}, store.Get("ec2ListAllRouteTables"))
			}

			changelog, err := diff.Diff(got, tt.want)
			assert.Nil(t, err)
			if len(changelog) > 0 {
				for _, change := range changelog {
					t.Errorf("%s: %s -> %s", strings.Join(change.Path, "."), change.From, change.To)
				}
				t.Fail()
			}
		})
	}
}

func Test_ec2Repository_ListAllVPCs(t *testing.T) {
	tests := []struct {
		name           string
		mocks          func(client *awstest.MockFakeEC2)
		wantVPC        []*ec2.Vpc
		wantDefaultVPC []*ec2.Vpc
		wantErr        error
	}{
		{
			name: "mixed default VPC and VPC",
			mocks: func(client *awstest.MockFakeEC2) {
				client.On("DescribeVpcsPages",
					&ec2.DescribeVpcsInput{},
					mock.MatchedBy(func(callback func(res *ec2.DescribeVpcsOutput, lastPage bool) bool) bool {
						callback(&ec2.DescribeVpcsOutput{
							Vpcs: []*ec2.Vpc{
								{
									VpcId:     aws.String("vpc-a8c5d4c1"),
									IsDefault: aws.Bool(true),
								},
								{
									VpcId: aws.String("vpc-0768e1fd0029e3fc3"),
								},
								{
									VpcId:     aws.String("vpc-020b072316a95b97f"),
									IsDefault: aws.Bool(false),
								},
							},
						}, false)
						callback(&ec2.DescribeVpcsOutput{
							Vpcs: []*ec2.Vpc{
								{
									VpcId:     aws.String("vpc-02c50896b59598761"),
									IsDefault: aws.Bool(false),
								},
							},
						}, true)
						return true
					})).Return(nil).Once()
			},
			wantVPC: []*ec2.Vpc{
				{
					VpcId: aws.String("vpc-0768e1fd0029e3fc3"),
				},
				{
					VpcId:     aws.String("vpc-020b072316a95b97f"),
					IsDefault: aws.Bool(false),
				},
				{
					VpcId:     aws.String("vpc-02c50896b59598761"),
					IsDefault: aws.Bool(false),
				},
			},
			wantDefaultVPC: []*ec2.Vpc{
				{
					VpcId:     aws.String("vpc-a8c5d4c1"),
					IsDefault: aws.Bool(true),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := cache.New(2)
			client := &awstest.MockFakeEC2{}
			tt.mocks(client)
			r := &ec2Repository{
				client: client,
				cache:  store,
			}
			gotVPCs, gotDefaultVPCs, err := r.ListAllVPCs()
			assert.Equal(t, tt.wantErr, err)

			if err == nil {
				// Check that results were cached
				cachedData, cachedDefaultData, err := r.ListAllVPCs()
				assert.NoError(t, err)
				assert.Equal(t, gotVPCs, cachedData)
				assert.Equal(t, gotDefaultVPCs, cachedDefaultData)
				assert.IsType(t, []*ec2.Vpc{}, store.Get("ec2ListAllVPCs"))
				assert.IsType(t, []*ec2.Vpc{}, store.Get("ec2ListAllDefaultVPCs"))
			}

			changelog, err := diff.Diff(gotVPCs, tt.wantVPC)
			assert.Nil(t, err)
			if len(changelog) > 0 {
				for _, change := range changelog {
					t.Errorf("%s: %s -> %s", strings.Join(change.Path, "."), change.From, change.To)
				}
				t.Fail()
			}
			changelog, err = diff.Diff(gotDefaultVPCs, tt.wantDefaultVPC)
			assert.Nil(t, err)
			if len(changelog) > 0 {
				for _, change := range changelog {
					t.Errorf("%s: %s -> %s", strings.Join(change.Path, "."), change.From, change.To)
				}
				t.Fail()
			}
		})
	}
}

func Test_ec2Repository_ListAllSecurityGroups(t *testing.T) {
	tests := []struct {
		name                     string
		mocks                    func(client *awstest.MockFakeEC2)
		wantSecurityGroup        []*ec2.SecurityGroup
		wantDefaultSecurityGroup []*ec2.SecurityGroup
		wantErr                  error
	}{
		{
			name: "List with 1 pages",
			mocks: func(client *awstest.MockFakeEC2) {
				client.On("DescribeSecurityGroupsPages",
					&ec2.DescribeSecurityGroupsInput{},
					mock.MatchedBy(func(callback func(res *ec2.DescribeSecurityGroupsOutput, lastPage bool) bool) bool {
						callback(&ec2.DescribeSecurityGroupsOutput{
							SecurityGroups: []*ec2.SecurityGroup{
								{
									GroupId:   aws.String("sg-0254c038e32f25530"),
									GroupName: aws.String("foo"),
								},
								{
									GroupId:   aws.String("sg-9e0204ff"),
									GroupName: aws.String("default"),
								},
							},
						}, true)
						return true
					})).Return(nil).Once()
			},
			wantSecurityGroup: []*ec2.SecurityGroup{
				{
					GroupId:   aws.String("sg-0254c038e32f25530"),
					GroupName: aws.String("foo"),
				},
			},
			wantDefaultSecurityGroup: []*ec2.SecurityGroup{
				{
					GroupId:   aws.String("sg-9e0204ff"),
					GroupName: aws.String("default"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := cache.New(2)
			client := &awstest.MockFakeEC2{}
			tt.mocks(client)
			r := &ec2Repository{
				client: client,
				cache:  store,
			}
			gotSecurityGroups, gotDefaultSecurityGroups, err := r.ListAllSecurityGroups()
			assert.Equal(t, tt.wantErr, err)

			if err == nil {
				// Check that results were cached
				cachedData, cachedDefaultData, err := r.ListAllSecurityGroups()
				assert.NoError(t, err)
				assert.Equal(t, gotSecurityGroups, cachedData)
				assert.Equal(t, gotDefaultSecurityGroups, cachedDefaultData)
				assert.IsType(t, []*ec2.SecurityGroup{}, store.Get("ec2ListAllSecurityGroups"))
				assert.IsType(t, []*ec2.SecurityGroup{}, store.Get("ec2ListAllDefaultSecurityGroups"))
			}

			changelog, err := diff.Diff(gotSecurityGroups, tt.wantSecurityGroup)
			assert.Nil(t, err)
			if len(changelog) > 0 {
				for _, change := range changelog {
					t.Errorf("%s: %s -> %s", strings.Join(change.Path, "."), change.From, change.To)
				}
				t.Fail()
			}
			changelog, err = diff.Diff(gotDefaultSecurityGroups, tt.wantDefaultSecurityGroup)
			assert.Nil(t, err)
			if len(changelog) > 0 {
				for _, change := range changelog {
					t.Errorf("%s: %s -> %s", strings.Join(change.Path, "."), change.From, change.To)
				}
				t.Fail()
			}
		})
	}
}

func Test_ec2Repository_ListAllNetworkACLs(t *testing.T) {

	testErr := errors.New("test")

	tests := []struct {
		name    string
		mocks   func(client *awstest.MockFakeEC2)
		want    []*ec2.NetworkAcl
		wantErr error
	}{
		{
			name: "List with 1 pages",
			mocks: func(client *awstest.MockFakeEC2) {
				client.On("DescribeNetworkAclsPages",
					&ec2.DescribeNetworkAclsInput{},
					mock.MatchedBy(func(callback func(res *ec2.DescribeNetworkAclsOutput, lastPage bool) bool) bool {
						callback(&ec2.DescribeNetworkAclsOutput{
							NetworkAcls: []*ec2.NetworkAcl{
								{
									NetworkAclId: aws.String("id1"),
								},
								{
									NetworkAclId: aws.String("id2"),
								},
							},
						}, true)
						return true
					})).Return(nil).Once()
			},
			want: []*ec2.NetworkAcl{
				{
					NetworkAclId: aws.String("id1"),
				},
				{
					NetworkAclId: aws.String("id2"),
				},
			},
		},
		{
			name: "List return error",
			mocks: func(client *awstest.MockFakeEC2) {
				client.On("DescribeNetworkAclsPages",
					&ec2.DescribeNetworkAclsInput{},
					mock.Anything,
				).Return(testErr)
			},
			wantErr: testErr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := cache.New(2)
			client := &awstest.MockFakeEC2{}
			tt.mocks(client)
			r := &ec2Repository{
				client: client,
				cache:  store,
			}
			got, err := r.ListAllNetworkACLs()
			assert.Equal(t, tt.wantErr, err)

			if err == nil {
				// Check that results were cached
				cachedData, err := r.ListAllNetworkACLs()
				assert.NoError(t, err)
				assert.Equal(t, got, cachedData)
				assert.IsType(t, []*ec2.NetworkAcl{}, store.Get("ec2ListAllNetworkACLs"))
			}

			changelog, err := diff.Diff(got, tt.want)
			assert.Nil(t, err)
			if len(changelog) > 0 {
				for _, change := range changelog {
					t.Errorf("%s: %s -> %s", strings.Join(change.Path, "."), change.From, change.To)
				}
				t.Fail()
			}
			client.AssertExpectations(t)
		})
	}
}

func Test_ec2Repository_DescribeLaunchTemplates(t *testing.T) {

	testErr := errors.New("test")

	tests := []struct {
		name    string
		mocks   func(client *awstest.MockFakeEC2)
		want    []*ec2.LaunchTemplate
		wantErr error
	}{
		{
			name: "List with 1 pages",
			mocks: func(client *awstest.MockFakeEC2) {
				client.On("DescribeLaunchTemplates",
					&ec2.DescribeLaunchTemplatesInput{},
				).Return(&ec2.DescribeLaunchTemplatesOutput{
					LaunchTemplates: []*ec2.LaunchTemplate{
						{
							LaunchTemplateId: aws.String("id1"),
						},
						{
							LaunchTemplateId: aws.String("id2"),
						},
					},
				}, nil).Once()
			},
			want: []*ec2.LaunchTemplate{
				{
					LaunchTemplateId: aws.String("id1"),
				},
				{
					LaunchTemplateId: aws.String("id2"),
				},
			},
		},
		{
			name: "List return error",
			mocks: func(client *awstest.MockFakeEC2) {
				client.On("DescribeLaunchTemplates",
					&ec2.DescribeLaunchTemplatesInput{},
				).Return(nil, testErr)
			},
			wantErr: testErr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := cache.New(1)
			client := &awstest.MockFakeEC2{}
			tt.mocks(client)
			r := &ec2Repository{
				client: client,
				cache:  store,
			}
			got, err := r.DescribeLaunchTemplates()
			assert.Equal(t, tt.wantErr, err)

			if err == nil {
				// Check that results were cached
				cachedData, err := r.DescribeLaunchTemplates()
				assert.NoError(t, err)
				assert.Equal(t, got, cachedData)
				assert.IsType(t, []*ec2.LaunchTemplate{}, store.Get("DescribeLaunchTemplates"))
			}

			changelog, err := diff.Diff(got, tt.want)
			assert.Nil(t, err)
			if len(changelog) > 0 {
				for _, change := range changelog {
					t.Errorf("%s: %s -> %s", strings.Join(change.Path, "."), change.From, change.To)
				}
				t.Fail()
			}
			client.AssertExpectations(t)
		})
	}
}

func Test_ec2Repository_IsEbsEncryptionEnabledByDefault(t *testing.T) {

	testErr := errors.New("test")

	tests := []struct {
		name    string
		mocks   func(client *awstest.MockFakeEC2, store *cache.MockCache)
		want    bool
		wantErr error
	}{
		{
			name: "test that encryption enabled by default",
			mocks: func(client *awstest.MockFakeEC2, store *cache.MockCache) {
				store.On("Get", "ec2IsEbsEncryptionEnabledByDefault").
					Return(nil).
					Once()

				client.On("GetEbsEncryptionByDefault",
					&ec2.GetEbsEncryptionByDefaultInput{},
				).Return(&ec2.GetEbsEncryptionByDefaultOutput{
					EbsEncryptionByDefault: aws.Bool(true),
				}, nil).Once()

				store.On("Put", "ec2IsEbsEncryptionEnabledByDefault", true).
					Return(false).
					Once()
			},
			want: true,
		},
		{
			name: "test that encryption enabled by default (cached)",
			mocks: func(client *awstest.MockFakeEC2, store *cache.MockCache) {
				store.On("Get", "ec2IsEbsEncryptionEnabledByDefault").
					Return(false).
					Once()
			},
			want: false,
		},
		{
			name: "error while getting default encryption value",
			mocks: func(client *awstest.MockFakeEC2, store *cache.MockCache) {
				store.On("Get", "ec2IsEbsEncryptionEnabledByDefault").
					Return(nil).
					Once()

				client.On("GetEbsEncryptionByDefault",
					&ec2.GetEbsEncryptionByDefaultInput{},
				).Return(nil, testErr).Once()
			},
			wantErr: testErr,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &cache.MockCache{}
			client := &awstest.MockFakeEC2{}
			tt.mocks(client, store)
			r := &ec2Repository{
				client: client,
				cache:  store,
			}
			got, err := r.IsEbsEncryptionEnabledByDefault()

			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)

			client.AssertExpectations(t)
			store.AssertExpectations(t)
		})
	}
}
