package utils

import (
	"github.com/Nike-Inc/cerberus-go-client/v3/api"
	"net/http"
)

var defaultHttpClient *http.Client = nil

func NewHttpClient(defaultHeaders http.Header) *http.Client {
	newClient := http.DefaultClient
	roundTripper := RoundTripperWithDefaultHeaders(http.DefaultTransport, defaultHeaders)
	newClient.Transport = roundTripper
	return newClient
}

func DefaultHttpClient() *http.Client {
	if defaultHttpClient == nil {
		newClient := http.DefaultClient
		roundTripper := RoundTripperWithDefaultHeaders(http.DefaultTransport, http.Header{})
		newClient.Transport = roundTripper
		defaultHttpClient = newClient
	}
	return defaultHttpClient
}

type roundTripperWithDefaultHeaders struct {
	http.Header
	rt http.RoundTripper
}

func RoundTripperWithDefaultHeaders(rt http.RoundTripper, defaultHeaders http.Header) roundTripperWithDefaultHeaders {
	if rt == nil {
		rt = http.DefaultTransport
	}
	return roundTripperWithDefaultHeaders{Header: defaultHeaders, rt: rt}
}

func (h roundTripperWithDefaultHeaders) RoundTrip(req *http.Request) (*http.Response, error) {
	for k, v := range h.Header {
		req.Header[k] = v
	}
	req.Header = AddClientHeader(req.Header)
	return h.rt.RoundTrip(req)
}

// utils.AddClientHeader is a helper to create the default client headers for every request
func AddClientHeader(headers http.Header) http.Header {
	if headers.Get("X-Cerberus-Client") == "" {
		headers.Set("X-Cerberus-Client", api.ClientHeader)
	}
	return headers
}
