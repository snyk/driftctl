package repository

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/kms/kmsiface"
)

type KMSRepository interface {
	ListAllKeys() ([]*kms.KeyListEntry, error)
}

type kmsRepository struct {
	client kmsiface.KMSAPI
}

func NewKMSRepository(session *session.Session) *kmsRepository {
	return &kmsRepository{
		kms.New(session),
	}
}

func (r *kmsRepository) ListAllKeys() ([]*kms.KeyListEntry, error) {
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
	return customerKeys, nil
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
