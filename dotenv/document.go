// Package dotenv parses and serializes .env-style files while preserving element order,
// comments, and quoting metadata via EnvDoc.
package dotenv

const (
	// represents a newline element in the document
	NEWLINE = 0
	// represents a comment element in the document
	COMMENT = 1
	// represents a variable element in the document
	VARIABLE   = 2
	empty_node = 3
)

// Element represents a single element in the dotenv document, `EnvDoc`.
type Element struct {
	Type   int
	Value  string
	Key    *string
	Inline bool
	Quote  *rune
}

// EnvDoc represents a dotenv document consisting of multiple elements.
type EnvDoc struct {
	tokens []Element
}

// NewDoc creates and returns a new instance of EnvDoc.
func NewDoc() *EnvDoc {
	return &EnvDoc{
		tokens: make([]Element, 0),
	}
}

// AddNewline adds a newline element to the document.
func (doc *EnvDoc) AddNewline() {
	doc.tokens = append(doc.tokens, Element{Type: NEWLINE, Value: "\n"})
}

// AddComment adds a comment element to the document.
// The comment will be placed on its own line.
func (doc *EnvDoc) AddComment(comment string) {
	doc.tokens = append(doc.tokens, Element{Type: COMMENT, Value: comment})
}

// AddInlineComment adds an inline comment element to the document.
// The comment will be placed on the same line as the preceding variable.
func (doc *EnvDoc) AddInlineComment(comment string) {
	doc.tokens = append(doc.tokens, Element{Type: COMMENT, Value: comment, Inline: true})
}

// AddVariable adds a variable element to the document.
func (doc *EnvDoc) AddVariable(key, value string) {
	if value == "" {
		doc.tokens = append(doc.tokens, Element{
			Type:  VARIABLE,
			Key:   &key,
			Value: "",
		})
		return
	}

	if value[0] == '"' || value[0] == '\'' {
		quote := rune(value[0])
		doc.tokens = append(doc.tokens, Element{
			Type:  VARIABLE,
			Value: value[1 : len(value)-1],
			Key:   &key,
			Quote: &quote,
		})
		return
	}

	quoted := false
	runes := []rune(value)
	min := rune(0)
	l := len(runes)
	for j := 0; j < l; j++ {
		c := runes[j]
		n := min
		if j+1 < l {
			n = runes[j+1]
		}

		if c == '\\' {
			if n == '\\' || n == 'n' || n == 'r' || n == 't' || n == 'u' || n == 'U' || n == 'b' || n == 'f' {
				quoted = true
				break
			}
		}

		if c == '"' || c == '\'' || c == '\n' || c == '\r' || c == '\t' || c == '=' || c == '#' || c == '\b' || c == '\f' || c == '\v' {
			quoted = true
			break
		}
	}

	var quote *rune

	if quoted {
		r := rune('"')
		quote = &r
	}

	doc.tokens = append(doc.tokens, Element{
		Type:  VARIABLE,
		Value: value,
		Key:   &key,
		Quote: quote,
	})
}

// AddQuotedVariable adds a variable element with a specific quote to the document.
func (doc *EnvDoc) AddQuotedVariable(key, value string, quote rune) {
	doc.tokens = append(doc.tokens, Element{
		Type:  VARIABLE,
		Value: value,
		Key:   &key,
		Quote: &quote,
	})
}

// Add adds a generic element to the document.
func (doc *EnvDoc) Add(token Element) {
	if token.Type == NEWLINE || token.Type == COMMENT || token.Type == VARIABLE {
		doc.tokens = append(doc.tokens, token)
	} else {
		// Ignore other types of tokens
		return
	}
}

// AddRange adds multiple elements to the document.
func (doc *EnvDoc) AddRange(tokens []Element) {
	if len(tokens) == 0 {
		return
	}

	for _, token := range tokens {
		doc.Add(token)
	}
}

// Len returns the number of elements in the document.
func (doc *EnvDoc) Len() int {
	return len(doc.tokens)
}

// At returns the element at the specified index.
func (doc *EnvDoc) At(index int) *Element {
	if index < 0 || index >= len(doc.tokens) {
		return nil
	}
	return &doc.tokens[index]
}

// ToArray returns the elements of the document as a slice.
func (doc *EnvDoc) ToArray() []Element {
	if doc == nil {
		return []Element{}
	}

	arr := make([]Element, len(doc.tokens))
	copy(arr, doc.tokens)
	return arr
}

// ToMap converts the document's variable elements to a map.
func (doc *EnvDoc) ToMap() map[string]string {
	m := make(map[string]string)
	for _, token := range doc.tokens {
		if token.Type == VARIABLE && token.Key != nil {
			m[*token.Key] = token.Value
		}
	}
	return m
}

// Get retrieves the value of a variable by its key.
func (doc *EnvDoc) Get(key string) (string, bool) {
	for _, token := range doc.tokens {
		if token.Type == VARIABLE && token.Key != nil && *token.Key == key {
			return token.Value, true
		}
	}
	return "", false
}

// Keys returns a slice of all variable keys in the document.
func (doc *EnvDoc) Keys() []string {
	keys := make([]string, 0, len(doc.tokens))
	for _, token := range doc.tokens {
		if token.Type == VARIABLE && token.Key != nil {
			keys = append(keys, *token.Key)
		}
	}
	return keys
}

// GetComments returns a slice of all comment values in the document.
func (doc *EnvDoc) GetComments() []string {
	comments := make([]string, 0, len(doc.tokens))
	for _, token := range doc.tokens {
		if token.Type == COMMENT {
			comments = append(comments, token.Value)
		}
	}
	return comments
}

// Set sets the value of a variable by its key. If the variable does not exist, it is added.
func (doc *EnvDoc) Set(key, value string) {
	isset := false
	for i, token := range doc.tokens {
		if token.Type == VARIABLE && token.Key != nil && *token.Key == key {
			doc.tokens[i].Value = value
			isset = true
			break
		}
	}
	if !isset {
		doc.AddVariable(key, value)
	}
}

// Merge merges another EnvDoc into the current document.
func (doc *EnvDoc) Merge(other *EnvDoc) {
	for _, token := range other.tokens {
		switch token.Type {
		case VARIABLE:
			doc.Set(*token.Key, token.Value)
		}
	}
}

// String converts the document to its string representation.
func (doc *EnvDoc) String() string {
	var result string
	empty := &Element{
		Type:  empty_node,
		Value: "",
	}

	for i := 0; i < len(doc.tokens); i++ {
		token := doc.tokens[i]
		nextToken := empty
		if i+1 < len(doc.tokens) {
			nextToken = &doc.tokens[i+1]
		}

		switch token.Type {
		case NEWLINE:
			result += "\n"
		case COMMENT:
			result += "\n# " + token.Value
		case VARIABLE:
			result += "\n"
			if token.Key == nil {
				continue
			}

			result += *token.Key + "="
			if token.Quote != nil {
				result += string(*token.Quote)
				result += token.Value
				result += string(*token.Quote)
			} else {
				result += token.Value
			}

			if nextToken.Type == COMMENT && nextToken.Inline {
				result += " # " + nextToken.Value
				i++
			}
		}
	}

	if len(result) > 0 && result[0] == '\n' {
		result = result[1:] // Remove leading newline
	}

	return result
}
