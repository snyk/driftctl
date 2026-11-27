package repository

import (
	"strings"
	"testing"

	"github.com/r3labs/diff/v2"
	"github.com/scaleway/scaleway-sdk-go/api/function/v1beta1"
	"github.com/snyk/driftctl/enumeration/remote/cache"
	scalewaytest "github.com/snyk/driftctl/test/scaleway"
	"github.com/stretchr/testify/assert"
)

func Test_functionRepository_ListAllNamespaces(t *testing.T) {
	tests := []struct {
		name    string
		mocks   func(client *scalewaytest.MockFakeFunction)
		want    []*function.Namespace
		wantErr error
	}{
		{
			name: "list",
			mocks: func(client *scalewaytest.MockFakeFunction) {
				client.On("ListNamespaces", &function.ListNamespacesRequest{}).
					Once().
					Return(
						&function.ListNamespacesResponse{
							Namespaces: []*function.Namespace{
								{
									Name: "namespace1",
								},
								{
									Name: "namespace2",
								},
							},
							TotalCount: 2,
						}, nil,
					)
			},
			want: []*function.Namespace{
				{
					Name: "namespace1",
				},
				{
					Name: "namespace2",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := cache.New(1)
			client := &scalewaytest.MockFakeFunction{}
			tt.mocks(client)
			r := &functionRepository{
				client: client,
				cache:  store,
			}
			got, err := r.ListAllNamespaces()
			assert.Equal(t, tt.wantErr, err)

			if err == nil {
				// Check that results were cached
				cachedData, err := r.ListAllNamespaces()
				assert.NoError(t, err)
				assert.Equal(t, got, cachedData)
				assert.IsType(t, []*function.Namespace{}, store.Get("functionListAllNamespaces"))
			}

			changelog, err := diff.Diff(got, tt.want)
			assert.Nil(t, err)
			if len(changelog) > 0 {
				for _, change := range changelog {
					t.Errorf("%s: %s -> %s", strings.Join(change.Path, "."), change.From, change.To)
				}
				t.Fail()
			}
		})
	}
}
