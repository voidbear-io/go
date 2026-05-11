// Package crypto defines shared crypto helpers (random bytes, hash registry) and
// optional symmetric cipher interfaces. Import as a named path, e.g.
//
//	import vcrypto "github.com/voidbear-io/go/crypto"
//
// if you also use the standard library crypto package in the same file.
package crypto

type SymmetricCipher interface {
	Encrypt(key []byte, data []byte) (encryptedData []byte, err error)

	EncryptWithMetadata(key []byte, data []byte, metadata []byte) (encryptedData []byte, err error)

	Decrypt(key []byte, encryptedData []byte) (data []byte, err error)

	DecryptWithMetadata(key []byte, encryptedData []byte) (data []byte, metadata []byte, err error)
}
