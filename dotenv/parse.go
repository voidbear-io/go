package dotenv

import (
	"errors"
	"fmt"
	"strconv"
	"unicode"
)

const (
	quote_none     = 0
	quote_single   = 1
	quote_double   = 2
	quote_backtick = 3
	TOKEN_COMMENT  = 1
	TOKEN_NAME     = 2
	TOKEN_VALUE    = 3
	TOKEN_NEWLINE  = 4
	token_none     = 0
)

type Mark struct {
	Line   int
	Column int
}

type Token struct {
	Type     int
	RawValue []rune
	value    *string
	Quote    int
	Start    *Mark
	End      *Mark
}

type parseState struct {
	Last          *Token
	Line          int
	Column        int
	Quote         int
	Buffer        []rune
	Tokens        []*Token
	Start         *Mark
	KeyTerminated bool
	Kind          int
}

func (state *parseState) SetKind(kind int) {
	if state.Kind != kind {
		state.Kind = kind
	}

	state.KeyTerminated = false
	state.Start = nil
}

type ParseError struct {
	Message string
	Line    int
	Column  int
}

func (t *Token) Value() string {
	if t.value != nil {
		return *t.value
	}
	return string(t.RawValue)
}

func (e *ParseError) Error() string {
	return e.Message + " at line " + strconv.Itoa(e.Line) + ", column " + strconv.Itoa(e.Column)
}

func (e *ParseError) String() string {
	return e.Message + " at line " + strconv.Itoa(e.Line) + ", column " + strconv.Itoa(e.Column)
}

func Lex(input string) ([]*Token, error) {

	state := &parseState{
		Last:   nil,
		Line:   1,
		Column: 0,
		Quote:  quote_none,
		Buffer: []rune{},
		Tokens: []*Token{},
		Start:  &Mark{Line: 1, Column: 1},
	}
	runes := []rune(input)
	max := len(runes)
	for i := 0; i < len(runes); i++ {
		state.Column++
		if state.Start == nil {
			state.Start = &Mark{Line: state.Line, Column: state.Column}
		}
		c := runes[i]

		var p rune
		p = rune(0)

		if i+1 < max {
			p = runes[i+1]
		}

		if state.Quote != quote_none {
			// println("handle quoted values")

			start := state.Start
			// println("Current character:", string(c), "Previous character:", string(b), "Next character:", string(p))
			switch state.Quote {
			case quote_single:
				if c == '\\' && p == '\'' {
					// println("Handling escaped single quote")
					state.Buffer = append(state.Buffer, '\'')
					i++ // skip next character
					state.Column++
					continue
				}

				if c == '\'' {
					// println("end single quote")
					captureToken(state, TOKEN_VALUE)
					state.SetKind(token_none)
					state.Start = start
				} else {
					state.Buffer = append(state.Buffer, c)
					continue
				}

			case quote_double:
				if c == '\\' && p == '"' {
					// println("Handling escaped double quote")
					state.Buffer = append(state.Buffer, '"')
					i++ // skip next character
					state.Column++
					continue
				}

				if c == '"' {
					// println("end double quote")
					captureToken(state, TOKEN_VALUE)
					state.SetKind(token_none)
					state.Start = start
				} else {
					// println("Handling escaped characters in quoted value " + string(c) + " with next character " + string(p))
					if c != '\\' {
						// println("Appending character to buffer:", string(c))
						state.Buffer = append(state.Buffer, c)
						continue
					}

					shift, err := handleQuotedChar(state, runes, i, c, p)
					if err != nil {
						return nil, err
					}
					if shift > 0 {
						i += shift
						state.Column += shift
					}
					continue
				}

			case quote_backtick:
				if c == '\\' && p == '`' {
					// println("Handling escaped backtick")
					state.Buffer = append(state.Buffer, '`')
					i++ // skip next character
					state.Column++
					continue
				}

				if c == '`' {
					// println("end backtick quote")
					captureToken(state, TOKEN_VALUE)
					state.SetKind(token_none)
					state.Start = start
				} else {

					// println("Handling escaped characters in quoted value " + string(c) + " with next character " + string(p))
					if c != '\\' {
						// println("Appending character to buffer:", string(c))
						state.Buffer = append(state.Buffer, c)
						continue
					}

					shift, err := handleQuotedChar(state, runes, i, c, p)
					if err != nil {
						return nil, err
					}
					if shift > 0 {
						i += shift
						state.Column += shift
					}
					continue
				}
			}

			i++
			state.Column++
			comment := false
			for i < max {
				c = runes[i]
				if c == '\n' || c == '\r' {
					n := i + 1
					if c == '\r' && n < max && runes[n] == '\n' {
						i++
					}

					if comment {
						captureToken(state, TOKEN_COMMENT)
					}

					state.Column = 0
					state.Line++
					// println("linebreak after quoted value")
					break
				}

				if comment {
					state.Buffer = append(state.Buffer, c)
					state.Column++
					i++
					continue
				}

				if c == '#' {
					comment = true
					state.SetKind(TOKEN_COMMENT)
					state.Start = &Mark{Line: state.Line, Column: state.Column}
					i++
					state.Column++
					continue
				}

				if !comment && !unicode.IsSpace(c) {
					return nil, errors.New("Invalid syntax: unexpected character '" + string(c) + "' after quoted value ended and before newline")
				}

				state.Column++
				i++
			}
			continue

		} else {
			linebreak := false
			shift := 0
			if c == '\n' || c == 0 {
				linebreak = true
			} else if c == '\r' && p == '\n' {
				shift = 1
				linebreak = true
			} else if c == '\r' {
				linebreak = true
			}

			if linebreak {
				// println("linebreak after " + string(state.Buffer))

				switch state.Kind {
				case TOKEN_NAME:
					{
						// println("token name", "buffer", string(state.Buffer))
						captureToken(state, TOKEN_NAME)
						i += shift
						// println("Adding TOKEN_NAME:", string(state.Last.RawValue))
						state.SetKind(TOKEN_VALUE)
						state.Column = 0
						state.Line++
					}
				case token_none:
					{
						// println("token none", "buffer", string(state.Buffer))
						if len(state.Buffer) == 0 {

							captureToken(state, TOKEN_NEWLINE)
							i += shift
							// println("Adding TOKEN_NAME:", string(state.Last.RawValue))
							state.SetKind(token_none)
							state.Column = 0
							state.Line++
						}

					}
				case TOKEN_COMMENT:
					{
						captureToken(state, TOKEN_COMMENT)
						i += shift
						state.SetKind(token_none)
						state.Column = 0
						state.Line++
					}
				case TOKEN_VALUE:
					{
						// println("copy buffer", string(state.Buffer))
						captureToken(state, TOKEN_VALUE)
						i += shift
						state.Column = 0
						state.Line++
						state.SetKind(token_none)
					}
				}

				continue
			}

			switch state.Kind {
			case token_none:
				if c == '\n' || c == '\r' {
					shift := 0
					if c == '\r' && i+1 < max && runes[i+1] == '\n' {
						shift = 1
						state.Column++
					}
					captureToken(state, TOKEN_NEWLINE)
					i += shift
					state.Line++
					state.Column = 0
					continue
				}

				if unicode.IsSpace(c) {
					continue
				}
				if c == '#' {
					state.SetKind(TOKEN_COMMENT)
					state.Start = &Mark{Line: state.Line, Column: state.Column}
					continue
				}

				if unicode.IsLetter(c) || unicode.IsDigit(c) || c == '_' {
					state.SetKind(TOKEN_NAME)
					state.Start = &Mark{Line: state.Line, Column: state.Column}
					state.Buffer = append(state.Buffer, c)
					continue
				}

				fail := &ParseError{
					Message: "Invalid syntax: unexpected character '" + string(c) + "'",
					Line:    state.Line,
					Column:  state.Column,
				}

				return nil, fail

			case TOKEN_NAME:
				{
					// do not append, ignore # and continue
					if c == '#' && len(state.Buffer) == 0 {
						state.SetKind(TOKEN_COMMENT)
						state.Start = nil
						continue
					}

					if c == '=' {
						// println("Found equal sign, capturing token")
						// println("Adding TOKEN_NAME:", string(state.Buffer))

						captureToken(state, TOKEN_NAME)
						state.SetKind(TOKEN_VALUE)
						continue
					}

					if unicode.IsLetter(c) || unicode.IsDigit(c) || c == '_' {

						if state.KeyTerminated {
							e := &ParseError{
								Message: "Invalid syntax: key terminated by whitespace. fix key",
								Line:    state.Line,
								Column:  state.Column,
							}

							return nil, e
						}

						state.Buffer = append(state.Buffer, c)
						continue
					}

					if unicode.IsSpace(c) {
						if len(state.Buffer) > 0 {
							state.KeyTerminated = true
						}

						continue
					}

					fail := &ParseError{
						Message: "Invalid syntax: unexpected character in name '" + string(c) + "'",
						Line:    state.Line,
						Column:  state.Column,
					}

					return nil, fail
				}
			case TOKEN_COMMENT:
				{
					// trim leading spaces in comment
					if len(state.Buffer) == 0 && unicode.IsSpace(c) {
						continue
					}

					state.Buffer = append(state.Buffer, c)
					continue
				}

			case TOKEN_VALUE:
				{
					if len(state.Buffer) == 0 {
						switch c {
						case '"':
							// println("Found double quote, starting quoted value")
							state.Quote = quote_double
							state.Start = &Mark{Line: state.Line, Column: state.Column}
							continue
						case '\'':
							// println("Found single quote, starting quoted value")
							state.Quote = quote_single
							state.Start = &Mark{Line: state.Line, Column: state.Column}
							continue
						case '`':
							// println("Found backtick, starting quoted value")
							state.Quote = quote_backtick
							state.Start = &Mark{Line: state.Line, Column: state.Column}
							continue
						default:
							if c == '\n' || c == '\r' {
								shift := 0
								if c == '\r' && i+1 < max && runes[i+1] == '\n' {
									shift = 1
								}
								captureToken(state, TOKEN_NEWLINE)
								state.SetKind(token_none)
								i += shift
								state.Start = nil
								state.Line++
								state.Column = 0
								continue
							}

							if unicode.IsSpace(c) {
								continue
							}

							// println("Appending character to buffer:", string(c))
							state.Buffer = append(state.Buffer, c)
							continue
						}
					}

					if c == '#' {
						// println("Found comment in value, capturing token")
						captureToken(state, TOKEN_VALUE)

						state.SetKind(TOKEN_COMMENT)
						continue
					}

					if len(state.Buffer) == 0 && unicode.IsSpace(c) {
						// println("found space in value, skipping")
						continue
					}

					// println("Appending character to buffer:", string(c))
					state.Buffer = append(state.Buffer, c)
				}
			}
		}
	}

	if len(state.Buffer) > 0 {
		switch state.Kind {
		case TOKEN_NAME:
			captureToken(state, TOKEN_NAME)
		case TOKEN_VALUE:
			captureToken(state, TOKEN_VALUE)
		case TOKEN_COMMENT:
			captureToken(state, TOKEN_COMMENT)
		case token_none:
			captureToken(state, TOKEN_NAME)
		}
	}

	return state.Tokens, nil
}

func handleQuotedChar(state *parseState, runes []rune, i int, c rune, p rune) (int, error) {
	max := len(runes)
	if c == '\\' {
		if p == 'n' || p == 'r' || p == 't' || p == '\\' || p == '"' || p == '\'' || p == 'b' || p == '`' {

			switch p {
			case 'n':
				state.Buffer = append(state.Buffer, '\n')
			case 'r':
				state.Buffer = append(state.Buffer, '\r')
			case 't':
				state.Buffer = append(state.Buffer, '\t')
			case 'b':
				state.Buffer = append(state.Buffer, '\b')
			default:
				state.Buffer = append(state.Buffer, p)
			}

			return 1, nil
		}

		if p == 'U' {

			if i+7 < max {
				// capture unicode escape sequence
				hex := string(runes[i+2 : i+10])
				if len(hex) == 8 {
					var codePoint rune
					_, err := fmt.Sscanf(hex, "%x", &codePoint)
					if err == nil {
						state.Buffer = append(state.Buffer, codePoint)
						return 9, nil // skip the next 9 characters
					} else {
						return 0, &ParseError{
							Message: "Invalid unicode escape sequence",
							Line:    state.Line,
							Column:  state.Column,
						}
					}
				}
			}
			return 0, nil
		}

		if p == 'u' {

			if i+5 < max {
				// capture unicode escape sequence
				hex := string(runes[i+2 : i+6])
				if len(hex) == 4 {
					var codePoint rune
					_, err := fmt.Sscanf(hex, "%x", &codePoint)
					if err == nil {
						state.Buffer = append(state.Buffer, codePoint)
						return 5, nil
					} else {
						return 0, &ParseError{
							Message: "Invalid unicode escape sequence",
							Line:    state.Line,
							Column:  state.Column,
						}
					}
				}
			}
			return 0, nil
		}

		state.Buffer = append(state.Buffer, c)
	}

	return 0, nil
}

func captureToken(state *parseState, kind int) *Token {
	copy2 := make([]rune, len(state.Buffer))
	copy(copy2, state.Buffer)

	if state.Quote == quote_none {
		l := len(copy2)
		pos := l - 1
		for pos >= 0 && unicode.IsSpace(copy2[pos]) {
			pos--
			l--
		}

		copy2 = copy2[:l]
	}

	token := &Token{
		Type:     kind,
		RawValue: copy2,
		Quote:    state.Quote,
		Start:    state.Start,
		End: &Mark{
			Line:   state.Line,
			Column: state.Column - 1,
		},
	}
	/*
		tokenName := ""
		switch kind {
		case TOKEN_COMMENT:
			tokenName = "TOKEN_COMMENT"
		case TOKEN_NAME:
			tokenName = "TOKEN_NAME"
		case TOKEN_VALUE:
			tokenName = "TOKEN_VALUE"
		case TOKEN_NEWLINE:
			tokenName = "TOKEN_NEWLINE"
		default:
			tokenName = "TOKEN_UNKNOWN"
		}

		println(tokenName, string(token.RawValue), "Quote:", token.Quote, "Start:", state.Start.Line, state.Start.Column, "End:", token.End.Line, token.End.Column) */

	if state.Kind == TOKEN_NAME {
		state.Kind = TOKEN_VALUE
	} else {
		state.Kind = token_none
	}

	state.Last = token
	state.Quote = quote_none
	state.Start = nil
	state.Buffer = []rune{}
	state.Tokens = append(state.Tokens, token)
	return token
}

func Parse(input string) (*EnvDoc, error) {
	tokens, err := Lex(input)
	if err != nil {
		return nil, err
	}

	doc := &EnvDoc{
		tokens: make([]Element, 0, len(tokens)),
	}

	var key *string

	for _, token := range tokens {
		switch token.Type {
		case TOKEN_NEWLINE:
			key = nil
			doc.AddNewline()
			// println("newline")
			continue
		case TOKEN_COMMENT:
			key = nil
			doc.AddComment(string(token.RawValue))
			// println("#", string(token.RawValue))

		case TOKEN_NAME:

			if key == nil {
				v := string(token.RawValue)
				key = &v
				// println("TOKEN_NAME:", *key)
				continue
			}
			// println("TOKEN_NAME:", *key)
			doc.AddVariable(*key, "")
			v2 := string(token.RawValue)
			key = &v2
		case TOKEN_VALUE:
			// println("TOKEN_VALUE:", string(token.RawValue))
			if key == nil {
				return nil, &ParseError{
					Message: "Invalid syntax: value without a key",
					Line:    token.Start.Line,
					Column:  token.Start.Column,
				}
			}

			if token.Quote == quote_none {
				doc.AddVariable(*key, string(token.RawValue))
				key = nil
				continue
			}

			var r rune
			switch token.Quote {
			case quote_single:
				r = '\''
			case quote_double:
				r = '"'
			case quote_backtick:
				r = '`'
			default:
				r = '"'
			}

			doc.AddQuotedVariable(*key, string(token.RawValue), r)
			key = nil
		}
	}

	if key != nil {
		doc.AddVariable(*key, "")
	}

	return doc, nil
}
