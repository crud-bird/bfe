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
	ResponseEnd   time.Time

	BackendFirst time.Time

	HeaderLenIn  int
	BodyLenIn    int
	HeaderLenOut int
	BodyLenOut   int

	IsCrossCLuster bool
}

func NewRequestStat(start time.Time) *RequestStat {
	return &RequestStat{
		ReadReqStart: start,
	}
}
