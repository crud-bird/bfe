package bfe_tls

import (
	"crypto/cipher"
)

type macFunction interface {
	Size() int
	MAC(digedtBuf, seq, header, data []byte) []byte
}

type aead interface {
	cipher.AEAD
	explicitNonceLen() int
}
