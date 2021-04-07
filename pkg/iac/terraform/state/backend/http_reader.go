package backend

import (
	"bytes"
	"io"
	"net/http"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
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

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(res.Body)
	if err != nil {
		return nil, err
	}

	logrus.WithFields(logrus.Fields{"body": buf.String()}).Trace("HTTP(s) backend response")

	if res.StatusCode < 200 || res.StatusCode >= 400 {
		return nil, errors.Errorf("error requesting HTTP(s) backend state: status code: %d", res.StatusCode)
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
