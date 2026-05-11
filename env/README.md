# env

Helpers for environment variables and shell-style expansion in Go: `Get` / `Set` / `PATH` helpers, plus `Expand` for `${VAR}`, defaults, optional Windows `%VAR%`, and `$(command)` substitution (direct execution or shell).

**Module:** `github.com/voidbear-io/go/env` · **Go:** 1.18+

Depends on [`github.com/voidbear-io/go/cmdargs`](https://github.com/voidbear-io/go/tree/main/cmdargs) for parsing command substitutions when shell expansion is disabled.

## Install

```bash
go get github.com/voidbear-io/go/env@v0.0.0-alpha.0
```

In this monorepo, `env/go.mod` uses a `replace` directive so the local `cmdargs` module is used until both modules are published.

## Usage

```go
import "github.com/voidbear-io/go/env"

func main() {
	_ = env.Set("FOO", "bar")
	out, err := env.Expand("${FOO:-default}")
	if err != nil {
		panic(err)
	}
	// out == "bar"
}
```

See tests in `expand_test.go` for substitution, defaults, and command substitution options.
