package bfe_module

import (
	"github.com/crud-bird/bfe/bfe_basic"
	"github.com/crud-bird/bfe/bfe_http"
)

type RequestFilter interface {
	FilterRequest(request *bfe_basic.Request) (int, *bfe_http.Response)
}

func NewRequestFilter(f func(request *bfe_basic.Request) (int, *bfe_http.Response)) RequestFilter {
	return &genericRequestFilter{
		f: f,
	}
}

type genericRequestFilter struct {
	f func(request *bfe_basic.Request) (int, *bfe_http.Response)
}

func (f *genericRequestFilter) FilterRequest(request *bfe_basic.Request) (int, *bfe_http.Response) {
	return f.f(request)
}

type ResponseFilter interface {
	FilterResponse(req *bfe_basic.Request, res *bfe_http.Response) int
}

func NewResponseFilter(f func(req *bfe_basic.Request, res *bfe_http.Response) int) ResponseFilter {
	return &genericResponseFilter{
		f: f,
	}
}

type genericResponseFilter struct {
	f func(req *bfe_basic.Request, res *bfe_http.Response) int
}

func (f *genericResponseFilter) FilterResponse(req *bfe_basic.Request, res *bfe_http.Response) int {
	return f.f(req, res)
}

type AcceptFilter interface {
	FilterAccept(*bfe_basic.Session) int
}

func NewAcceptFilter(f func(*bfe_basic.Session) int) AcceptFilter {
	return &genericAcceptFilter{
		f: f,
	}
}

type genericAcceptFilter struct {
	f func(*bfe_basic.Session) int
}

func (f *genericAcceptFilter) FilterAccept(session *bfe_basic.Session) int {
	return f.f(session)
}

type ForwardFilter interface {
	FilterForward(*bfe_basic.Request) int
}

func NewForwardFilter(f func(*bfe_basic.Request) int) ForwardFilter {
	return &genericForwardFilter{
		f: f,
	}
}

type genericForwardFilter struct {
	f func(*bfe_basic.Request) int
}

func (f *genericForwardFilter) FilterForward(req *bfe_basic.Request) int {
	return f.f(req)
}

type FinishFilter interface {
	FilterFinish(*bfe_basic.Session) int
}

func NewFinishFilter(f func(*bfe_basic.Session) int) FinishFilter {
	return &genericFinishFilter{
		f: f,
	}
}

type genericFinishFilter struct {
	f func(*bfe_basic.Session) int
}

func (f *genericFinishFilter) FilterFinish(session *bfe_basic.Session) int {
	return f.f(session)
}
