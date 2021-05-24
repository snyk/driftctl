package cache

import (
	"fmt"
	"testing"

	"github.com/cloudskiff/driftctl/pkg/resource/aws"
	"github.com/stretchr/testify/assert"
)

func BenchmarkCache(b *testing.B) {
	cache := New()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("test-key-%d", i)
		data := make([]*aws.AwsLambdaFunction, 1000)
		assert.Equal(b, false, cache.Set(key, data))
		assert.Equal(b, data, cache.Get(key))
	}
}

func TestCache(t *testing.T) {
	t.Run("should return nil on non-existing key", func(t *testing.T) {
		cache := New()
		assert.Equal(t, nil, cache.Get("test"))
	})

	t.Run("should retrieve newly added key", func(t *testing.T) {
		cache := New()
		assert.Equal(t, false, cache.Set("s3", []string{}))
		assert.Equal(t, []string{}, cache.Get("s3"))
	})

	t.Run("should override existing key", func(t *testing.T) {
		cache := New()
		assert.Equal(t, false, cache.Set("s3", []string{}))
		assert.Equal(t, []string{}, cache.Get("s3"))

		assert.Equal(t, true, cache.Set("s3", []string{"test"}))
		assert.Equal(t, []string{"test"}, cache.Get("s3"))
	})
}
