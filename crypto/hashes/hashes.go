package hashes

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"hash"

	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/blake2s"
	"golang.org/x/crypto/sha3"
)

type HashType string

const (
	MD5         HashType = "MD5"
	SHA1        HashType = "SHA1"
	SHA224      HashType = "SHA224"
	SHA256      HashType = "SHA256"
	SHA384      HashType = "SHA384"
	SHA512      HashType = "SHA512"
	SHA3_224    HashType = "SHA3-224"
	SHA3_256    HashType = "SHA3-256"
	SHA3_384    HashType = "SHA3-384"
	SHA3_512    HashType = "SHA3-512"
	BLAKE2B_256 HashType = "BLAKE2B-256"
	BLAKE2B_384 HashType = "BLAKE2B-384"
	BLAKE2B_512 HashType = "BLAKE2B-512"
	BLAKE2S_128 HashType = "BLAKE2S-128"
	BLAKE2S_256 HashType = "BLAKE2S-256"
	Unknown     HashType = "Unknown" // Represents an unknown or unsupported hash type
)

func (h HashType) String() string {
	return string(h)
}

func FromId(id int16) HashType {
	switch id {
	case 1:
		return MD5
	case 2:
		return SHA1
	case 3:
		return SHA224
	case 4:
		return SHA256
	case 5:
		return SHA384
	case 6:
		return SHA512
	case 7:
		return SHA3_224
	case 8:
		return SHA3_256
	case 9:
		return SHA3_384
	case 10:
		return SHA3_512
	case 11:
		return BLAKE2B_256
	case 12:
		return BLAKE2B_384
	case 13:
		return BLAKE2B_512
	case 14:
		return BLAKE2S_128
	case 15:
		return BLAKE2S_256
	default:
		return Unknown // Represents an unknown or unsupported hash type
	}
}

func (h HashType) IsValid() bool {
	switch h {
	case MD5, SHA1, SHA224, SHA256, SHA384, SHA512,
		SHA3_224, SHA3_256, SHA3_384, SHA3_512,
		BLAKE2B_256, BLAKE2B_384, BLAKE2B_512,
		BLAKE2S_128, BLAKE2S_256:
		return true
	default:
		return false // Unknown hash type
	}
}

func (h HashType) IsUnknown() bool {
	return h == Unknown
}

func (h HashType) Id() int16 {
	switch h {
	case MD5:
		return 1
	case SHA1:
		return 2
	case SHA224:
		return 3
	case SHA256:
		return 4
	case SHA384:
		return 5
	case SHA512:
		return 6
	case SHA3_224:
		return 7
	case SHA3_256:
		return 8
	case SHA3_384:
		return 9
	case SHA3_512:
		return 10
	case BLAKE2B_256:
		return 11
	case BLAKE2B_384:
		return 12
	case BLAKE2B_512:
		return 13
	case BLAKE2S_128:
		return 14
	case BLAKE2S_256:
		return 15

	default:
		return -1 // Unknown hash type
	}
}

func (h HashType) Size() int {
	switch h {
	case MD5:
		return md5.Size
	case SHA1:
		return sha1.Size
	case SHA224:
		return sha256.Size224
	case SHA256:
		return sha256.Size
	case SHA384:
		return sha512.Size384
	case SHA512:
		return sha512.Size
	case SHA3_224:
		return 28
	case SHA3_256:
		return 32
	case SHA3_384:
		return 48
	case SHA3_512:
		return 64
	case BLAKE2B_256:
		return 32
	case BLAKE2B_384:
		return 48
	case BLAKE2B_512:
		return 64
	case BLAKE2S_128:
		return 16
	case BLAKE2S_256:
		return 32
	default:
		return 0 // Unknown hash type
	}
}

func (h HashType) HashNew() func() hash.Hash {
	switch h {
	case MD5:
		return md5.New
	case SHA1:
		return sha1.New
	case SHA224:
		return sha256.New224
	case SHA256:
		return sha256.New
	case SHA384:
		return sha512.New384
	case SHA512:
		return sha512.New
	case SHA3_224:
		return func() hash.Hash { return sha3.New224() }
	case SHA3_256:
		return func() hash.Hash { return sha3.New256() }
	case SHA3_384:
		return func() hash.Hash { return sha3.New384() }
	case SHA3_512:
		return func() hash.Hash { return sha3.New512() }
	case BLAKE2B_256:
		return func() hash.Hash {
			h, _ := blake2b.New256(nil) // Default to nil key
			return h
		}
	case BLAKE2B_384:
		return func() hash.Hash {
			h, _ := blake2b.New384(nil) // Default to nil key
			return h
		}
	case BLAKE2B_512:
		return func() hash.Hash {
			h, _ := blake2b.New512(nil) // Default to nil key
			return h
		}
	case BLAKE2S_128:
		return func() hash.Hash {
			h, _ := blake2s.New128(nil) // Default to nil key
			return h
		}
	case BLAKE2S_256:
		return func() hash.Hash {
			h, _ := blake2s.New256(nil) // Default to nil key
			return h
		}
	default:
		return nil // Unknown hash type
	}
}

func (h HashType) NewHmac(key []byte) hash.Hash {
	switch h {
	case MD5:
		return hmac.New(md5.New, key)
	case SHA1:
		return hmac.New(sha1.New, key)
	case SHA224:
		return hmac.New(sha256.New224, key)
	case SHA256:
		return hmac.New(sha256.New, key)
	case SHA384:
		return hmac.New(sha512.New384, key)
	case SHA512:
		return hmac.New(sha512.New, key)
	case SHA3_224:
		return hmac.New(func() hash.Hash { return sha3.New224() }, key)
	case SHA3_256:
		return hmac.New(func() hash.Hash { return sha3.New256() }, key)
	case SHA3_384:
		return hmac.New(func() hash.Hash { return sha3.New384() }, key)
	case SHA3_512:
		return hmac.New(func() hash.Hash { return sha3.New512() }, key)
	case BLAKE2B_256:
		return func() hash.Hash {
			h, _ := blake2b.New256(key)
			return h
		}()
	case BLAKE2B_384:
		return func() hash.Hash {
			h, _ := blake2b.New384(key)
			return h
		}()
	case BLAKE2B_512:
		return func() hash.Hash {
			h, _ := blake2b.New512(key)
			return h
		}()
	case BLAKE2S_128:
		return func() hash.Hash {
			h, _ := blake2s.New128(key)
			return h
		}()
	case BLAKE2S_256:
		return func() hash.Hash {
			h, _ := blake2s.New256(key)
			return h
		}()
	default:
		return nil // Unknown hash type
	}
}
