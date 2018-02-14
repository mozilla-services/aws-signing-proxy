[![Go Report Card](https://goreportcard.com/badge/github.com/milescrabill/aws-signing-proxy)](https://goreportcard.com/report/github.com/milescrabill/aws-signing-proxy)

# aws-signing-proxy
signs http requests using AWS V4 signer

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
