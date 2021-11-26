package acceptance

import (
	"context"
	"errors"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAccTestCase_resolveTerraformEnv(t *testing.T) {

	os.Clearenv()
	os.Setenv("ACC_TEST_VAR", "foobar")
	os.Setenv("TEST_VAR", "barfoo")
	os.Setenv("TEST_VAR_2", "barfoo")
	os.Setenv("ACC_TEST_VAR_3", "")
	os.Setenv("TEST_VAR_3", "barfoo")
	os.Setenv("TEST_VAR_4", "barfoo")
	os.Setenv("ACC_TEST_VAR_4", "")

	testCase := AccTestCase{}
	env := testCase.resolveTerraformEnv()
	expected := map[string]string{
		"TEST_VAR":   "foobar",
		"TEST_VAR_2": "barfoo",
		"TEST_VAR_3": "",
		"TEST_VAR_4": "",
	}

	if !reflect.DeepEqual(expected, env) {
		t.Fatalf("Variable env override not working, got: %+v, expected %+v", env, expected)
	}

}

func TestRetryFor(t *testing.T) {
	randomError := errors.New("random error")

	cases := []struct {
		name    string
		timeout time.Duration
		f       func(c chan struct{}) error
		err     error
	}{
		{
			name:    "success on first try",
			timeout: 100 * time.Millisecond,
			f: func(c chan struct{}) error {
				c <- struct{}{}
				return nil
			},
		},
		{
			name:    "timeout exceeded",
			timeout: 1 * time.Millisecond,
			f: func(c chan struct{}) error {
				return nil
			},
			err: context.DeadlineExceeded,
		},
		{
			name:    "error before timeout exceeded",
			timeout: 1 * time.Millisecond,
			f: func(c chan struct{}) error {
				return randomError
			},
			err: randomError,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(tt *testing.T) {
			err := RetryFor(c.timeout, c.f)
			assert.Equal(tt, c.err, err)
		})
	}
}
