package http

import "net/http"

// HTTPClient is an interface for http.Client type
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}
