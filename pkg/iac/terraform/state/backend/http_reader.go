package backend

import (
	"io"
	"net/http"

	"github.com/pkg/errors"
)

const BackendKeyHTTP = "http"
const BackendKeyHTTPS = "https"

type HTTPBackend struct {
	url    string
	reader io.ReadCloser
}

func NewHTTPReader(url string) (*HTTPBackend, error) {
	client := &http.Client{}

	res, err := client.Get(url)
	if err != nil {
		return nil, err
	}

	return &HTTPBackend{url, res.Body}, nil
}

func (h *HTTPBackend) Read(p []byte) (n int, err error) {
	if h.reader == nil {
		return 0, errors.New("Reader not initialized")
	}
	return h.reader.Read(p)
}

func (h *HTTPBackend) Close() error {
	if h.reader != nil {
		return h.reader.Close()
	}
	return errors.New("Unable to close reader as nothing was opened")
}
