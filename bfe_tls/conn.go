package bfe_tls

import (
	"bytes"
	"crypto/cipher"
	"crypto/subtle"
	"crypto/x509"
	"errors"
	"io"
	"net"
	"sync"
	"time"
)

type ConnParam interface {
	GetVip() net.IP
}

type Conn struct {
	conn net.Conn

	isClient bool

	param ConnParam

	handshakeMutex      sync.Mutex
	handshakeErr        error
	vers                uint16
	haveVers            bool
	config              *Config
	handshakeTime       time.Duration
	handshakeComplete   bool
	didResume           bool
	cipherSuite         uint16
	ocspStaple          bool
	ocspResponse        []byte
	peerCertificates    []*x509.Certificate
	verifiedChains      [][]*x509.Certificate
	serverName          string
	grade               string
	clientAuth          ClientAuthType
	clientCAs           *x509.CertPool
	enableDynamicRecord bool
	clientRandom        []byte
	serverRandom        []byte
	masterSecret        []byte
	clientCiphers       []uint16

	clinetProtocol         string
	clientProtocolFallback bool

	in, out  halfConn
	rawInput *block
	input    *block
	hand     bytes.Buffer

	byteOut          int
	readFromUntilLen int

	lastOut time.Time

	tmp [16]byte

	sslv2Data []byte
}

func (c *Conn) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *Conn) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *Conn) GetNetConn() net.Conn {
	return c.conn
}

func (c *Conn) GetVip() net.IP {
	if c.param == nil {
		return nil
	}

	return c.param.GetVip()
}

func (c *Conn) GetServerName() string {
	return c.serverName
}

func (c *Conn) SetConnParam(param ConnParam) {
	c.param = param
}

func (c *Conn) SetDeadline(t time.Time) error {
	return c.conn.SetDeadline(t)
}

func (c *Conn) SetWriteDeadline(t time.Time) error {
	return c.conn.SetWriteDeadline(t)
}

func (c *Conn) SetReadDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}

func (c *Conn) getClientCAs() *x509.CertPool {
	clientCAs := c.config.ClientCAs
	if c.clientCAs != nil {
		clientCAs = c.clientCAs
	}

	return clientCAs
}

type halfConn struct {
	sync.Mutex

	err     error
	version uint16
	cipher  interface{}
	mac     macFunction
	seq     [8]byte
	bfree   *block

	nextCipher interface{}
	nextMac    macFunction

	inDigestBuf, outDigestBuf []byte
}

func (hc *halfConn) setErrorLocked(err error) error {
	hc.err = err
	return err
}

func (hc *halfConn) error() error {
	hc.Lock()
	err := hc.err
	hc.Unlock()

	return err
}

func (hc *halfConn) prepareCipherSpec(version uint16, cipher interface{}, mac macFunction) {
	hc.version = version
	hc.nextCipher = cipher
	hc.nextMac = mac
}

func (hc *halfConn) changeCipherSpec() error {
	if hc.nextCipher == nil {
		return alertInternalError
	}

	hc.cipher = hc.nextCipher
	hc.mac = hc.nextMac
	hc.nextCipher = nil
	hc.nextMac = nil

	for i := range hc.seq {
		hc.seq[i] = 0
	}

	return nil
}

func (hc *halfConn) incSeq() {
	for i := 7; i >= 0; i-- {
		hc.seq[i]++
		if hc.seq[i] != 0 {
			return
		}
	}

	panic("TLS: sequence number wraparoud")
}

func (hc *halfConn) resetSeq() {
	for i := range hc.seq {
		hc.seq[i] = 0
	}
}

func removePadding(payload []byte) ([]byte, byte) {
	if len(payload) < 1 {
		return payload, 0
	}

	paddingLen := payload[len(payload)-1]
	t := uint(len(payload)-1) - uint(paddingLen)
	good := byte(int32(^t) >> 31)

	toCheck := 255

	if toCheck+1 > len(payload) {
		toCheck = len(payload) - 1
	}

	for i := 0; i < toCheck; i++ {
		t := uint(paddingLen) - uint(i)
		mask := byte(int32(^t) >> 31)
		b := payload[len(payload)-1-i]
		good &^= mask&paddingLen ^ mask&b
	}

	good &= good << 4
	good &= good << 2
	good &= good << 1
	good = uint8(int8(good) >> 7)

	toRemove := good&paddingLen + 1
	return payload[:len(payload)-int(toRemove)], good
}

func removePaddinSSL30(payload []byte) ([]byte, byte) {
	if len(payload) < 1 {
		return payload, 0
	}

	paddingLen := int(payload[len(payload)-1]) + 1
	if paddingLen > len(payload) {
		return payload, 0
	}

	return payload[:len(payload)-paddingLen], 255
}

func roundUp(a, b int) int {
	return a + (b-a%b)%b
}

type cbcMode interface {
	cipher.BlockMode
	SetIV([]byte)
}

func (hc *halfConn) decrypt(b *block) (ok bool, prefixLen int, alertValue alert) {
	payload := b.data[recordHeaderLen:]

	macSize := 0
	if hc.mac != nil {
		macSize = hc.mac.Size()
	}

	paddingGood := byte(255)
	explicitIVLen := 0

	if hc.cipher != nil {
		switch c := hc.cipher.(type) {
		case cipher.Stream:
			c.XORKeyStream(payload, payload)
		case aead:
			explicitIVLen = c.explicitNonceLen()
			if len(payload) < explicitIVLen {
				return false, 0, alertBadRecordMAC
			}
			nonce := payload[:explicitIVLen]
			payload = payload[explicitIVLen:]
			if len(nonce) == 0 {
				nonce = hc.seq[:]
			}

			var additionalData [13]byte
			copy(additionalData[:], hc.seq[:])
			copy(additionalData[8:], b.data[:3])
			n := len(payload) - c.Overhead()
			additionalData[11] = byte(n >> 8)
			additionalData[12] = byte(n)
			var err error
			payload, err = c.Open(payload[:0], nonce, payload, additionalData[:])
			if err != nil {
				return false, 0, alertBadRecordMAC
			}
			b.resize(recordHeaderLen + explicitIVLen + len(payload))
		case cbcMode:
			blockSize := c.BlockSize()
			if hc.version >= VersionTLS11 {
				explicitIVLen = blockSize
			}

			if len(payload)%blockSize != 0 || len(payload) < roundUp(explicitIVLen+macSize+1, blockSize) {
				return false, 0, alertBadRecordMAC
			}

			if explicitIVLen > 0 {
				c.SetIV(payload[:explicitIVLen])
				payload = payload[explicitIVLen:]
			}

			c.CryptBlocks(payload, payload)
			if hc.version == VersionSSL30 {
				payload, paddingGood = removePaddinSSL30(payload)
			} else {
				payload, paddingGood = removePadding(payload)
			}
			b.resize(recordHeaderLen + explicitIVLen + len(payload))
		default:
			panic("unknown cipher type")
		}
	}
	if hc.mac != nil {
		if len(payload) < macSize {
			return false, 0, alertBadRecordMAC
		}

		n := len(payload) - macSize
		b.data[3] = byte(n >> 8)
		b.data[4] = byte(n)
		b.resize(recordHeaderLen + explicitIVLen + n)
		remoteMac := payload[n:]
		localMac := hc.mac.MAC(hc.inDigestBuf, hc.seq[0:], b.data[:recordHeaderLen], payload[:n])

		if subtle.ConstantTimeCompare(localMac, remoteMac) != 1 || paddingGood != 255 {
			return false, 0, alertBadRecordMAC
		}
		hc.inDigestBuf = localMac
	}
	hc.incSeq()

	return true, recordHeaderLen + explicitIVLen, 0
}

func padToBlockSize(payload []byte, blockSize int) (prefix, finalBlock []byte) {
	overrun := len(payload) % blockSize
	paddingLen := blockSize - overrun
	prefix = payload[:len(payload)-overrun]
	finalBlock = make([]byte, blockSize)
	copy(finalBlock, payload[len(payload)-overrun:])
	for i := overrun; i < blockSize; i++ {
		finalBlock[i] = byte(paddingLen - 1)
	}

	return
}

func (hc *halfConn) encrypt(b *block, explicitIVLen int) (bool, alert) {
	if hc.mac != nil {
		mac := hc.mac.MAC(hc.outDigestBuf, hc.seq[0:], b.data[:recordHeaderLen], b.data[recordHeaderLen+explicitIVLen:])
		n := len(b.data)
		b.resize(n + len(mac))
		copy(b.data[n:], mac)
		hc.outDigestBuf = mac
	}

	payload := b.data[recordHeaderLen:]

	if hc.cipher != nil {
		switch c := hc.cipher.(type) {
		case cipher.Stream:
			c.XORKeyStream(payload, payload)
		case aead:
			payloadLen := len(b.data) - recordHeaderLen - explicitIVLen
			b.resize(len(b.data) + c.Overhead())
			nonce := b.data[recordHeaderLen : recordHeaderLen+explicitIVLen]
			if len(nonce) == 0 {
				nonce = hc.seq[:]
			}
			payload := b.data[recordHeaderLen+explicitIVLen:]
			payload = payload[:payloadLen]

			var additionalData [13]byte
			copy(additionalData[:], hc.seq[:])
			copy(additionalData[:], b.data[:3])
			additionalData[11] = byte(payloadLen >> 8)
			additionalData[12] = byte(payloadLen)

			c.Seal(payload[:0], nonce, payload, additionalData[:])
		case cbcMode:
			blockSize := c.BlockSize()
			if blockSize > 0 {
				c.SetIV(payload[:explicitIVLen])
				payload = payload[explicitIVLen:]
			}
			prefix, finalBlock := padToBlockSize(payload, blockSize)
			c.CryptBlocks(b.data[recordHeaderLen+explicitIVLen:], prefix)
			c.CryptBlocks(b.data[recordHeaderLen+explicitIVLen+len(prefix):], finalBlock)
		default:
			panic("unknown cipher type")
		}
	}

	n := len(b.data) - recordHeaderLen
	b.data[3] = byte(n >> 8)
	b.data[4] = byte(n)
	hc.incSeq()

	return true, 0
}

type block struct {
	data []byte
	off  int
	link *block
}

func (b *block) resize(n int) {
	if n > cap(b.data) {
		b.reserve(n)
	}

	b.data = b.data[0:n]
}

func (b *block) reserve(n int) {
	if cap(b.data) >= n {
		return
	}

	m := cap(b.data)
	if m == 0 {
		m = 1024
	}
	for m < n {
		m *= 2
	}
	data := make([]byte, len(b.data), m)
	copy(data, b.data)
	b.data = data
}

func (b *block) readFromUntil(c *Conn, n int) error {
	r := c.conn
	if len(b.data) >= n {
		return nil
	}

	b.reserve(n)
	for {
		m, err := r.Read(b.data[len(b.data):cap(b.data)])
		b.data = b.data[0 : len(b.data)+m]
		c.readFromUntilLen += m
		if len(b.data) >= n {
			break
		}
		if err != nil {
			return err
		}
	}

	return nil
}

func (b *block) Read(p []byte) (n int, err error) {
	n = copy(p, b.data[b.off:])
	b.off += n
	return
}

func (hc *halfConn) newBlock() *block {
	b := hc.bfree
	if b == nil {
		return new(block)
	}

	hc.bfree = b.link
	b.link = nil

	return b
}

func (hc *halfConn) freeBLock(b *block) {
	b.link = hc.bfree
	hc.bfree = b
}

func (hc *halfConn) splitBLock(b *block, n int) (*block, *block) {
	if len(b.data) <= n {
		return b, nil
	}

	bb := hc.newBlock()
	bb.resize(len(b.data) - n)
	copy(bb.data, b.data[n:])
	b.data = b.data[0:n]

	return b, bb
}

func convertSSLv2ClientHello(c *Conn, b *block) error {
	if (uint8(b.data[0]) & 128) != 128 {
		return c.sendAlert(alertUnexpectedMessage)
	}

	msgLength := (uint16(b.data[0]&0x7f) << 8) | uint16(b.data[1])
	if msgLength < 12 {
		return c.sendAlert(alertHandshakeFailure)
	}

	msgType := uint8(b.data[2])
	majorVer := uint8(b.data[3])
	minorVer := uint8(b.data[4])
	version := uint16(majorVer<<8) | uint16(minorVer)
	if !(msgType == typeClientHello && version >= VersionSSL30) {
		c.sendAlert(alertProtocolVersion)
		state.TlsHandshakeSslv2NotSupport.Inc(1)
		return c.in.setErrorLocked(errors.New("tls: unsupported SSLv2 handshake received"))
	}
	state.TlsHandshakeAcceptSslv2ClientHello.Inc(1)

	if err := b.readFromUntil(c, int(2+msgLength)); err != nil {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		if e, ok := err.(net.Error); !ok || !e.Temporary() {
			c.in.setErrorLocked(err)
		}

		return err
	}

	cipherSpecLength := uint16(b.data[5])<<8 | uint16(b.data[6])
	if cipherSpecLength <= 0 || (cipherSpecLength%3) != 0 {
		return c.sendAlert(alertHandshakeFailure)
	}

	sessionIdLength := uint16(b.data[7])<<8 | uint16(b.data[8])
	if sessionIdLength != 0 && sessionIdLength != 16 {
		return c.sendAlert(alertHandshakeFailure)
	}

	if len(b.data) < 11+int(cipherSpecLength+sessionIdLength) {
		return c.sendAlert(alertHandshakeFailure)
	}

	cipherSpecs := b.data[11 : 11+cipherSpecLength]
	challengeData := b.data[11+cipherSpecLength+sessionIdLength:]

	b, c.rawInput = c.in.splitBLock(b, int(2+msgLength))
	b.off = 2

	helloMsg := clientHelloMsg{
		vers:               version,
		sessionId:          []byte{0},
		compressionMethods: []uint8{compressionNone},
	}

	if len(challengeData) >= 32 {
		helloMsg.random = challengeData[:32]
	} else {
		helloMsg.random = make([]byte, 32-len(challengeData))
		helloMsg.random = append(helloMsg.random, challengeData...)
	}

	helloMsg.cipherSuites = make([]uint16, 0)
	for i := 0; i < len(cipherSpecs); i += 3 {
		if cipherSpecs[i] == 0 {
			cipher := uint16(cipherSpecs[i+1])<<8 | uint16(cipherSpecs[i+2])
			helloMsg.cipherSuites = append(helloMsg.cipherSuites, cipher)
		}
	}

	c.hand.Write(helloMsg.marshal())

	c.sslv2Data = b.data[2:]
	c.in.freeBLock(b)

	return nil
}

func (c *Conn) SendAlertLocked(err alert) error {
	switch err {
	case alertNoRenegotiation, alertCloseNotify:
		c.tmp[0] = alertLevelWarning
	default:
		c.tmp[0] = alertLevelError
	}
	c.tmp[1] = byte(err)
	c.writeRecord(recordTypeAlert, c.tmp[0:2])

	if err != alertCloseNotify {
		return c.out.setErrorLocked(&net.OpError{
			Op:  "local error",
			Err: err,
		})
	}

	return nil
}

func (c *Conn) sendAlert(err alert) error {
	c.out.Lock()
	defer c.out.Unlock()

	return c.SendAlertLocked(err)
}

func (c *Conn) choosePlaintextSize() int {
	if !c.enableDynamicRecord {
		return initPlaintext
	}

	if c.byteOut < bytesThreshould {
		return initPlaintext
	}

	if time.Since(c.lastOut) < inactiveSeconds {
		return maxPlaintext
	}

	c.byteOut = 0

	return initPlaintext
}

func (c *Conn) writeRecord(typ recordType, data []byte) (n int, err error) {
	plaintextSize := maxPlaintext
	if typ == recordTypeApplicationData {
		plaintextSize = c.choosePlaintextSize()
	}

	b := c.out.newBlock()
	for len(data) > 0 {
		m := len(data)
		if m > plaintextSize {
			m = plaintextSize
		}
		explicitIVLen := 0
		explicitIVIsSeq := false

		var cbc cbcMode
		if c.out.version >= VersionTLS11 {
			var ok bool
			if cbc, ok = c.out.cipher.(cbcMode); ok {
				explicitIVLen = cbc.BlockSize()
			}
		}
		if explicitIVLen == 0 {
			if c, ok := c.out.cipher.(aead); ok {
				explicitIVLen = c.explicitNonceLen()
				explicitIVIsSeq = explicitIVLen > 0
			}
		}
		b.resize(recordHeaderLen + explicitIVLen + m)
		b.data[0] = byte(typ)
		vers := c.vers
		if vers == 0 {
			vers = VersionTLS10
		}
		b.data[1] = byte(vers >> 8)
		b.data[2] = byte(vers)
		b.data[3] = byte(m >> 8)
		b.data[4] = byte(m)
		if explicitIVLen > 0 {
			explicitIV := b.data[recordHeaderLen : recordHeaderLen+explicitIVLen]
			if explicitIVIsSeq {
				copy(explicitIV, c.out.seq[:])
			} else {
				if _, err = io.ReadFull(c.config.rand(), explicitIV); err != nil {
					break
				}
			}
		}
		copy(b.data[recordHeaderLen+explicitIVLen:], data)
		c.out.encrypt(b, explicitIVLen)
		_, err = c.conn.Write(b.data)
		if err != nil {
			break
		}
		n += m
		data = data[m:]
	}
	c.out.freeBLock(b)

	if typ == recordTypeChangeCipherSpec {
		err = c.out.changeCipherSpec()
		if err != nil {
			c.tmp[0] = alertLevelError
			c.tmp[1] = byte(err.(alert))
			c.writeRecord(recordTypeAlert, c.tmp[0:2])
			return n, c.out.setErrorLocked(&net.OpError{
				Op:  "local error",
				Err: err,
			})
		}
	}

	if typ == recordTypeApplicationData {
		c.byteOut += n
		c.lastOut = time.Now()
	}

	return
}
