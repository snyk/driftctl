package envproxy

import (
	"os"
	"strings"
	"testing"
)

func TestEnvProxy(t *testing.T) {
	tests := []struct {
		name        string
		proxyArgs   []string
		initialEnv  []string
		modifiedEnv []string
	}{
		{
			name:        "Without args on SetProxy",
			proxyArgs:   []string{"", ""},
			initialEnv:  []string{"TEST_DCTL_S3_PROFILE=test_dctl_s3_profile", "TEST_AWS_PROFILE=test_aws_profile"},
			modifiedEnv: []string{"TEST_DCTL_S3_PROFILE=test_dctl_s3_profile", "TEST_AWS_PROFILE=test_aws_profile"},
		},
		{
			name:        "With args on SetProxy",
			proxyArgs:   []string{"TEST_DCTL_S3_", "TEST_AWS_"},
			initialEnv:  []string{"TEST_DCTL_S3_PROFILE=test_dctl_s3_profile", "TEST_AWS_PROFILE=test_aws_profile"},
			modifiedEnv: []string{"TEST_DCTL_S3_PROFILE=test_dctl_s3_profile", "TEST_AWS_PROFILE=test_dctl_s3_profile"},
		},
		{
			name:        "Without toPrefix on SetProxy",
			proxyArgs:   []string{"TEST_DCTL_S3_", ""},
			initialEnv:  []string{"TEST_DCTL_S3_PROFILE=test_dctl_s3_profile", "TEST_AWS_PROFILE=test_aws_profile"},
			modifiedEnv: []string{"TEST_DCTL_S3_PROFILE=test_dctl_s3_profile", "TEST_AWS_PROFILE=test_aws_profile"},
		},
		{
			name:        "Without fromPrefix on SetProxy",
			proxyArgs:   []string{"", "TEST_AWS_"},
			initialEnv:  []string{"TEST_DCTL_S3_PROFILE=test_dctl_s3_profile", "TEST_AWS_PROFILE=test_aws_profile"},
			modifiedEnv: []string{"TEST_DCTL_S3_PROFILE=test_dctl_s3_profile", "TEST_AWS_PROFILE=test_aws_profile"},
		},
		{
			name:        "Without initialEnv",
			proxyArgs:   []string{"TEST_DCTL_S3_", "TEST_AWS_"},
			initialEnv:  []string{},
			modifiedEnv: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			for _, value := range tt.initialEnv {
				tmp := strings.SplitN(value, "=", 2)
				os.Setenv(tmp[0], tmp[1])
			}

			envProxy := NewEnvProxy(tt.proxyArgs[0], tt.proxyArgs[1])

			envProxy.Apply()

			currentEnv := os.Environ()
			if !compareEnv(currentEnv, tt.modifiedEnv) {
				t.Errorf("Expected %v, got %v", tt.modifiedEnv, currentEnv)
			}

			envProxy.Restore()

			currentEnv = os.Environ()
			if !compareEnv(currentEnv, tt.initialEnv) {
				t.Errorf("Expected %v, got %v", tt.initialEnv, currentEnv)
			}
		})
	}
}

func compareEnv(currentEnv, testEnv []string) bool {
	isValid := 0
	for _, initialValue := range testEnv {
		for _, value := range currentEnv {
			if initialValue == value {
				isValid++
			}
		}
	}
	return isValid == len(testEnv)
}
