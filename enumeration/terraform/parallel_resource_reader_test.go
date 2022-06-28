package terraform

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/snyk/driftctl/enumeration/parallel"

	"github.com/r3labs/diff/v2"
	"github.com/stretchr/testify/assert"

	"github.com/zclconf/go-cty/cty"
)

func TestParallelResourceReader_Wait(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		name    string
		execs   []func() (cty.Value, error)
		want    []cty.Value
		wantErr bool
	}{
		{
			name: "Working // read resource",
			execs: []func() (cty.Value, error){
				func() (cty.Value, error) {
					return cty.BoolVal(true), nil
				},
				func() (cty.Value, error) {
					return cty.StringVal("test"), nil
				},
			},
			want:    []cty.Value{cty.BoolVal(true), cty.StringVal("test")},
			wantErr: false,
		},

		{
			name: "failing // read resource",
			execs: []func() (cty.Value, error){
				func() (cty.Value, error) {
					return cty.BoolVal(true), nil
				},
				func() (cty.Value, error) {
					return cty.NilVal, errors.New("error")
				},
				func() (cty.Value, error) {
					return cty.StringVal("test"), nil
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParallelResourceReader(parallel.NewParallelRunner(context.TODO(), 10))

			for _, fun := range tt.execs {
				p.Run(fun)
			}

			got, err := p.Wait()
			assert.Equal(tt.wantErr, err != nil)
			if tt.want != nil {
				changelog, err := diff.Diff(got, tt.want)
				if err != nil {
					panic(err)
				}
				if len(changelog) > 0 {
					for _, change := range changelog {
						t.Errorf("%s got = %v, want %v", strings.Join(change.Path, "."), change.From, change.To)
					}
				}
			}
		})
	}
}
