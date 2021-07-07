package repository

import (
	"fmt"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/kms/kmsiface"
	"github.com/cloudskiff/driftctl/pkg/remote/cache"
	"github.com/sirupsen/logrus"
)

type KMSRepository interface {
	ListAllKeys() ([]*kms.KeyListEntry, error)
	ListAllAliases() ([]*kms.AliasListEntry, error)
}

type kmsRepository struct {
	client          kmsiface.KMSAPI
	cache           cache.Cache
	describeKeyLock *sync.Mutex
}

func NewKMSRepository(session *session.Session, c cache.Cache) *kmsRepository {
	return &kmsRepository{
		kms.New(session),
		c,
		&sync.Mutex{},
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

	result, err := r.filterAliases(aliases)
	if err != nil {
		return nil, err
	}
	r.cache.Put("kmsListAllAliases", result)
	return result, nil
}

func (r *kmsRepository) describeKey(keyId *string) (*kms.DescribeKeyOutput, error) {
	var results interface{}
	// Since this method can be call in parallel, we should lock and unlock if we want to be sure to hit the cache
	r.describeKeyLock.Lock()
	defer r.describeKeyLock.Unlock()
	cacheKey := fmt.Sprintf("kmsDescribeKey-%s", *keyId)
	results = r.cache.Get(cacheKey)
	if results == nil {
		var err error
		results, err = r.client.DescribeKey(&kms.DescribeKeyInput{KeyId: keyId})
		if err != nil {
			return nil, err
		}
		r.cache.Put(cacheKey, results)
	}
	describeKey := results.(*kms.DescribeKeyOutput)
	if aws.StringValue(describeKey.KeyMetadata.KeyState) == kms.KeyStatePendingDeletion {
		return nil, nil
	}
	return describeKey, nil
}

func (r *kmsRepository) filterKeys(keys []*kms.KeyListEntry) ([]*kms.KeyListEntry, error) {
	var customerKeys []*kms.KeyListEntry
	for _, key := range keys {
		k, err := r.describeKey(key.KeyId)
		if err != nil {
			return nil, err
		}
		if k == nil {
			logrus.WithFields(logrus.Fields{
				"id": *key.KeyId,
			}).Debug("Ignored kms key from listing since it is pending from deletion")
			continue
		}
		if k.KeyMetadata.KeyManager != nil && *k.KeyMetadata.KeyManager != "AWS" {
			customerKeys = append(customerKeys, key)
		}
	}
	return customerKeys, nil
}

func (r *kmsRepository) filterAliases(aliases []*kms.AliasListEntry) ([]*kms.AliasListEntry, error) {
	var customerAliases []*kms.AliasListEntry
	for _, alias := range aliases {
		if alias.AliasName != nil && !strings.HasPrefix(*alias.AliasName, "alias/aws/") {
			k, err := r.describeKey(alias.TargetKeyId)
			if err != nil {
				return nil, err
			}
			if k == nil {
				logrus.WithFields(logrus.Fields{
					"id":    *alias.TargetKeyId,
					"alias": *alias.AliasName,
				}).Debug("Ignored kms key alias from listing since it is linked to a pending from deletion key")
				continue
			}
			customerAliases = append(customerAliases, alias)
		}
	}
	return customerAliases, nil
}
