# crypto

Practical building blocks when you already know you need **AES-CBC**, **PBKDF2**, and **HMAC**, plus small utilities for **random bytes** and **hash algorithms** in one place.

**Module:** [`github.com/voidbear-io/go/crypto`](https://github.com/voidbear-io/go/tree/main/crypto) · **Go:** 1.22+

## Naming the import

Go’s standard library already has a package named `crypto`. If your file imports both, give this module an alias so the compiler (and your readers) are never confused:

```go
import vcrypto "github.com/voidbear-io/go/crypto"
```

Subpackages (`aescbc`, `hashes`) use their own names and do not clash.

## Install

```bash
go get github.com/voidbear-io/go/crypto
```

## What you get

| Package | In plain terms |
|--------|----------------|
| **Root** (`crypto`) | **`RandBytes`** for cryptographically random slices, plus **`SymmetricCipher`** if you want a shared interface shape. |
| **`hashes`** | Named hash types (MD5 through SHA-3 and Blake2), **`HashNew`**, **`NewHmac`**, and numeric **IDs** for formats that store a hash kind in a header. |
| **`aescbc`** | **`AesCBC`** — PBKDF2 stretches your passphrase, AES-CBC encrypts the payload with PKCS#7 padding, then **encrypt-then-MAC** (HMAC over the ciphertext) checks integrity on decrypt. **`New128`** / **`New256`** are ready-made defaults. |

## Quick start (encrypt and decrypt)

```go
package main

import (
	"fmt"
	"log"

	"github.com/voidbear-io/go/crypto/aescbc"
)

func main() {
	c := aescbc.New256()
	key := []byte("0123456789abcdef0123456789abcdef") // 32 bytes for AES-256
	plain := []byte("hello, secret")

	out, err := c.Encrypt(key, plain)
	if err != nil {
		log.Fatal(err)
	}

	back, err := c.Decrypt(key, out)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(back)) // hello, secret
}
```

You can swap KDF/HMAC algorithms via **`KdfHash`** and **`HmacHash`** on **`AesCBC`** (see tests for Blake2b examples).

## Before you ship

This library is meant as **focused helpers**, not a full threat model write-up. Prefer **AEAD** (for example **AES-GCM** from `crypto/cipher`) or a high-level format like **age** when you want modern defaults and fewer foot-guns. Use strong, random keys in production; the sample key above is only for illustration.

## License

[MIT](LICENSE.md)
