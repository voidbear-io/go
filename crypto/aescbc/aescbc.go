package aescbc

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"encoding/binary"
	"fmt"

	"github.com/voidbear-io/go/crypto"
	"github.com/voidbear-io/go/crypto/hashes"
	"golang.org/x/crypto/pbkdf2"
)

// AesCBC implements AES encryption in CBC mode with PKCS#7 padding.
// It supports key derivation using PBKDF2 to generate the actual symmetric key
// for aes and the HMAC key for integrity verification.
//
// An encrypt-then-mac approach is used, where the data is encrypted first,
// and then an HMAC is computed over the ciphertext to ensure integrity.
type AesCBC struct {
	Iterations  int32
	KeySize     int
	Version     int16
	KdfSaltSize int16
	KdfHash     hashes.HashType
	HmacHash    hashes.HashType
}

func New256() *AesCBC {
	return &AesCBC{
		Iterations:  60000,
		KeySize:     32,
		Version:     1,
		KdfSaltSize: 8,
		KdfHash:     hashes.SHA256,
		HmacHash:    hashes.SHA256,
	}
}

func New128() *AesCBC {
	return &AesCBC{
		Iterations:  60000,
		KeySize:     16,
		Version:     1,
		KdfSaltSize: 8,
		KdfHash:     hashes.SHA256,
		HmacHash:    hashes.SHA256,
	}
}

func (a *AesCBC) Encrypt(key []byte, data []byte) (encryptedData []byte, err error) {
	return a.EncryptWithMetadata(key, data, nil)
}

func (a *AesCBC) EncryptWithMetadata(key []byte, data []byte, metadata []byte) (encryptedData []byte, err error) {
	if a.Version != 1 {
		return nil, fmt.Errorf("unsupported version: %d", a.Version)
	}

	saltSize := a.KdfSaltSize
	keySizeField := int16(a.KeySize)

	buf := new(bytes.Buffer)
	err = binary.Write(buf, binary.LittleEndian, a.Version)
	if err != nil {
		return nil, err
	}

	err = binary.Write(buf, binary.LittleEndian, a.KdfSaltSize)
	if err != nil {
		return nil, err
	}

	err = binary.Write(buf, binary.LittleEndian, keySizeField)
	if err != nil {
		return nil, err
	}

	err = binary.Write(buf, binary.LittleEndian, a.KdfHash.Id())
	if err != nil {
		return nil, err
	}

	err = binary.Write(buf, binary.LittleEndian, a.HmacHash.Id())
	if err != nil {
		return nil, err
	}

	err = binary.Write(buf, binary.LittleEndian, a.Iterations)
	if err != nil {
		return nil, err
	}

	metadataSize := int32(len(metadata))
	err = binary.Write(buf, binary.LittleEndian, metadataSize)
	if err != nil {
		return nil, err
	}

	salt, err := crypto.RandBytes(int(saltSize))
	if err != nil {
		return nil, err
	}
	err = binary.Write(buf, binary.LittleEndian, salt)
	if err != nil {
		return nil, err
	}

	iv, err := crypto.RandBytes(16)
	if err != nil {
		return nil, err
	}

	err = binary.Write(buf, binary.LittleEndian, iv)
	if err != nil {
		return nil, err
	}

	if metadataSize > 0 {
		buf.Write(metadata)
	}

	cdr := pbkdf2.Key(key, salt, int(a.Iterations), a.KeySize, a.KdfHash.HashNew())
	paddedData := pad(data)
	ciphertext := make([]byte, len(paddedData))
	c, err := aes.NewCipher(cdr)
	if err != nil {
		return nil, err
	}
	ctr := cipher.NewCBCEncrypter(c, iv)
	ctr.CryptBlocks(ciphertext, paddedData)

	h := a.HmacHash.NewHmac(cdr)
	if h == nil {
		return nil, fmt.Errorf("unsupported HMAC hash: %s", a.HmacHash)
	}
	h.Write(ciphertext)
	tag := h.Sum(nil)

	bufLen := buf.Len()
	tagLen := len(tag)
	ciphertextLen := len(ciphertext)

	result := make([]byte, bufLen+tagLen+ciphertextLen)
	copy(result, buf.Bytes())
	copy(result[bufLen:], tag)
	copy(result[bufLen+tagLen:], ciphertext)

	return result, nil
}

func pad(in []byte) []byte {
	out := make([]byte, len(in))
	copy(out, in)
	padding := 16 - (len(out) % 16)
	for i := 0; i < padding; i++ {
		out = append(out, byte(padding))
	}
	return out
}

func unpad(in []byte) []byte {
	if len(in) == 0 {
		return nil
	}

	padding := in[len(in)-1]
	if int(padding) > len(in) || padding > aes.BlockSize {
		return nil
	} else if padding == 0 {
		return nil
	}

	for i := len(in) - 1; i > len(in)-int(padding)-1; i-- {
		if in[i] != padding {
			return nil
		}
	}
	return in[:len(in)-int(padding)]
}

func (a *AesCBC) Decrypt(key []byte, encryptedData []byte) (data []byte, err error) {
	decryptedData, _, err := a.DecryptWithMetadata(key, encryptedData)
	return decryptedData, err
}

func (a *AesCBC) DecryptWithMetadata(key []byte, encryptedData []byte) (data []byte, metadata []byte, err error) {
	var version int16
	reader := bytes.NewReader(encryptedData)
	err = binary.Read(reader, binary.LittleEndian, &version)
	if err != nil {
		return nil, nil, err
	}

	if version != a.Version {
		return nil, nil, fmt.Errorf("invalid version %d for AesCBC", version)
	}

	var saltSize int16
	err = binary.Read(reader, binary.LittleEndian, &saltSize)
	if err != nil {
		return nil, nil, err
	}

	var keySizeShort int16
	err = binary.Read(reader, binary.LittleEndian, &keySizeShort)
	if err != nil {
		return nil, nil, err
	}
	_ = keySizeShort // stored for format compatibility; derived key length uses a.KeySize

	var kdfHashId int16
	err = binary.Read(reader, binary.LittleEndian, &kdfHashId)
	if err != nil {
		return nil, nil, err
	}

	var hmacHashId int16
	err = binary.Read(reader, binary.LittleEndian, &hmacHashId)
	if err != nil {
		return nil, nil, err
	}

	var iterations int32
	err = binary.Read(reader, binary.LittleEndian, &iterations)
	if err != nil {
		return nil, nil, err
	}

	var metadataSize int32
	err = binary.Read(reader, binary.LittleEndian, &metadataSize)
	if err != nil {
		return nil, nil, err
	}

	sliceStart := 18

	salt := encryptedData[sliceStart : sliceStart+int(saltSize)]
	sliceStart += int(saltSize)

	iv := encryptedData[sliceStart : sliceStart+16]
	sliceStart += 16

	if metadataSize > 0 {
		metadata = encryptedData[sliceStart : sliceStart+int(metadataSize)]
		sliceStart += int(metadataSize)
	}

	kdfType := hashes.FromId(kdfHashId)
	hmacType := hashes.FromId(hmacHashId)
	if !kdfType.IsValid() || !hmacType.IsValid() {
		return nil, nil, fmt.Errorf("invalid hash id in payload (kdf=%d hmac=%d)", kdfHashId, hmacHashId)
	}

	tagSize := hmacType.Size()
	if tagSize <= 0 || sliceStart+tagSize > len(encryptedData) {
		return nil, nil, fmt.Errorf("invalid tag size or truncated payload")
	}

	tag := encryptedData[sliceStart : sliceStart+tagSize]
	sliceStart += len(tag)

	ciphertext := encryptedData[sliceStart:]

	deriveLen := a.KeySize
	if keySizeShort > 0 {
		deriveLen = int(keySizeShort)
	}
	cdr := pbkdf2.Key(key, salt, int(iterations), deriveLen, kdfType.HashNew())
	h := hmacType.NewHmac(cdr)
	if h == nil {
		return nil, nil, fmt.Errorf("unsupported HMAC hash in payload")
	}
	h.Write(ciphertext)
	expectedTag := h.Sum(nil)

	if !hmac.Equal(tag, expectedTag) {
		return nil, nil, fmt.Errorf("hash mismatch")
	}

	c, err := aes.NewCipher(cdr)
	if err != nil {
		return nil, nil, err
	}
	ctr := cipher.NewCBCDecrypter(c, iv)
	plaintext := make([]byte, len(ciphertext))
	ctr.CryptBlocks(plaintext, ciphertext)

	return unpad(plaintext), metadata, nil
}
