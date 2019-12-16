package bfe_http

type RoundTripper interface {
	RoudTrip(*Request) (*Response, error)
}
