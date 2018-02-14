package proxy

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
)

var (
	// ErrMissingDestination is returned when New is called with a nil destination
	ErrMissingDestination = fmt.Errorf("no destination specified")
)

// New creates a Handler using the input destination and client
func New(destination *url.URL, client *http.Client) (*Handler, error) {
	if destination == nil {
		return nil, ErrMissingDestination
	}
	if client == nil {
		client = http.DefaultClient
	}
	return &Handler{destination, client}, nil
}

// Handler satisfies http.Handler
type Handler struct {
	destination *url.URL
	client      *http.Client
}

func (proxy Handler) newProxiedRequest(r *http.Request) (*http.Request, error) {
	proxiedURL := *r.URL
	proxiedURL.Host = proxy.destination.Host
	proxiedURL.Scheme = proxy.destination.Scheme

	return http.NewRequest(
		r.Method,
		proxiedURL.String(),
		r.Body,
	)
}

// satisfies http.Handler
// per https://golang.org/pkg/net/http/#Handler
// the server will recover panic() and log a stack trace
func (proxy Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	proxiedReq, err := proxy.newProxiedRequest(r)
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		panic(err)
	}

	resp, err := proxy.client.Do(proxiedReq)
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		panic(err)
	}
	defer resp.Body.Close()

	// add all response headers to request
	for k, vs := range resp.Header {
		for _, v := range vs {
			w.Header().Add(k, v)
		}
	}

	w.WriteHeader(resp.StatusCode)
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		http.Error(w, "Internal Server Error", 500)
		panic(err)
	}
}
