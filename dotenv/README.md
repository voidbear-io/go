# dotenv

Read and write **`.env`-style** files in Go without throwing away structure. This package keeps **order**, **comments**, and **how values were quoted**, so you can parse a file, change a few keys, and write it back without flattening everything into a map and losing `#` lines.

**Module:** [`github.com/voidbear-io/go/dotenv`](https://github.com/voidbear-io/go/tree/main/dotenv) · **Go:** 1.18+

## Install

```bash
go get github.com/voidbear-io/go/dotenv@v0.0.0-alpha.0
```

## What you get

| Area | What it does |
|------|----------------|
| **`Parse`** | Turn file text into an **`EnvDoc`**. Syntax errors come back as **`ParseError`** with **line and column**. |
| **`EnvDoc`** | Walk the file as **elements** (variables, comments, blank lines) or use **`Get` / `Set` / `Keys` / `ToMap` / `Merge`**. |
| **`String()`** | Serialize back to `.env`-like text, including comments and quoting hints where the doc tracks them. |
| **Building docs** | **`NewDoc`**, **`AddVariable`**, **`AddQuotedVariable`**, **`AddComment`**, **`AddNewline`** for programmatic edits. |

Need **`${VAR}`** or **`$(command)`** on values? Parse with **dotenv**, then run values through **[`env.Expand`](https://github.com/voidbear-io/go/tree/main/env)** from the sibling **`env`** module.

## Parsing behavior

This is what the **`Parse`** lexer does today (see **`parse.go`** and **`mod_test.go`** for edge cases).

### Lines and whitespace

- **Line endings:** Unix (`\n`), old Mac (`\r`), and Windows (`\r\n`) are treated as line breaks.
- **Outside quoted values:** Spaces and tabs are skipped where they separate tokens. For **unquoted** values, **trailing** spaces before the next line or `#` are trimmed from the stored value.
- **Blank lines** become newline elements in the document so round-trips can preserve spacing.

### Keys (`KEY=`)

- Names use **Unicode letters**, **Unicode digits**, and **underscore** (`_`). The first character must be a letter, digit, or `_` (digits are allowed at the start).
- **`=`** ends the name and starts the value. Extra spaces around `=` are allowed in the usual “loose” `.env` style (spaces after `=` are skipped before the value starts).
- A key split across lines by whitespace (key “terminated” then more characters) is a **parse error**.

### Comments

- **`#`** starts a comment **at the beginning of a line** (after optional whitespace) or continues a **whole-line** comment token.
- **`#` inside an unquoted value** ends the value and starts an **inline comment** (the `#` and rest of the line are not part of the value).
- After a **quoted** value closes, only **whitespace** and an optional **`# …`** comment are allowed before the line ends; anything else is an error.

Comment **text** stored in the doc does **not** include the leading `#` character (the writer adds `# ` when serializing).

### Quoted values

Values may be wrapped in any of:

| Quote | Role |
|--------|------|
| **`"`** double | C-style escapes (see below), `\"`, and newlines inside the string. |
| **`'`** single | Mostly **literal**. The only special case is **`\'`** for an embedded single quote. **`\\`** does **not** turn `\n` into a newline—backslashes are kept as-is except for that `\'` pair. |
| **`` ` ``** backtick | Like double quotes: **`` \` ``** for a literal backtick, otherwise **`\`** starts the same escape set as in double-quoted strings. |

If the first non-space character after `=` is **not** a quote, the value is **unquoted** until end of line, **`#`**, or a line break (with trailing spaces trimmed as above). **`KEY=`** with nothing after the equals (before newline) yields an **empty string** value.

### Escapes (inside **`"`** and **`` ` ``** only)

After **`\`**, the following are recognized (others may leave a literal `\` and the next character is read normally):

| Sequence | Meaning |
|----------|---------|
| `\n` `\r` `\t` | Newline, carriage return, tab |
| `\b` | Backspace |
| `\\` | Backslash |
| `\"` `` \` `` | Quote character |
| `\'` | Apostrophe (handled in the escape helper; double/backtick strings) |
| `\uXXXX` | Unicode code point, 4 hex digits |
| `\UXXXXXXXX` | Unicode code point, 8 hex digits |

Invalid `\u` / `\U` sequences produce a **`ParseError`**.

### Unicode

- The input is scanned as **UTF-8 runes**, so keys and values can contain **any Unicode** (emoji, CJK, etc.) without special syntax.
- You can also spell code points with **`\u`** / **`\U`** inside double- or backtick-quoted values.

### Errors

- Failures return **`*ParseError`** with **`Error()`** / **`String()`** text that includes **line and column** (1-based in the lexer).
- Typical messages start with **`Invalid syntax:`** (unexpected character, value without key, bad key shape, etc.).

### Not in this package

- **Variable interpolation** (`$VAR`, `${VAR}`) and **command substitution** are **not** applied during **`Parse`**. Use the **`env`** package on individual values if you need that.

## Quick start: parse and read

```go
package main

import (
	"fmt"
	"log"

	"github.com/voidbear-io/go/dotenv"
)

func main() {
	raw := `FOO=bar
# database
DB_URL="postgres://localhost/dev"
`
	doc, err := dotenv.Parse(raw)
	if err != nil {
		log.Fatal(err)
	}

	val, ok := doc.Get("FOO")
	if ok {
		fmt.Println("FOO =", val)
	}

	// Or ignore structure and use a map:
	fmt.Println(doc.ToMap())
}
```

## Quick start: build a doc in code

```go
package main

import (
	"fmt"

	"github.com/voidbear-io/go/dotenv"
)

func main() {
	doc := dotenv.NewDoc()
	doc.AddComment("App settings")
	doc.AddVariable("PORT", "8080")
	doc.AddQuotedVariable("MESSAGE", `hello "world"`, '"')
	fmt.Print(doc.String())
}
```

Use `doc.String()` as the body for `os.WriteFile` (or your config writer) when you want a real file on disk.

## Tips

- **`ToMap()`** is handy for config structs; use **`ToArray()`** or **`At(i)`** when order and comments matter.
- **`Merge`** overlays another doc’s variables (comments in the source doc are not copied).
- See **`mod_test.go`** for quoted values, escapes, unicode, and edge cases you can copy into your own tests.

## License

MIT — see [`LICENSE.md`](LICENSE.md).
