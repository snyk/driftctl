package test

import (
	"os"
	"path"
	"runtime"
)

func WriteTestFile(p string, content []byte) error {
	_, filename, _, _ := runtime.Caller(1)
	return os.WriteFile(path.Join(path.Dir(filename), p), content, 0644)
}

func ReadTestFile(p string) ([]byte, error) {
	_, filename, _, _ := runtime.Caller(1)
	return os.ReadFile(path.Join(path.Dir(filename), p))
}
