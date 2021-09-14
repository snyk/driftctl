package repository

import (
	"context"
	"sync"

	"cloud.google.com/go/storage"
	"github.com/cloudskiff/driftctl/pkg/remote/cache"
)

type StorageRepository interface {
	ListAllBindings(bucketName string) map[string][]string
}

type storageRepository struct {
	client *storage.Client
	cache  cache.Cache
	lock   sync.Locker
}

func NewStorageRepository(cache cache.Cache) *storageRepository {
	client, err := storage.NewClient(context.Background())
	if err != nil {
		panic(err)
	}

	return &storageRepository{
		client: client,
		cache:  cache,
		lock:   &sync.Mutex{},
	}
}

func (s storageRepository) ListAllBindings(bucketName string) map[string][]string {

	s.lock.Lock()
	defer s.lock.Unlock()
	if cachedResults := s.cache.Get("ListAllBindings"); cachedResults != nil {
		return cachedResults.(map[string][]string)
	}

	bucket := s.client.Bucket(bucketName)
	policy, err := bucket.IAM().Policy(context.Background())
	if err != nil {
		panic(err)
	}
	bindings := make(map[string][]string)
	for _, name := range policy.Roles() {
		members := policy.Members(name)
		bindings[string(name)] = members
	}

	s.cache.Put("ListAllBindings", bindings)

	return bindings
}
