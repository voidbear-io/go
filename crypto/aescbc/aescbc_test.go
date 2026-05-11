package aescbc_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/voidbear-io/go/crypto/aescbc"
	"github.com/voidbear-io/go/crypto/hashes"
)

func Test256(t *testing.T) {
	cipher := aescbc.New256()

	plaintext := []byte("Hello, World!")
	key := []byte("0123456789abcdef0123456789abcdef")
	encrypted, err := cipher.Encrypt(key, plaintext)
	assert.NoError(t, err)
	assert.Greater(t, len(encrypted), len(plaintext))

	decrypted, err := cipher.Decrypt(key, encrypted)
	assert.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestBlake2b256(t *testing.T) {
	cipher := aescbc.New256()
	cipher.KdfHash = hashes.BLAKE2B_256
	cipher.HmacHash = hashes.BLAKE2B_256

	plaintext := []byte("Hello, World!")
	key := []byte("0123456789abcdef0123456789abcdef")
	encrypted, err := cipher.Encrypt(key, plaintext)
	assert.NoError(t, err)

	decrypted, err := cipher.Decrypt(key, encrypted)
	assert.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}
