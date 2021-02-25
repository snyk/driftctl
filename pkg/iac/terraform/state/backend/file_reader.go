package backend

import (
	"os"
)

const BackendKeyFile = ""

func NewFileReader(path string) (Backend, error) {
	return os.Open(path)
}
