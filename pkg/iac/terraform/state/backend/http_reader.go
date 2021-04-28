package backend

import (
	"bytes"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
)

const BackendKeyHTTP = "http"
const BackendKeyHTTPS = "https"

type HTTPBackend struct {
	url    string
	reader io.ReadCloser
}

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

func NewHTTPReader(client HttpClient, rawURL string, opts *Options) (*HTTPBackend, error) {
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, err
	}

	for key, value := range opts.Headers {
		req.Header.Add(key, value)
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	logrus.WithFields(logrus.Fields{"body": string(body)}).Trace("HTTP(s) backend response")

	if res.StatusCode < 200 || res.StatusCode >= 400 {
		return nil, errors.Errorf("error requesting HTTP(s) backend state: status code: %d", res.StatusCode)
	}

	return &HTTPBackend{rawURL, ioutil.NopCloser(bytes.NewBuffer(body))}, nil
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
