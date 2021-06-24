package envproxy

import (
	"os"
	"strings"
)

type EnvProxy struct {
	varPrefix  string
	varPattern string
	DefaultEnv map[string]string
}

func NewEnvProxy() *EnvProxy {
	envMap := map[string]string{}
	for _, variable := range os.Environ() {
		tmp := strings.SplitN(variable, "=", 2)
		envMap[tmp[0]] = tmp[1]
	}
	return &EnvProxy{
		DefaultEnv: envMap,
	}
}

func (s *EnvProxy) SetProxy(prefix, pattern string) {
	s.varPrefix = prefix
	s.varPattern = pattern
}

func (s *EnvProxy) Apply() {
	if s.varPrefix == "" || s.varPattern == "" {
		return
	}
	for key, value := range s.DefaultEnv {
		if strings.HasPrefix(key, s.varPrefix) {
			key = strings.Replace(key, s.varPrefix, s.varPattern, 1)
			os.Setenv(key, value)
		}
	}
}

func (s *EnvProxy) Restore() {
	if s.varPrefix == "" || s.varPattern == "" {
		return
	}
	for key, value := range s.DefaultEnv {
		if strings.HasPrefix(key, s.varPrefix) {
			key = strings.Replace(key, s.varPrefix, s.varPattern, 1)
			value = s.DefaultEnv[key]
		}
		os.Setenv(key, value)
	}
}
