package envproxy

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnvProxy_Apply(t *testing.T) {
	tests := []struct {
		name      string
		proxyArgs []string
		modifier  bool
	}{
		{
			name:      "Without args on SetProxy",
			proxyArgs: []string{"", ""},
			modifier:  false,
		},
		{
			name:      "With args on SetProxy",
			proxyArgs: []string{"TEST_DCTL_S3_", "TEST_AWS_"},
			modifier:  true,
		},
		{
			name:      "With no pattern on SetProxy",
			proxyArgs: []string{"TEST_DCTL_S3_", ""},
			modifier:  false,
		},
		{
			name:      "With no prefix on SetProxy",
			proxyArgs: []string{"", "TEST_AWS_"},
			modifier:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("TEST_DCTL_S3_PROFILE", "dctl_env")
			os.Setenv("TEST_AWS_PROFILE", "aws_env")
			expectedEnv := os.Environ()

			envProxy := NewEnvProxy()
			envProxy.SetProxy(tt.proxyArgs[0], tt.proxyArgs[1])

			envProxy.Apply()

			newEnv := os.Environ()
			for index, value := range expectedEnv {
				if tt.modifier && value == "TEST_AWS_PROFILE=aws_env" {
					expectedEnv[index] = strings.Replace(
						value,
						"TEST_AWS_PROFILE=aws_env",
						"TEST_AWS_PROFILE=dctl_env",
						1,
					)
				}
			}

			if !assert.Equal(t, newEnv, expectedEnv) {
				t.Errorf("Expected %v, got %v", expectedEnv, newEnv)
			}
		})
	}
}

func TestEnvProxy_Restore(t *testing.T) {
	tests := []struct {
		name      string
		proxyArgs []string
	}{
		{
			name:      "With args on SetProxy",
			proxyArgs: []string{"TEST_DCTL_S3_", "TEST_AWS_"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("TEST_DCTL_S3_PROFILE", "dctl_env")
			os.Setenv("TEST_AWS_PROFILE", "aws_env")
			expectedEnv := os.Environ()

			envProxy := NewEnvProxy()
			envProxy.SetProxy(tt.proxyArgs[0], tt.proxyArgs[1])
			os.Setenv("TEST_AWS_PROFILE", "new_aws_env")

			envProxy.Restore()

			newEnv := os.Environ()
			if !assert.Equal(t, newEnv, expectedEnv) {
				t.Errorf("Expected %v, got %v", expectedEnv, newEnv)
			}
		})
	}
}
