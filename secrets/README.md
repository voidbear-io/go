# secrets

Generate **cryptographically random** strings (passwords, tokens) with configurable character classes, and **mask** known secret substrings in text so logs stay safe.

**Module:** [`github.com/voidbear-io/go/secrets`](https://github.com/voidbear-io/go/tree/main/secrets) · **Go:** 1.18+

## Install

```bash
go get github.com/voidbear-io/go/secrets@v0.0.0-alpha.0
```

## Quick start

```go
package main

import (
	"fmt"
	"log"

	"github.com/voidbear-io/go/secrets"
)

func main() {
	token, err := secrets.Generate(32, secrets.WithSymbols("!@#"))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(token)

	m := secrets.NewSecretMasker()
	m.AddValue("sk-live-abc123")
	fmt.Println(m.Mask(`msg="sk-live-abc123"`))
}
```

## Features

- **`Generate` / `GenerateRunes` / `GenerateBytes`** — Uses `crypto/rand`; options for lower/upper/digits/symbols, custom alphabet (`WithChars`), optional validator, retries.
- **`SecretMasker`** — Register literal secrets (plus optional **generators** that derive variants); **`Mask`** replaces matches case-insensitively with `****`. **`DefaultMasker`** is a package-level singleton.

## License

MIT — see [`LICENSE.md`](LICENSE.md).
