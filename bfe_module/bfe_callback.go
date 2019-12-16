package bfe_module

import (
	"fmt"
	"github.com/sirupsen/logrus"
)

const (
	HANDLE_ACCEPT          = 0
	HANDLE_HANDSHAKE       = 1
	HANDLE_BEFORE_LOCATION = 2
	HANDLE_FOUND_PRODUCT   = 3
	HANDLE_AFTER_LOCATION  = 4
	HANDLE_FORWARD         = 5
	HANDLE_READ_BACKEND    = 6
	HANDLE_REQUEST_FINISH  = 7
	HANDLE_FINISH          = 8
)

type BfeCallbacks struct {
	callbacks map[int]*HandlerList
}

func NewBfeCallbacks() *BfeCallbacks {
	return &BfeCallbacks{
		callbacks: map[int]*HandlerList{
			HANDLE_ACCEPT:    NewHandlerList(HANDLERS_ACCEPT),
			HANDLE_HANDSHAKE: NewHandlerList(HANDLERS_ACCEPT),

			HANDLE_BEFORE_LOCATION: NewHandlerList(HANDLERS_REQUEST),
			HANDLE_FOUND_PRODUCT:   NewHandlerList(HANDLERS_REQUEST),
			HANDLE_AFTER_LOCATION:  NewHandlerList(HANDLERS_REQUEST),

			HANDLE_FORWARD: NewHandlerList(HANDLERS_FORWARD),

			HANDLE_READ_BACKEND:   NewHandlerList(HANDLERS_RESPONSE),
			HANDLE_REQUEST_FINISH: NewHandlerList(HANDLERS_RESPONSE),

			HANDLE_FINISH: NewHandlerList(HANDLERS_FINISH),
		},
	}
}

func (bcd *BfeCallbacks) AddFilter(point int, f interface{}) error {
	hl, ok := bcd.callbacks[point]

	if !ok {
		return fmt.Errorf("invalid callback point[%d]", point)
	}

	var err error
	switch hl.h_type {
	case HANDLERS_ACCEPT:
		err = hl.AddAcceptFilter(f)
	case HANDLERS_REQUEST:
		err = hl.AddRequestFilter(f)
	case HANDLERS_FORWARD:
		err = hl.AddForwardFilter(f)
	case HANDLERS_RESPONSE:
		err = hl.AddResponseFilter(f)
	case HANDLERS_FINISH:
		err = hl.AddFinishFilter(f)
	default:
		err = fmt.Errorf("invalid type of handler list[%d]", hl.h_type)
	}

	return err
}

func (bcb *BfeCallbacks) GetHandlerList(point int) *HandlerList {
	hl, ok := bcb.callbacks[point]

	if !ok {
		logrus.Warn("GetHandlerList(): invalid callback point[%d]", point)
		return nil
	}

	return hl
}
