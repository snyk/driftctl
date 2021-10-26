package repository

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/privatedns/armprivatedns"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/cloudskiff/driftctl/pkg/remote/azurerm/common"
	"github.com/cloudskiff/driftctl/pkg/remote/cache"
)

type PrivateDNSRepository interface {
	ListAllPrivateZones() ([]*armprivatedns.PrivateZone, error)
	ListAllARecords(zone *armprivatedns.PrivateZone) ([]*armprivatedns.RecordSet, error)
	ListAllAAAARecords(zone *armprivatedns.PrivateZone) ([]*armprivatedns.RecordSet, error)
}

type privateDNSZoneListPager interface {
	pager
	PageResponse() armprivatedns.PrivateZonesListResponse
}

type privateDNSRecordSetListPager interface {
	pager
	PageResponse() armprivatedns.RecordSetsListResponse
}

type privateRecordSetClient interface {
	List(resourceGroupName string, privateZoneName string, options *armprivatedns.RecordSetsListOptions) privateDNSRecordSetListPager
}

type privateRecordSetClientImpl struct {
	client *armprivatedns.RecordSetsClient
}

func (c *privateRecordSetClientImpl) List(resourceGroupName string, privateZoneName string, options *armprivatedns.RecordSetsListOptions) privateDNSRecordSetListPager {
	return c.client.List(resourceGroupName, privateZoneName, options)
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
	zoneClient   privateZonesClient
	recordClient privateRecordSetClient
	cache        cache.Cache
}

func NewPrivateDNSRepository(con *arm.Connection, config common.AzureProviderConfig, cache cache.Cache) *privateDNSRepository {
	return &privateDNSRepository{
		&privateZonesClientImpl{armprivatedns.NewPrivateZonesClient(con, config.SubscriptionID)},
		&privateRecordSetClientImpl{armprivatedns.NewRecordSetsClient(con, config.SubscriptionID)},
		cache,
	}
}

func (s *privateDNSRepository) listAllRecords(zone *armprivatedns.PrivateZone) ([]*armprivatedns.RecordSet, error) {
	cacheKey := fmt.Sprintf("listAllRecords-%s", *zone.ID)
	v := s.cache.GetAndLock(cacheKey)
	defer s.cache.Unlock(cacheKey)
	if v != nil {
		return v.([]*armprivatedns.RecordSet), nil
	}

	res, err := azure.ParseResourceID(*zone.ID)
	if err != nil {
		return nil, err
	}

	pager := s.recordClient.List(res.ResourceGroup, *zone.Name, nil)
	results := make([]*armprivatedns.RecordSet, 0)
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

func (s *privateDNSRepository) ListAllARecords(zone *armprivatedns.PrivateZone) ([]*armprivatedns.RecordSet, error) {
	cacheKey := fmt.Sprintf("privateDNSListAllARecords-%s", *zone.ID)
	if v := s.cache.Get(cacheKey); v != nil {
		return v.([]*armprivatedns.RecordSet), nil
	}

	records, err := s.listAllRecords(zone)
	if err != nil {
		return nil, err
	}
	results := make([]*armprivatedns.RecordSet, 0)
	for _, record := range records {
		if record.Properties.ARecords == nil {
			continue
		}
		results = append(results, record)

	}

	return results, nil
}

func (s *privateDNSRepository) ListAllAAAARecords(zone *armprivatedns.PrivateZone) ([]*armprivatedns.RecordSet, error) {
	cacheKey := fmt.Sprintf("ListAllAAAARecords-%s", *zone.ID)
	if v := s.cache.Get(cacheKey); v != nil {
		return v.([]*armprivatedns.RecordSet), nil
	}

	records, err := s.listAllRecords(zone)
	if err != nil {
		return nil, err
	}
	results := make([]*armprivatedns.RecordSet, 0)
	for _, record := range records {
		if record.Properties.AaaaRecords == nil {
			continue
		}
		results = append(results, record)

	}

	return results, nil
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
