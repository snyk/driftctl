package cache

import (
	"fmt"
	"testing"

	"github.com/cloudskiff/driftctl/pkg/resource"
	"github.com/stretchr/testify/assert"
)

func BenchmarkCache(b *testing.B) {
	cache := New(500)
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("test-key-%d", i)
		data := make([]*resource.Resource, 1024)
		assert.Equal(b, false, cache.Put(key, data))
		assert.Equal(b, data, cache.Get(key))
	}
}

func TestCache(t *testing.T) {
	t.Run("should return nil on non-existing key", func(t *testing.T) {
		cache := New(5)
		assert.Equal(t, nil, cache.Get("test"))
		assert.Equal(t, 0, cache.Len())
	})

	t.Run("should retrieve newly added key", func(t *testing.T) {
		cache := New(5)
		assert.Equal(t, false, cache.Put("s3", []string{}))
		assert.Equal(t, []string{}, cache.Get("s3"))
		assert.Equal(t, 1, cache.Len())
	})

	t.Run("should override existing key", func(t *testing.T) {
		cache := New(5)
		assert.Equal(t, false, cache.Put("s3", []string{}))
		assert.Equal(t, []string{}, cache.Get("s3"))

		assert.Equal(t, true, cache.Put("s3", []string{"test"}))
		assert.Equal(t, []string{"test"}, cache.Get("s3"))
		assert.Equal(t, 1, cache.Len())
	})

	t.Run("should delete the least used keys", func(t *testing.T) {
		keys := []struct {
			key   string
			value interface{}
		}{
			{key: "test-0", value: nil},
			{key: "test-1", value: nil},
			{key: "test-2", value: nil},
			{key: "test-3", value: nil},
			{key: "test-4", value: nil},
			{key: "test-5", value: nil},
			{key: "test-6", value: "value"},
			{key: "test-7", value: "value"},
			{key: "test-8", value: "value"},
			{key: "test-9", value: "value"},
			{key: "test-10", value: "value"},
		}

		cache := New(5)
		for i := 0; i <= 10; i++ {
			cache.Put(fmt.Sprintf("test-%d", i), "value")
		}
		for _, k := range keys {
			assert.Equal(t, k.value, cache.Get(k.key))
		}
		assert.Equal(t, 5, cache.Len())
	})

	t.Run("should ignore keys when capacity is 0", func(t *testing.T) {
		keys := []struct {
			key   string
			value interface{}
		}{
			{
				"test",
				[]string{"slice"},
			},
			{
				"test",
				[]string{},
			},
			{
				"test2",
				[]*resource.Resource{},
			},
		}
		cache := New(0)

		for _, k := range keys {
			assert.Equal(t, false, cache.Put(k.key, k.value))
			assert.Equal(t, nil, cache.Get(k.key))
		}
		assert.Equal(t, 0, cache.Len())
	})
}
