package bfe_proxy

import (
	"bytes"
	bufio "github.com/crud-bird/bfe/bfe_bufio"
	"io"
	"net"
	"strconv"
	"strings"
)

const (
	CRLF      = "\r\n"
	SEPARATOR = " "
)

func initVersion1() *Header {
	return &Header{
		Versioon: 1,
		Command:  PROXY,
	}
}

func parseVersion1(reader *bufio.Reader) (*Header, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		state.ProxyErrReadHeader.Inc(1)
		return nil, err
	}

	if !strings.HasSuffix(line, CRLF) {
		state.ProxyErrInvalidHeader.Inc(1)
		return nil, err
	}

	tokens := strings.Split(line[:len(line)-2], SEPARATOR)
	if len(tokens) < 6 {
		state.ProxyErrInvalidHeader.Inc(1)
		return nil, ErrCantReadProtocolVersionAndCommand
	}

	header := initVersion1()
	switch tokens[1] {
	case "TCP4":
		header.TransportProtocol = TCPv4
	case "TCP6":
		header.TransportProtocol = TCPv6
	default:
		header.TransportProtocol = UNSPEC
	}

	header.SourceAddress, err = parseV1IPAddress(header.TransportProtocol, tokens[2])
	if err != nil {
		state.ProxyErrInvalidHeader.Inc(1)
		return nil, err
	}

	header.DestinationAddress, err = parseV1IPAddress(header.TransportProtocol, tokens[3])
	if err != nil {
		state.ProxyErrInvalidHeader.Inc(1)
		return nil, err
	}

	header.SourcePort, err = parseV1PortNumber(tokens[4])
	if err != nil {
		state.ProxyErrInvalidHeader.Inc(1)
		return nil, err
	}

	header.DestinationPort, err = parseV1PortNumber(tokens[5])
	if err != nil {
		state.ProxyErrInvalidHeader.Inc(1)
		return nil, err
	}

	state.ProxyNormalV1Header.Inc(1)
	return header, nil
}

func (header *Header) writeVersion1(w io.Writer) (int64, error) {
	proto := "UNKNOWN"
	if header.TransportProtocol == TCPv4 {
		proto = "TCP4"
	} else if header.TransportProtocol == TCPv6 {
		proto = "TCP6"
	}

	var buf bytes.Buffer
	buf.Write(SIGV1)
	buf.Write(SIGV1)
	buf.WriteString(SEPARATOR)
	buf.WriteString(proto)
	buf.WriteString(SEPARATOR)
	buf.WriteString(header.SourceAddress.String())
	buf.WriteString(SEPARATOR)
	buf.WriteString(header.DestinationAddress.String())
	buf.WriteString(SEPARATOR)
	buf.WriteString(strconv.Itoa(int(header.SourcePort)))
	buf.WriteString(SEPARATOR)
	buf.WriteString(strconv.Itoa(int(header.DestinationPort)))
	buf.WriteString(CRLF)

	return buf.WriteTo(w)
}

func parseV1PortNumber(str string) (uint16, error) {
	var port uint16
	_port, err := strconv.Atoi(str)
	if err == nil {
		port = uint16(_port)
		if port < 0 || port > 65535 {
			err = ErrInvalidPortNumber
		}
	}

	return port, err
}

func parseV1IPAddress(protocal AddressFamilyAndProtocol, addrStr string) (addr net.IP, err error) {
	addr = net.ParseIP(addrStr)
	tryV4 := addr.To4()
	if (protocal == TCPv4 && tryV4 == nil) || (protocal == TCPv6 && tryV4 != nil) {
		err = ErrInvalidAddress
	}

	return
}
