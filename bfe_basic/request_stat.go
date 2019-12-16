package bfe_basic

import (
	"time"
)

type RequestStat struct {
	ReadReqStart time.Time
	ReadReqEnd   time.Time

	FindProStart time.Time
	FindProEnd   time.Time

	LocateStart time.Time
	LocateEnd   time.Time

	ClusterStart time.Time
	CLusterEnd   time.Time

	BackendStart time.Time
	BackendEnd   time.Time

	ResponseStart time.Time
	ResponseENs   time.Time

	HeaderLenIn  int
	BodyLenIm    int
	HeaderLenOut int
	BodyLenOut   int

	IsCrossCLuster bool
}

func NewRequestStat(start time.Time) *RequestStat {
	return &RequestStat{
		ReadReqStart: start,
	}
}
