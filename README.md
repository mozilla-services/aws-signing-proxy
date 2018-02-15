[![Go Report Card](https://goreportcard.com/badge/github.com/mozilla-services/aws-signing-proxy)](https://goreportcard.com/report/github.com/mozilla-services/aws-signing-proxy)

# aws-signing-proxy
signs http requests using AWS V4 signer

### Usage:

Running the signing proxy should be as simple as running a compiled binary or the `main.go` file.

### Configuration:

The signing proxy is configured via environment variables with the prefix `SIGNING_PROXY_`. The [config struct](https://github.com/mozilla-services/aws-signing-proxy/blob/master/main.go#L83-L92) has details on default values and variable types.

Available environment variables:

    - LOG_REQUESTS
        type: bool
        description: enable logging of request method and path
        default: "true"
    - STATSD
        type: bool
        description: enable statsd reporting
        default: "true"
    - STATSD_LISTEN
        type: string
        description: address to send statsd metrics to
        default: "127.0.0.1:8125"
    - STATSD_NAMESPACE
        type: string
        description: prefix for statsd metrics. "." is appended as a separator.
        default: "SIGNING_PROXY"
    - LISTEN
        type: string
        description: address for the proxy to listen on
        default: "localhost:8000"
    - SERVICE
        type: string
        description: aws service to sign requests for
        default: "s3"
    - REGION
        type: string
        description: aws region to sign requests for
        default: "us-east-1"
    - DESTINATION
        type: string
        description: valid URL that serves as a template for proxied requests. Scheme and Host are preserved for proxied requests.
        default: "https://s3.amazonaws.com"

### Building:

`go build` should be sufficient to build a binary

To build linux binaries on OSX for containers, I use [`gox`](https://github.com/mitchellh/gox): `gox -osarch="linux/amd64"`

### Docker:

We're using a `Dockerfile` `FROM scratch`, meaning there's nothing in there at the start.
We have the [Mozilla CA certificate store](https://curl.haxx.se/docs/caextract.html) in this repo, and copy it into our containers at build time.
This makes our image less than 11mb!

Images are available from Docker Hub: `docker pull mozilla/aws-signing-proxy`

### Development:

`dep` is used for package management:
  `dep ensure` to keep Gopkg.lock and vendored packages in sync

There is a simple `version` const in `main.go` for now that we can use to manually track versions.
