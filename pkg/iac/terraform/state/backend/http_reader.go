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

func NewHTTPReader(rawURL string, opts *Options) (*HTTPBackend, error) {
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}

	for key, value := range opts.Headers {
		req.Header.Add(key, value)
	}

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return &HTTPBackend{rawURL, res.Body}, nil
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
