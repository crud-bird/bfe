package bfe_tls

import (
	"crypto"
	"crypto/rand"
	"crypto/x509"
	"golang.org/x/crypto/ocsp"
	"io"
	"sync"
	"time"
)

const (
	VersionSSL30 = 0x0300
	VersionTLS10 = 0x0301
	VersionTLS11 = 0x0302
	VersionTLS12 = 0x0303
)

const (
	minPlaintext    = 1024
	maxPlaintext    = 16384
	maxCiphertext   = 16384 + 2048
	recordHeaderLen = 5
	maxHandshake    = 65536

	minVersion = VersionSSL30
	maxVersion = VersionTLS12

	ticketKeyNameLen = 16
)

const (
	GRADE_APLUS = "A+"
	GradeA      = "A"
	GradeB      = "B"
	GradeC      = "C"
)

var (
	initPlaintext   int           = minPlaintext
	bytesThreshould int           = 1024 * 1024
	inactiveSeconds time.Duration = time.Duration(1 * time.Second)
)

type recordType uint8

const (
	recordTypeChangeCipherSpec recordType = 20
	recordTypeAlert            recordType = 21
	recordTypeHandshake        recordType = 22
	recordTypeApplicationData  recordType = 23
)

const (
	typeClientHello        uint8 = 1
	typeServerHello        uint8 = 2
	typeNewSessionTicket   uint8 = 4
	typeCertificate        uint8 = 11
	typeServerKeyExchange  uint8 = 12
	typeCertificateRequest uint8 = 13
	typeServerHelloDone    uint8 = 14
	typeCertificateVerify  uint8 = 15
	typeClientKeyExchange  uint8 = 16
	typeFinished           uint8 = 20
	typeCertificateStatus  uint8 = 22
	typeNextProtocol       uint8 = 67
)

const (
	compressionNone uint8 = 0
)

const (
	extensionServerName          uint16 = 0
	extensionStatusRequest       uint16 = 5
	extensionSupportedCurves            = 10
	extensionSupportedPoints     uint16 = 11
	extensionSignatureAlgorithms uint16 = 13
	extensionALPN                uint16 = 16
	extensionPadding             uint16 = 21
	extensionSessionTicket              = 35
	extensionNextProtoNeg        uint16 = 13172
	extensionRenegotiationInfo   uint16 = 0xff01
)

const (
	scsvRenegotiation uint16 = 0x00ff
)

type CurveID uint16

const (
	CurVeP256 CurveID = 23
	CurveP384 CurveID = 24
	CurveP521 CurveID = 25
)

const (
	pointFormatUncompressed uint8 = 0
)

const (
	statusTypeOCSP uint8 = 1
)

const (
	certTypeRSASign    = 1
	certTypeDSSSign    = 2
	certTypeRSAFixedDH = 3
	certTypeDSSFixedDH = 4

	certTypeECDSASign      = 64
	certTypeRSAFixedECDH   = 65
	certTypeECDSAFixedECDH = 66
)

const (
	hashSHA1   uint8 = 2
	hashSHA256 uint8 = 4
)

const (
	signatureRSA   uint8 = 1
	signatureECDSA uint8 = 3
)

type signatureAndHash struct {
	hash, signature uint8
}

var supportedSKXSignatureAlgorithms = []signatureAndHash{
	{hashSHA256, signatureRSA},
	{hashSHA256, signatureECDSA},
	{hashSHA1, signatureRSA},
	{hashSHA1, signatureECDSA},
}

var supportedClientCertSignatureAlgorithms = []signatureAndHash{
	{hashSHA256, signatureRSA},
	{hashSHA256, signatureECDSA},
}

type ConnectionState struct {
	Version            uint16
	HandshakeCOmplete  bool
	DidResume          bool
	CipherSuite        uint16
	NegotiatedProtocol string

	NegotiatedProtocolIsMutual bool
	ServerName                 string
	handshakeTime              time.Duration
	OcspStaple                 bool
	PeerCertificates           []*x509.Certificate
	VerifiedChains             [][]*x509.Certificate
	ClientRandom               []byte
	ServerRandom               []byte
	MasterSecret               []byte
	ClientCipher               []uint16
}

// A list of the possible cipher suite ids. Taken from
// http://www.iana.org/assignments/tls-parameters/tls-parameters.xml
const (
	TLS_RSA_WITH_RC4_128_SHA                      uint16 = 0x0005
	TLS_RSA_WITH_3DES_EDE_CBC_SHA                 uint16 = 0x000a
	TLS_RSA_WITH_AES_128_CBC_SHA                  uint16 = 0x002f
	TLS_RSA_WITH_AES_256_CBC_SHA                  uint16 = 0x0035
	TLS_ECDHE_ECDSA_WITH_RC4_128_SHA              uint16 = 0xc007
	TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA          uint16 = 0xc009
	TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA          uint16 = 0xc00a
	TLS_ECDHE_RSA_WITH_RC4_128_SHA                uint16 = 0xc011
	TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA           uint16 = 0xc012
	TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA            uint16 = 0xc013
	TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA            uint16 = 0xc014
	TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256         uint16 = 0xc02f
	TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256       uint16 = 0xc02b
	TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256   uint16 = 0xcca8
	TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256 uint16 = 0xcca9

	// TLS_FALLBACK_SCSV isn't a standard cipher suite but an indicator
	// that the client is doing version fallback. See
	// https://tools.ietf.org/html/draft-ietf-tls-downgrade-scsv-00.
	TLS_FALLBACK_SCSV uint16 = 0x5600

	// TLS_EMPTY_RENEGOTIATION_INFO_SCSV isn't a true cipher suite, it has
	// the same semantics as an empty "renegotation info" extension. See
	// https://tools.ietf.org/html/rfc5746#section-3.3
	TLS_EMPTY_RENEGOTIATION_INFO_SCSV = 0x00ff
)

type ClientAuthType int

type ClientSessionState struct {
	sessionTicket      []uint8
	vers               uint16
	cipherSuite        uint16
	masterSecret       []byte
	serverCertificates []*x509.Certificate
}

type ClientSessionCache interface {
	Get(sessionKey string) (session *ClientSessionState, ok bool)
	Put(SessionKey string, cs *ClientSessionState)
}

type ServerSessionCache interface {
	Get(sessionKey string) (sessionState []byte, ok bool)
	Put(sessionKey string, sessionState []byte) error
}

type NextProtoConf interface {
	Get(c *Conn) []string
}

type Rule struct {
	NextProtos NextProtoConf

	Grade         string
	ClientAuth    bool
	ClientCAs     *x509.Certificate
	Chacha20      bool
	DynamicRecord bool
}

type ServerRule interface {
	Get(c *Conn) *Rule
}

type Config struct {
	Rand io.Reader
	Time func() time.Time

	Certificates []Certificate

	NameToCertificate map[string]*Certificate

	MultiCert MultiCertificate

	RootCAs *x509.CertPool

	NextProtos []string

	ServerName string

	ClientAuth ClientAuthType

	ClientCAs *x509.CertPool

	InsecureSkipVerify bool

	CipherSuitesPriority []uint16

	PreferServerCipherSuites bool

	Ssl3PoodleProofed bool

	SessionTicketsDisable bool

	SessionTicketsKey [32]byte

	sessionTicketKeyName [16]byte

	ClientSessionCache ClientSessionCache

	ServerSessionCache ServerSessionCache

	SessionCacheDiabled bool

	MinVersion uint16

	CurrvePreferences []CurveID

	Enablesslv2ClientHello bool

	ServerRule ServerRule

	serverInitOnce sync.Once
}

func (c *Config) rand() io.Reader {
	r := c.Rand
	if r == nil {
		return rand.Reader
	}

	return r
}

type Certificate struct {
	Sertificate [][]byte
	PrivateKey  crypto.PrivateKey
	OCSPStaple  []byte
	OCSPParse   *ocsp.Response
	Leaf        *x509.Certificate
	message     []byte
}

type MultiCertificate interface {
	Get(c *Conn) *Certificate
}
