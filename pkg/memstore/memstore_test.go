package memstore

import (
	"sync"
	"testing"

	"github.com/cloudskiff/driftctl/test/resource"
	"github.com/stretchr/testify/assert"
)

func TestStore(t *testing.T) {
	cases := []struct {
		name         string
		bucket       string
		values       map[string]interface{}
		expectedJSON string
	}{
		{
			name:   "test basic store usage",
			bucket: "test-bucket-1",
			values: map[string]interface{}{
				"test-value_|)": 13,
				"duration_key":  "23",
				"null":          nil,
				"res":           &resource.FakeResource{Id: "id", Type: "type"},
			},
			expectedJSON: `{"duration_key":"23","null":null,"res":{"Id":"id","Type":"type","Attrs":null},"test-value_|)":13}`,
		},
		{
			name:         "test empty bucket",
			bucket:       "test-bucket-empty",
			values:       map[string]interface{}{},
			expectedJSON: `{}`,
		},
		{
			name:   "test bucket with nil values",
			bucket: "test-bucket-empty",
			values: map[string]interface{}{
				"version":         nil,
				"total_resources": nil,
				"total_managed":   nil,
			},
			expectedJSON: `{"total_managed":null,"total_resources":null,"version":null}`,
		},
	}

	for _, tt := range cases {
		kv := New()

		t.Run(tt.name, func(t *testing.T) {
			var wg sync.WaitGroup

			for key, val := range tt.values {
				wg.Add(1)
				go func(key string, val interface{}, wg *sync.WaitGroup) {
					defer wg.Done()
					kv.Bucket(tt.bucket).Set(key, val)
					assert.Equal(t, val, kv.Bucket(tt.bucket).Get(key))
					assert.Equal(t, nil, kv.Bucket("dummybucketname").Get(key))
				}(key, val, &wg)
			}

			wg.Wait()

			b, err := kv.Bucket(tt.bucket).MarshallJSON()
			assert.Nil(t, err)
			assert.Equal(t, tt.expectedJSON, string(b))
		})
	}
}
