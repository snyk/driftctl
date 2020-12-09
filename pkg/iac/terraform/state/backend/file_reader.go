package backend

import (
	"os"
)

const backendFile = ""

func NewFileReader(path string) (Backend, error) {
	return os.Open(path)
}
