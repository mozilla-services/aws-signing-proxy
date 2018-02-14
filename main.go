package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/aws/signer/v4"

	"github.com/DataDog/datadog-go/statsd"
	"github.com/kelseyhightower/envconfig"
	"github.com/sha1sum/aws_signing_client"

	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/mozilla-services/aws-signing-proxy/proxy"
)

const (
	appNamespace = "SIGNING_PROXY"
	version      = "1.0.1"
)

var (
	statsdClient *statsd.Client
	httpClient   *http.Client
	pool         *x509.CertPool
)

// get CA certs for our http.Client
func init() {
	// cacert.pem is a runtime dependency!
	bs, err := ioutil.ReadFile("cacert.pem")
	if err != nil {
		log.Fatal(err.Error())
	}

	pool = x509.NewCertPool()
	pool.AppendCertsFromPEM(bs)

	// default http client with a timeout
	httpClient = &http.Client{
		Timeout: time.Second * 10,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{RootCAs: pool},
		},
	}
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.Method + " " + r.RequestURI)
		next.ServeHTTP(w, r)
	})
}

func statsdMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := statsdClient.Incr("requests", []string{}, 1.0)
		if err != nil {
			log.Println(err.Error())
		}
		next.ServeHTTP(w, r)
	})
}

func getEC2Tags(metadata *ec2metadata.EC2Metadata) []string {
	region, err := metadata.Region()
	if err != nil {
		log.Println(err.Error())
	}
	return []string{
		"region:" + region,
	}
}

func main() {
	config := struct {
		LogRequests     bool   `default:"true" split_words:"true"`
		Statsd          bool   `default:"true"`
		StatsdListen    string `default:"127.0.0.1:8125" split_words:"true"`
		StatsdNamespace string `default:""`
		Listen          string `default:"localhost:8000"`
		Service         string `default:"s3"`
		Region          string `default:"us-east-1"`
		Destination     string `default:"https://s3.amazonaws.com"`
	}{}

	// load envconfig
	err := envconfig.Process(appNamespace, &config)
	if err != nil {
		log.Fatal(err.Error())
	}

	// *url.URL for convenience
	destinationURL, err := url.Parse(config.Destination)
	if err != nil {
		log.Fatal("Could not parse destination URL: " + err.Error())
	}

	// initialize AWS session
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(config.Region),
	}))

	ec2tags := []string{}
	metadata := ec2metadata.New(sess)
	if metadata.Available() {
		ec2tags = getEC2Tags(metadata)
	}

	// create signing http client
	signer := v4.NewSigner(sess.Config.Credentials)
	signingClient, err := aws_signing_client.New(
		signer,
		httpClient,
		config.Service,
		config.Region,
	)
	if err != nil {
		log.Fatal(err.Error())
	}

	// create proxy using signing http client
	prxy, err := proxy.New(
		destinationURL,
		signingClient,
	)
	if err != nil {
		log.Fatal(err.Error())
	}

	var handler http.Handler
	handler = prxy

	// wrap handler for logging
	if config.LogRequests {
		handler = loggingMiddleware(handler)
	}

	// wrap handler for statsd
	if config.Statsd {
		statsdClient, err := statsd.New(config.StatsdListen)
		if err != nil {
			log.Fatal(err.Error())
		}
		// prepended to metrics
		if config.StatsdNamespace == "" {
			statsdClient.Namespace = appNamespace + "."
		} else {
			statsdClient.Namespace = config.StatsdNamespace + "."
		}
		statsdClient.Tags = append(statsdClient.Tags, ec2tags...)
		handler = statsdMiddleware(handler)
	}

	// sane default timeouts
	srv := &http.Server{
		Addr:         config.Listen,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
		Handler:      handler,
	}

	err = srv.ListenAndServe()
	if err != nil {
		log.Fatal(err.Error())
	}
}
