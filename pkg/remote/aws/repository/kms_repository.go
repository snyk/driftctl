package repository

import (
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/kms/kmsiface"
	"github.com/cloudskiff/driftctl/pkg/remote/cache"
)

type KMSRepository interface {
	ListAllKeys() ([]*kms.KeyListEntry, error)
	ListAllAliases() ([]*kms.AliasListEntry, error)
}

type kmsRepository struct {
	client kmsiface.KMSAPI
	cache  cache.Cache
}

func NewKMSRepository(session *session.Session, c cache.Cache) *kmsRepository {
	return &kmsRepository{
		kms.New(session),
		c,
	}
}

func (r *kmsRepository) ListAllKeys() ([]*kms.KeyListEntry, error) {
	if v := r.cache.Get("kmsListAllKeys"); v != nil {
		return v.([]*kms.KeyListEntry), nil
	}

	var keys []*kms.KeyListEntry
	input := kms.ListKeysInput{}
	err := r.client.ListKeysPages(&input,
		func(resp *kms.ListKeysOutput, lastPage bool) bool {
			keys = append(keys, resp.Keys...)
			return !lastPage
		},
	)
	if err != nil {
		return nil, err
	}
	customerKeys, err := r.filterKeys(keys)
	if err != nil {
		return nil, err
	}

	r.cache.Put("kmsListAllKeys", customerKeys)
	return customerKeys, nil
}

func (r *kmsRepository) ListAllAliases() ([]*kms.AliasListEntry, error) {
	if v := r.cache.Get("kmsListAllAliases"); v != nil {
		return v.([]*kms.AliasListEntry), nil
	}

	var aliases []*kms.AliasListEntry
	input := kms.ListAliasesInput{}
	err := r.client.ListAliasesPages(&input,
		func(resp *kms.ListAliasesOutput, lastPage bool) bool {
			aliases = append(aliases, resp.Aliases...)
			return !lastPage
		},
	)
	if err != nil {
		return nil, err
	}

	result := r.filterAliases(aliases)
	r.cache.Put("kmsListAllAliases", result)
	return result, nil
}

func (r *kmsRepository) filterKeys(keys []*kms.KeyListEntry) ([]*kms.KeyListEntry, error) {
	var customerKeys []*kms.KeyListEntry
	for _, key := range keys {
		k, err := r.client.DescribeKey(&kms.DescribeKeyInput{
			KeyId: key.KeyId,
		})
		if err != nil {
			return nil, err
		}
		if k.KeyMetadata.KeyManager != nil && *k.KeyMetadata.KeyManager != "AWS" {
			customerKeys = append(customerKeys, key)
		}
	}
	return customerKeys, nil
}

func (r *kmsRepository) filterAliases(aliases []*kms.AliasListEntry) []*kms.AliasListEntry {
	var customerAliases []*kms.AliasListEntry
	for _, alias := range aliases {
		if alias.AliasName != nil && !strings.HasPrefix(*alias.AliasName, "alias/aws/") {
			customerAliases = append(customerAliases, alias)
		}
	}
	return customerAliases
}
