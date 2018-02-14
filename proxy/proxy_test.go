package proxy

import (
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/signer/v4"

	"github.com/sha1sum/aws_signing_client"

	"net/http"
	"net/url"
	"os"
	"testing"
)

var (
	creds         = credentials.NewStaticCredentials("ID", "SECRET", "TOKEN")
	signer        *v4.Signer
	client        *http.Client
	signingClient *http.Client
	proxy         *ProxyHandler
	signingProxy  *ProxyHandler
	service       string
	region        string
	destination   *url.URL
	err           error
)

func setup() {
	signer = v4.NewSigner(creds)
	client = http.DefaultClient
	service = "s3"
	region = "us-east-1"
	destination, _ = url.Parse("https://mozilla.org/")
	if err != nil {
		panic(err)
	}

	signingClient, _ = aws_signing_client.New(signer, http.DefaultClient, service, region)

	proxy, _ = New(destination, client)
	signingProxy, _ = New(destination, signingClient)

	err = nil
}

func teardown() {
	// no op
}

func TestMain(m *testing.M) {
	setup()
	ret := m.Run()
	teardown()
	os.Exit(ret)
}

func TestNewProxiedRequest(t *testing.T) {
	inputURL, _ := url.Parse("ftp://firefox.com/foo/bar")
	expectedURL, _ := url.Parse("https://mozilla.org/foo/bar")

	expectedMethod := "POST"
	request, err := http.NewRequest(expectedMethod, inputURL.String(), nil)
	if err != nil {
		panic(err)
	}

	for _, p := range []*ProxyHandler{proxy, signingProxy} {
		request, err = p.newProxiedRequest(request)
		if err != nil {
			panic(err)
		}
		// method should be unchanged by the proxy
		if request.Method != expectedMethod {
			t.Errorf("Expected request Method to be %s, got %s", expectedMethod, request.Method)
		}
		// url scheme should be changed to proxy's
		if request.URL.Scheme != expectedURL.Scheme {
			t.Errorf("Expected request Scheme to be %s, got %s", expectedURL.Scheme, request.URL.Scheme)
		}
		// url host should be changed to proxy's
		if request.URL.String() != expectedURL.String() {
			t.Errorf("Expected request URL to be %s, got %s", expectedURL.String(), request.URL.String())
		}
	}
}
