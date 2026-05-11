# env

Work with environment variables in Go without re‑implementing shell rules. Read and write values, tweak `PATH`, and expand strings that look like shell snippets (`${VAR}`, defaults, optional `$(commands)`, and more).

**Module:** [`github.com/voidbear-io/go/env`](https://github.com/voidbear-io/go/env) · **Go:** 1.18+

## Install

```bash
go get github.com/voidbear-io/go/env@v0.0.0-alpha.0
```

If you are hacking in this repo, `env/go.mod` uses a `replace` so the local `cmdargs` sibling is used automatically.

## What you get

| Area | What it does |
|------|----------------|
| **Basics** | `Get`, `Set`, `Unset`, `Has`, and `All` for normal env access. |
| **PATH** | `GetPath`, `SetPath`, `SplitPath`, `JoinPath`, `PrependPath`, `AppendPath`, `HasPath` — with sensible behavior on Unix vs Windows. |
| **Expand** | Turn a template string into a final string: `${NAME}`, `${NAME:-default}`, `${NAME:=setdefault}`, `${NAME:?error}`, `$NAME`, `%NAME%` on Windows, and optional `$(…)` command substitution. |
| **Hooks** | `WithGet` / `WithSet` for tests or overlays; `WithCustomExpander` to rewrite the **whole** expanded string (for example, turn `akv://vault/key` into a real secret). |

`Expand` uses [`github.com/voidbear-io/go/cmdargs`](https://github.com/voidbear-io/go/tree/main/cmdargs) to parse command lines when substitution runs a command without going through a shell.

## Quick start: read and write

```go
package main

import (
	"fmt"
	"log"

	"github.com/voidbear-io/go/env"
)

func main() {
	if err := env.Set("GREETING", "hello"); err != nil {
		log.Fatal(err)
	}
	fmt.Println(env.Get("GREETING")) // hello
	fmt.Println(env.Has("MISSING"))  // false
}
```

## Quick start: expansion

Defaults and simple interpolation:

```go
out, err := env.Expand("Hello, ${USER:-friend}!")
if err != nil {
	log.Fatal(err)
}
// Uses $USER from the environment, or the text "friend" if it is empty.
```

Use options when you need control (tests, fake env, Windows‑style vars, or custom secret handling):

```go
package main

import (
	"log"
	"strings"

	"github.com/voidbear-io/go/env"
)

func example() {
	out, err := env.Expand(
		"Secret: ${VAULT_REF}",
		env.WithGet(func(key string) string {
			if key == "VAULT_REF" {
				return "akv://my-vault/my-key"
			}
			return ""
		}),
		env.WithCustomExpander(func(s string) (string, error) {
			// After normal expansion, rewrite any remaining scheme you care about.
			return strings.ReplaceAll(s, "akv://my-vault/my-key", "pancakes"), nil
		}),
	)
	if err != nil {
		log.Fatal(err)
	}
	_ = out
}
```

The custom step sees the **entire** result of expansion (both substituted parts and literal text), so you can scan once for patterns like custom URI schemes.

## Learn more

The tests in [`expand_test.go`](expand_test.go) are short examples you can copy from: defaults, nesting, command substitution, Windows `%VAR%`, and error cases.

## License

MIT — see [`LICENSE.md`](LICENSE.md).
