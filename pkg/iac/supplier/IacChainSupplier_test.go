package supplier

import (
	"reflect"
	"testing"

	"github.com/pkg/errors"
	"github.com/snyk/driftctl/enumeration/resource"
	resource2 "github.com/snyk/driftctl/pkg/resource"
)

func TestIacChainSupplier_Resources(t *testing.T) {
	tests := []struct {
		name          string
		initSuppliers func(suppliers *[]resource2.IaCSupplier)
		want          []*resource.Resource
		wantErr       bool
	}{
		{
			name: "All failed",
			initSuppliers: func(suppliers *[]resource2.IaCSupplier) {
				sup := &resource2.MockIaCSupplier{}
				sup.On("Resources").Return(nil, errors.New("1"))
				*suppliers = append(*suppliers, sup)

				sup = &resource2.MockIaCSupplier{}
				sup.On("Resources").Return(nil, errors.New("2"))
				*suppliers = append(*suppliers, sup)

				sup = &resource2.MockIaCSupplier{}
				sup.On("Resources").Return(nil, errors.New("3"))
				*suppliers = append(*suppliers, sup)
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Partial failed",
			initSuppliers: func(suppliers *[]resource2.IaCSupplier) {
				sup := &resource2.MockIaCSupplier{}
				sup.On("Resources").Return(nil, errors.New("1"))
				*suppliers = append(*suppliers, sup)

				sup = &resource2.MockIaCSupplier{}
				sup.On("Resources").Return(nil, errors.New("2"))
				*suppliers = append(*suppliers, sup)

				sup = &resource2.MockIaCSupplier{}
				sup.On("Resources").Return([]*resource.Resource{
					&resource.Resource{
						Id:    "ID",
						Type:  "TYPE",
						Attrs: nil,
					},
				}, nil)
				*suppliers = append(*suppliers, sup)
			},
			want: []*resource.Resource{
				&resource.Resource{
					Id:    "ID",
					Type:  "TYPE",
					Attrs: nil,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewIacChainSupplier()
			suppliers := make([]resource2.IaCSupplier, 0)
			tt.initSuppliers(&suppliers)

			for _, supplier := range suppliers {
				r.AddSupplier(supplier)
			}

			got, err := r.Resources()
			if (err != nil) != tt.wantErr {
				t.Errorf("Resources() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Resources() got = %v, want %v", got, tt.want)
			}
		})
	}
}
