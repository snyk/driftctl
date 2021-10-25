package repository

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/privatedns/armprivatedns"
	"github.com/cloudskiff/driftctl/pkg/remote/azurerm/common"
	"github.com/cloudskiff/driftctl/pkg/remote/cache"
)

type PrivateDNSRepository interface {
	ListAllPrivateZones() ([]*armprivatedns.PrivateZone, error)
}

type privateDNSZoneListPager interface {
	pager
	PageResponse() armprivatedns.PrivateZonesListResponse
}

type privateZonesClient interface {
	List(options *armprivatedns.PrivateZonesListOptions) privateDNSZoneListPager
}

type privateZonesClientImpl struct {
	client *armprivatedns.PrivateZonesClient
}

func (c *privateZonesClientImpl) List(options *armprivatedns.PrivateZonesListOptions) privateDNSZoneListPager {
	return c.client.List(options)
}

type privateDNSRepository struct {
	zoneClient privateZonesClient
	cache      cache.Cache
}

func NewPrivateDNSRepository(con *arm.Connection, config common.AzureProviderConfig, cache cache.Cache) *privateDNSRepository {
	return &privateDNSRepository{
		&privateZonesClientImpl{armprivatedns.NewPrivateZonesClient(con, config.SubscriptionID)},
		cache,
	}
}

func (s *privateDNSRepository) ListAllPrivateZones() ([]*armprivatedns.PrivateZone, error) {
	cacheKey := "privateDNSListAllPrivateZones"
	if v := s.cache.Get(cacheKey); v != nil {
		return v.([]*armprivatedns.PrivateZone), nil
	}

	pager := s.zoneClient.List(nil)
	results := make([]*armprivatedns.PrivateZone, 0)
	for pager.NextPage(context.Background()) {
		resp := pager.PageResponse()
		if err := pager.Err(); err != nil {
			return nil, err
		}
		results = append(results, resp.Value...)
	}

	if err := pager.Err(); err != nil {
		return nil, err
	}

	s.cache.Put(cacheKey, results)

	return results, nil
}
