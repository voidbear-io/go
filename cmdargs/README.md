# cmdargs

Small helpers for working with command-line argument slices: parse strings like a shell, quote them back for the current OS, and read or mutate tokens without ad-hoc string handling.

**Module:** `github.com/voidbear-io/go/cmdargs` · **Go:** 1.18+

## Install

```bash
go get github.com/voidbear-io/go/cmdargs
```

## What you get

- **`Args`** — Wrapper around `[]string` with copy-safe `ToArray()`, `Len()`, `Get(i)`, and a fluent `String()` that re-encodes arguments for the platform (Unix vs Windows quoting rules via `AppendCliArg`).
- **Parsing** — `Split` turns one string into tokens (quotes, spaces, backticks, line continuations). `SplitAndExpand` does the same but runs `$…` expansion inside double-quoted segments using your callback.
- **Reading** — Typed getters by index (`GetInt`, `GetBool`, …) and by first matching token (`GetAny`, `GetIntAny`, …). Repeated flags: `GetSlice` / `GetSliceAny` (`-e a -e b`), and `GetMap` / `GetMapAny` (`-D k=v` with `key=value` pairs).
- **Search** — `Index` / `Contains` (case-insensitive) vs `IndexAny` / `ContainsAny` (exact token match for the needles you pass).
- **Mutation** — `Push`, `Append`, `Prepend`, `Shift`, `Pop`, `Remove`, `RemoveAt`, `Set`, and option-oriented `SetValue` / `SetInt` / `SetFloat` / `SetBool`.

## Quick start

```go
package main

import (
	"fmt"
	"os"

	"github.com/voidbear-io/go/cmdargs"
)

func main() {
	// From os.Args (skip program name)
	a := cmdargs.New(os.Args[1:])

	// Or parse a command line
	line := cmdargs.Split(`git commit -m "first release"`)
	fmt.Println(line.ToArray()) // [git commit -m first release]

	// Repeated -e flags → []string
	envs, ok := cmdargs.New([]string{"run", "-e", "a", "-e", "b"}).GetSlice("-e")
	if ok {
		fmt.Println(envs) // [a b]
	}

	// Repeated -D key=value → map (last key wins)
	defs, ok := cmdargs.New([]string{"-D", "os=linux", "-D", "arch=arm64"}).GetMap("-D")
	if ok {
		fmt.Println(defs["os"])
	}

	// Serialize for a shell command (quoting depends on GOOS)
	fmt.Println(a.String())
}
```

## Parsing

`Split` only consumes the **first line** of input unless newlines are escaped (e.g. `foo \` newline `bar`) or appear inside quotes. It understands single- and double-quoted runs and common continuation patterns used in shell and PowerShell-style lines.

`SplitAndExpand` is for double-quoted segments that may contain `$VAR` (or your own syntax): you supply `expand func(string) (string, error)`; expansion errors abort parsing.

## Matching rules (important)

| API | How option / needle tokens are matched |
|-----|----------------------------------------|
| `Index`, `SetValue`, `SetInt`, … | Case-insensitive (`EqualFold`) |
| `IndexAny`, `GetAny`, `ContainsAny`, … | Exact string match to your literals |
| `GetSlice`, `GetMap`, `GetSliceAny`, `GetMapAny` | Option name: case-insensitive (`EqualFold`) |

So `-e` and `-E` are the same flag for slice/map collection, but `IndexAny([]string{"-e"})` only matches `-e`, not `-E`.

## Repeated options

- **`GetSlice` / `GetSliceAny`** — Each time an option appears, the **next** argument is appended. Values are separate argv elements (`-e first -e second`), not `-e=first`.
- **`GetMap` / `GetMapAny`** — Each value must contain `=`; the first `=` splits key and value. Tokens without `=` are skipped. Duplicate keys: **last** occurrence wins.

## Formatting

- **`(*Args).String()`** — Joins arguments with spaces using `AppendCliArg` per token.
- **`AppendCliArg(*strings.Builder, string)`** — Escape/quote a single argument; behavior differs on Windows vs Unix builds (`go:build`).

## License

See the repository root for license terms.
