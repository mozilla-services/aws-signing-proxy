package main

import (
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/aws/signer/v4"

    "github.com/sha1sum/aws_signing_client"
    "github.com/gorilla/mux"
    "github.com/kelseyhightower/envconfig"

    "fmt"
    "io"
    "net/url"
    "net/http"
    "time"
)

type SigningProxy struct {
    Destination *url.URL
    Signer *v4.Signer
    ServiceName string
    Region string
}

func LoggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        fmt.Println(r.RequestURI)
        next.ServeHTTP(w, r)
    })
}

// satisfies http.Handler
func (proxy SigningProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {

    defer r.Body.Close()

    proxiedURL := *r.URL
    proxiedURL.Host = proxy.Destination.Host
    proxiedURL.Scheme = proxy.Destination.Scheme

    proxiedReq, err := http.NewRequest(
        r.Method,
        proxiedURL.String(),
        r.Body,
    )
    if err != nil {
        http.Error(w, "Internal Server Error", 500)
        return
    }

    awsClient, err := aws_signing_client.New(
        proxy.Signer,
        // default http client with a timeout
        &http.Client{
            Timeout: time.Second * 10,
        },
        proxy.ServiceName,
        proxy.Region,
    )
    if err != nil {
        panic(err)
    }

    resp, err := awsClient.Do(proxiedReq)
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

func main() {
    config := struct {
        // SIGNING_PROXY_LOG_REQUESTS
        LogRequests bool `default:"true" split_words:"true"`
        Listen string `default:"localhost:8000"`
        Service string `default:"s3"`
        Region string `default:"us-east-1"`
        Destination string `default:"https://s3.amazonaws.com"`
    }{}

    err := envconfig.Process("SIGNING_PROXY", &config)
    if err != nil {
        panic(err)
    }
    fmt.Println(config)

    destinationURL, err := url.Parse(config.Destination)
    if err != nil {
        panic(err)
    }

    sess, err := session.NewSession(&aws.Config{
        Region: aws.String(config.Region),
    })
    signer := v4.NewSigner(sess.Config.Credentials)

    proxy := SigningProxy{
        destinationURL,
        signer,
        config.Service,
        config.Region,
    }

    router := mux.NewRouter()
    router.Handle("/", proxy)

    if config.LogRequests {
        router.Use(LoggingMiddleware)
    }

    server := &http.Server{
        Addr: config.Listen,
        ReadTimeout: 5 * time.Second,
        WriteTimeout: 10 * time.Second,
        IdleTimeout: 60 * time.Second,
        Handler: router,
    }

    fmt.Println(server.ListenAndServe())
}