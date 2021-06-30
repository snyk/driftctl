package test

import (
	"io/ioutil"
	"path"
	"runtime"
)

func WriteTestFile(p string, content []byte) error {
	_, filename, _, _ := runtime.Caller(1)
	return ioutil.WriteFile(path.Join(path.Dir(filename), p), content, 0644)
}

func ReadTestFile(p string) ([]byte, error) {
	_, filename, _, _ := runtime.Caller(1)
	return ioutil.ReadFile(path.Join(path.Dir(filename), p))
}
