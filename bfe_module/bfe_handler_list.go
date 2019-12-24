package bfe_module

import (
	"container/list"
	"fmt"

	"github.com/crud-bird/bfe/bfe_basic"
	"github.com/crud-bird/bfe/bfe_http"
	"github.com/sirupsen/logrus"
)

const (
	HANDLERS_ACCEPT   = 0 // for AcceptFilter
	HANDLERS_REQUEST  = 1 // for RequestFilter
	HANDLERS_FORWARD  = 2 // for ForwardFilter
	HANDLERS_RESPONSE = 3 // for ResponseFilter
	HANDLERS_FINISH   = 4 // for FinishFilter
)

const (
	BFE_HANDLER_FINISH   = 0 // to close the connection after response
	BFE_HANDLER_GOON     = 1 // to go on next handler
	BFE_HANDLER_REDIRECT = 2 // to redirect
	BFE_HANDLER_RESPONSE = 3 // to send response
	BFE_HANDLER_CLOSE    = 4 // to close the connection directly, with no data sent.
)

type HandlerList struct {
	h_type   int
	handlers *list.List
}

func NewHandlerList(h_type int) *HandlerList {
	return &HandlerList{
		h_type:   h_type,
		handlers: list.New(),
	}
}

func (hl *HandlerList) FilterAccept(session *bfe_basic.Session) int {
	retVal := BFE_HANDLER_GOON

LOOP:
	for e := hl.handlers.Front(); e != nil; e = e.Next() {
		switch filter := e.Value.(type) {
		case AcceptFilter:
			retVal = filter.FilterAccept(session)
			if retVal != BFE_HANDLER_GOON {
				break LOOP
			}

		default:
			logrus.Errorf("%v (%T) is not a AcceptFilter", e.Value, e.Value)
			break LOOP
		}
	}

	return retVal
}

func (hl *HandlerList) FilterRequest(req *bfe_basic.Request) (int, *bfe_http.Response) {
	var res *bfe_http.Response
	retVal := BFE_HANDLER_GOON
LOOP:
	for e := hl.handlers.Front(); e != nil; e = e.Next() {
		switch filter := e.Value.(type) {
		case RequestFilter:
			retVal, res = filter.FilterRequest(req)
			if retVal != BFE_HANDLER_GOON {
				break LOOP
			}
		default:
			logrus.Errorf("%v (%T) is not a RequestFilter", e.Value, e.Value)
			break LOOP
		}
	}

	return retVal, res
}

func (hl *HandlerList) FilterForward(req *bfe_basic.Request) int {
	retVal := BFE_HANDLER_GOON
LOOP:
	for e := hl.handlers.Front(); e != nil; e = e.Next() {
		switch filter := e.Value.(type) {
		case ForwardFilter:
			retVal = filter.FilterForward(req)
			if retVal != BFE_HANDLER_GOON {
				break LOOP
			}
		default:
			logrus.Errorf("%v (%T) is not a ForwardFilter", e.Value, e.Value)
			break LOOP
		}
	}

	return retVal
}

func (hl *HandlerList) FilterResponse(req *bfe_basic.Request, res *bfe_http.Response) int {
	retVal := BFE_HANDLER_GOON
LOOP:
	for e := hl.handlers.Front(); e != nil; e = e.Next() {
		switch filter := e.Value.(type) {
		case ResponseFilter:
			retVal = filter.FilterResponse(req, res)
			if retVal != BFE_HANDLER_GOON {
				break LOOP
			}
		default:
			logrus.Errorf("%v (%T) is not a ResponseFilter", e.Value, e.Value)
			break LOOP
		}
	}

	return retVal
}

func (hl *HandlerList) FilterFinish(session *bfe_basic.Session) int {
	retVal := BFE_HANDLER_GOON

LOOP:
	for e := hl.handlers.Front(); e != nil; e = e.Next() {
		switch filter := e.Value.(type) {
		case FinishFilter:
			retVal = filter.FilterFinish(session)
			if retVal != BFE_HANDLER_GOON {
				break LOOP
			}
		default:
			logrus.Errorf("%v (%T) is not a FinishFilter", e.Value, e.Value)
			break LOOP
		}
	}

	return retVal
}

func (hl *HandlerList) AddAcceptFilter(f interface{}) error {
	callback, ok := f.(func(*bfe_basic.Session) int)
	if !ok {
		return fmt.Errorf("AddAcceptFilter(): invalid callback function")
	}

	hl.handlers.PushBack(NewAcceptFilter(callback))
	return nil
}

func (hl *HandlerList) AddRequestFilter(f interface{}) error {
	callback, ok := f.(func(*bfe_basic.Request) (int, *bfe_http.Response))
	if !ok {
		return fmt.Errorf("AddRequestFilter(); invalid callback function")
	}

	hl.handlers.PushBack(NewRequestFilter(callback))

	return nil
}

func (hl *HandlerList) AddForwardFilter(f interface{}) error {
	callback, ok := f.(func(*bfe_basic.Request) int)
	if !ok {
		return fmt.Errorf("AddForwardFilter(); invalid callback function")
	}

	hl.handlers.PushBack(NewForwardFilter(callback))
	return nil
}

func (hl *HandlerList) AddResponseFilter(f interface{}) error {
	callback, ok := f.(func(*bfe_basic.Request, *bfe_http.Response) int)
	if !ok {
		return fmt.Errorf("AddResponseFilter(): invalid callback function")
	}

	hl.handlers.PushBack(NewResponseFilter(callback))
	return nil
}

func (hl HandlerList) AddFinishFilter(f interface{}) error {
	callback, ok := f.(func(*bfe_basic.Session) int)
	if !ok {
		return fmt.Errorf("AddFInishFilter(): invalid callback function")
	}

	hl.handlers.PushBack(NewFinishFilter(callback))
	return nil
}
