package bfe_proxy

import (
	"bytes"
	"errors"
	bufio "github.com/crud-bird/bfe/bfe_bufio"
	"io"
	"net"
	"time"
)

var (
	SIGV1 = []byte{'\x50', '\x52', '\x4F', '\x58', '\x59'}
	SIGV2 = []byte{'\x0D', '\x0A', '\x0D', '\x0A', '\x00', '\x0D', '\x0A', '\x51', '\x55', '\x49', '\x54', '\x0A'}
)

var (
	ErrCantReadProtocolVersionAndCommand    = errors.New("Can't read proxy protocol version and command")
	ErrCantReadAddressFamilyAndProtocol     = errors.New("Can't read address family or protocol")
	ErrCantReadLength                       = errors.New("Can't read length")
	ErrCantResolveSourceUnixAddress         = errors.New("Can't resolve source Unix address")
	ErrCantResolveDestinationUnixAddress    = errors.New("Can't resolve destination Unix address")
	ErrNoProxyProtocol                      = errors.New("Proxy protocol signature not present")
	ErrUnknownProxyProtocolVersion          = errors.New("Unknown proxy protocol version")
	ErrUnsupportedProtocolVersionAndCommand = errors.New("Unsupported proxy protocol version and command")
	ErrUnsupportedAddressFamilyAndProtocol  = errors.New("Unsupported address family and protocol")
	ErrInvalidLength                        = errors.New("Invalid length")
	ErrLengthExceeded                       = errors.New("Length Exceeded")
	ErrInvalidAddress                       = errors.New("Invalid address")
	ErrInvalidPortNumber                    = errors.New("Invalid port number")
)

type Header struct {
	Versioon           byte
	Command            ProtocolVersionAndCommand
	TransportProtocol  AddressFamilyAndProtocol
	SourceAddress      net.IP
	DestinationAddress net.IP
	SourcePort         uint16
	DestinationPort    uint16
}

func (header *Header) EqualTo(q *Header) bool {
	if header == nil || q == nil {
		return false
	}

	if header.Command.IsLocal() {
		return true
	}

	return header.TransportProtocol == q.TransportProtocol &&
		header.SourceAddress.String() == q.SourceAddress.String() &&
		header.DestinationAddress.String() == q.DestinationAddress.String() &&
		header.SourcePort == q.SourcePort &&
		header.DestinationPort == q.DestinationPort
}

func (header *Header) WriteTo(w io.Writer) (int64, error) {
	switch header.Versioon {
	case 1:
		return header.writeVersion1(w)
	case 2:
		return header.writeVersion2(w)
	default:
		return 0, ErrUnknownProxyProtocolVersion
	}
}

func Read(reader *bufio.Reader) (*Header, error) {
	b1, err := reader.Peek(1)
	if err != nil {
		state.ProxyErrReadHeader.Inc(1)
		return nil, err
	}
	if !bytes.Equal(b1[:1], SIGV1[:1]) && !bytes.Equal(b1[:1], SIGV2[:1]) {
		state.ProxyErrNoProxyProtocol.Inc(1)
		return nil, ErrNoProxyProtocol
	}

	signature, err := reader.Peek(5)
	if err != nil {
		state.ProxyErrReadHeader.Inc(1)
		return nil, err
	}
	if bytes.Equal(signature[:5], SIGV1) {
		state.ProxyMatchedV1Signature.Inc(1)
		return parseVersion1(reader)
	}

	signature, err = reader.Peek(12)
	if err != nil {
		state.ProxyErrReadHeader.Inc(1)
		return nil, err
	}
	if bytes.Equal(signature[:12], SIGV2) {
		state.ProxyMatchedV2Signature.Inc(2)
		return parseVersion2(reader)
	}

	state.ProxyErrNoProxyProtocol.Inc(1)

	return nil, ErrNoProxyProtocol
}

func ReadTimeout(reader *bufio.Reader, timeout time.Duration) (*Header, error) {
	type header struct {
		h *Header
		e error
	}

	read := make(chan *header, 1)

	go func() {
		h := &header{}
		h.h, h.e = Read(reader)
		read <- h
	}()
	timer := time.NewTimer(timeout)
	select {
	case res := <-read:
		timer.Stop()
		return res.h, res.e
	case <-timer.C:
		return nil, ErrNoProxyProtocol
	}
}
