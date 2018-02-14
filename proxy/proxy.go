package proxy

import (
    "fmt"
    "io"
    "net/url"
    "net/http"
)

var (
    // error types
    MissingDestinationError = fmt.Errorf("No destination specified.")
)

func New(destination *url.URL, client *http.Client) (*ProxyHandler, error) {
    if destination == nil {
        return nil, MissingDestinationError
    }
    if client == nil {
        client = http.DefaultClient
    }
    return &ProxyHandler{destination, client}, nil
}

type ProxyHandler struct {
    destination *url.URL
    client *http.Client
}

func (proxy ProxyHandler) newProxiedRequest(r *http.Request) (*http.Request, error) {
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
func (proxy ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    defer r.Body.Close()

    proxiedReq, err := proxy.newProxiedRequest(r)
    if err != nil {
        http.Error(w, "Internal Server Error", 500)
        return
    }

    resp, err := proxy.client.Do(proxiedReq)
    if err != nil {
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
        panic(err)
    }
}
