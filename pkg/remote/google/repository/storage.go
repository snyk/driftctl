package repository

import (
	"context"
	"fmt"
	"sync"

	"cloud.google.com/go/storage"
	"github.com/cloudskiff/driftctl/pkg/remote/cache"
)

type StorageRepository interface {
	ListAllBindings(bucketName string) (map[string][]string, error)
}

type storageRepository struct {
	client *storage.Client
	cache  cache.Cache
	lock   sync.Locker
}

func NewStorageRepository(client *storage.Client, cache cache.Cache) *storageRepository {
	return &storageRepository{
		client: client,
		cache:  cache,
		lock:   &sync.Mutex{},
	}
}

func (s storageRepository) ListAllBindings(bucketName string) (map[string][]string, error) {

	s.lock.Lock()
	defer s.lock.Unlock()
	if cachedResults := s.cache.Get(fmt.Sprintf("%s-%s", "ListAllBindings", bucketName)); cachedResults != nil {
		return cachedResults.(map[string][]string), nil
	}

	bucket := s.client.Bucket(bucketName)
	policy, err := bucket.IAM().Policy(context.Background())
	if err != nil {
		return nil, err
	}
	bindings := make(map[string][]string)
	for _, name := range policy.Roles() {
		members := policy.Members(name)
		bindings[string(name)] = members
	}

	s.cache.Put("ListAllBindings", bindings)

	return bindings, nil
}
