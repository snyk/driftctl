package google

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"

	"cloud.google.com/go/storage"
)

type FakeStorageServer struct {
	routes map[string]http.HandlerFunc
}

func (s *FakeStorageServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for path, handler := range s.routes {
		if path == r.RequestURI {
			handler(w, r)
			return
		}
	}
	w.WriteHeader(http.StatusNotFound)
}

func NewFakeStorageServer(routes map[string]http.HandlerFunc) (*storage.Client, *httptest.Server, error) {
	return newStorageClient(&FakeStorageServer{routes: routes})
}

func newStorageClient(fakeServer *FakeStorageServer) (*storage.Client, *httptest.Server, error) {
	ts := httptest.NewServer(fakeServer)
	listenUrl, err := url.Parse(ts.URL)
	if err != nil {
		return nil, nil, err
	}

	_ = os.Setenv("STORAGE_EMULATOR_HOST", listenUrl.Host)
	defer os.Setenv("STORAGE_EMULATOR_HOST", "")

	client, err := storage.NewClient(context.Background())
	if err != nil {
		return nil, nil, err
	}
	return client, ts, nil
}
