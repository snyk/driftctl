package cache

import (
	"fmt"
	"testing"

	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/stretchr/testify/assert"
)

func BenchmarkCache(b *testing.B) {
	cache := New(2048)
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("test-key-%d", i)
		data := make([]*aws.AwsLambdaFunction, 1024)
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
		cache := New(5)

		for i := 0; i <= 10; i++ {
			cache.Put(fmt.Sprintf("test-%d", i), "value")
		}

		assert.Equal(t, "value", cache.Get("test-10"))
		assert.Equal(t, 5, cache.Len())
	})
}
