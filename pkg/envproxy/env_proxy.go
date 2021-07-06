package envproxy

import (
	"os"
	"strings"
)

type EnvProxy struct {
	fromPrefix string
	toPrefix   string
	defaultEnv map[string]string
}

func NewEnvProxy(fromPrefix, toPrefix string) *EnvProxy {
	envMap := map[string]string{}
	for _, variable := range os.Environ() {
		tmp := strings.SplitN(variable, "=", 2)
		envMap[tmp[0]] = tmp[1]
	}
	return &EnvProxy{
		fromPrefix: fromPrefix,
		toPrefix:   toPrefix,
		defaultEnv: envMap,
	}
}

func (s *EnvProxy) Apply() {
	if s.fromPrefix == "" || s.toPrefix == "" {
		return
	}
	for key, value := range s.defaultEnv {
		if strings.HasPrefix(key, s.fromPrefix) {
			key = strings.Replace(key, s.fromPrefix, s.toPrefix, 1)
			os.Setenv(key, value)
		}
	}
}

func (s *EnvProxy) Restore() {
	if s.fromPrefix == "" || s.toPrefix == "" {
		return
	}
	for key, value := range s.defaultEnv {
		if strings.HasPrefix(key, s.fromPrefix) {
			key = strings.Replace(key, s.fromPrefix, s.toPrefix, 1)
			value = s.defaultEnv[key]
		}
		os.Setenv(key, value)
	}
}
