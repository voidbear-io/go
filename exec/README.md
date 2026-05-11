# exec

A thin layer on top of Go’s `os/exec` that fits real shell-style workflows: **one string** for a full command line (quotes and all), **`Result` objects** with exit code and captured I/O, **pipelines**, optional **logging**, and **`PATH` / registry lookups** wired to the sibling [`env`](https://github.com/voidbear-io/go/tree/main/env) helpers.

**Module:** [`github.com/voidbear-io/go/exec`](https://github.com/voidbear-io/go/tree/main/exec) · **Go:** 1.18+

## Import name

This package is named `exec`, like the standard library’s import path `os/exec`. If you use both in the same file, give this module an alias:

```go
import (
	osexec "os/exec"

	vbexec "github.com/voidbear-io/go/exec"
)
```

## Install

```bash
go get github.com/voidbear-io/go/exec@v0.0.0-alpha.0
```

(Or drop the version to track the latest pseudo-version from your proxy.)

## How this differs from `os/exec`

| Topic | `os/exec` | This module |
|--------|-----------|-------------|
| **Construction** | You split argv yourself: `exec.Command("git", "status", "-s")`. | You can still use `New("git", "status", "-s")`, or pass **one shell-like line**: `Command("git commit -m \"fix\"")` via [`cmdargs.Split`](https://github.com/voidbear-io/go/tree/main/cmdargs). |
| **Convenience runners** | `cmd.Output()` returns `[]byte` + `error`; `Run()` returns only `error`. | `Run("…")` / `Output("…")` return a **`Result`** (exit code, stdout, stderr, timing) so non-zero exits are visible without parsing `exec.ExitError`. |
| **Streaming** | First-class: you wire `Stdin`/`Stdout`/`Stderr` to pipes, files, or `io.Reader`/`Writer`. | **`Run`** and **`Output`** pick sensible defaults (inherit vs capture). Use embedded `*exec.Cmd` fields when you need full streaming control. |
| **Resolution** | `exec.LookPath` for a single executable name. | **`Which` / `WhichFirst`** walk `PATH` with optional cache and prepended dirs; **`Find`** uses a small **registry** and env overrides (see `finder.go`). |
| **Pipelines** | You connect `Stdout` of one cmd to `Stdin` of the next by hand. | **`Pipe` / `PipeCommand`** build a linear pipeline and run it. |

`os/exec` remains the right low-level API for maximum control. This package is for **ergonomics** and **consistent `Result` handling** when you would otherwise repeat the same glue code.

## Examples

### Build a command from a string (quoting handled for you)

```go
package main

import (
	"fmt"
	"log"

	vbexec "github.com/voidbear-io/go/exec"
)

func main() {
	res, err := vbexec.Command(`echo 'hello world'`).Output()
	if err != nil {
		log.Fatal(err)
	}
	if !res.IsOk() {
		log.Fatal(res.ToError())
	}
	fmt.Println(res.Text()) // hello world
}
```

### Program + args (same as `exec.Command`)

```go
import (
	"fmt"
	"log"
	"strings"

	vbexec "github.com/voidbear-io/go/exec"
)

res, err := vbexec.New("go", "version").Output()
if err != nil {
	log.Fatal(err)
}
fmt.Printf("exit=%d out=%q\n", res.Code, strings.TrimSpace(res.Text()))
```

### Run attached to the terminal (inherit stdin/stdout/stderr)

```go
_, err := vbexec.New("git", "status").Run()
// Interactive tools behave like running them in a shell.
```

### Capture quietly (discard child output)

```go
res, err := vbexec.New("sleep", "0").Quiet()
```

### Environment and working directory

```go
res, err := vbexec.New("pwd").
	WithCwd("/tmp").
	WithEnvMap(map[string]string{"FOO": "bar"}).
	Output()
```

### Pipeline: shell-style pipe

```go
res, err := vbexec.Command("echo 'Hello World'").PipeCommand("grep Hello").Output()
if err != nil {
	log.Fatal(err)
}
fmt.Println(strings.TrimSpace(res.Text())) // Hello World
```

### Result helpers

```go
res, err := vbexec.Output("some-cli --json")
if err != nil {
	log.Fatal(err)
}
if err := res.ToErrorIf(nil); err != nil { // non-zero exit → error
	log.Fatal(err)
}
for _, line := range res.Lines() {
	fmt.Println(line)
}
obj, err := res.Json() // decode stdout as JSON
_ = obj
```

### Optional global logger (debug)

```go
vbexec.SetLogger(func(c *vbexec.Cmd) {
	log.Printf("running: %v", c.Args)
})
```

### Resolve an executable on `PATH`

```go
if path, ok := vbexec.Which("git"); ok {
	fmt.Println("git is", path)
}
```

## License

[MIT](LICENSE.md)
