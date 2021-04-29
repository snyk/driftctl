package backend

import (
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"io"
	"net/http"
)

const BackendKeyHTTP = "http"
const BackendKeyHTTPS = "https"

type HTTPBackend struct {
	request *http.Request
	client  HttpClient
	reader  io.ReadCloser
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

	return &HTTPBackend{req, client, nil}, nil
}

func (h *HTTPBackend) Read(p []byte) (n int, err error) {
	if h.reader == nil {
		res, err := h.client.Do(h.request)
		if err != nil {
			return 0, err
		}
		h.reader = res.Body

		if res.StatusCode < 200 || res.StatusCode >= 400 {
			body, _ := io.ReadAll(h.reader)
			logrus.WithFields(logrus.Fields{"body": string(body)}).Trace("HTTP(s) backend response")

			return 0, errors.Errorf("error requesting HTTP(s) backend state: status code: %d", res.StatusCode)
		}
	}
	return h.reader.Read(p)
}

func (h *HTTPBackend) Close() error {
	if h.reader != nil {
		return h.reader.Close()
	}
	return errors.New("Unable to close reader as nothing was opened")
}
