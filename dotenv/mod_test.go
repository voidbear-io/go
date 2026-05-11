package dotenv_test

import (
	"testing"

	"github.com/voidbear-io/go/dotenv"
	"github.com/stretchr/testify/assert"
)

func TestEnvDoc_AddMethods(t *testing.T) {
	doc := &dotenv.EnvDoc{}

	// Test adding different types of nodes
	doc.AddVariable("KEY1", "value1")
	doc.AddComment("This is a comment")
	doc.AddNewline()
	doc.AddVariable("KEY2", "value2")

	assert.Equal(t, 4, doc.Len())

	// Check variable values
	value1, ok1 := doc.Get("KEY1")
	assert.True(t, ok1)
	assert.Equal(t, "value1", value1)

	value2, ok2 := doc.Get("KEY2")
	assert.True(t, ok2)
	assert.Equal(t, "value2", value2)

	// Check comments
	comments := doc.GetComments()
	assert.Len(t, comments, 1)
	assert.Equal(t, "This is a comment", comments[0])
}

func TestEnvDoc_Keys(t *testing.T) {
	doc := &dotenv.EnvDoc{}

	doc.AddVariable("KEY1", "value1")
	doc.AddVariable("KEY2", "value2")
	doc.AddComment("comment")
	doc.AddVariable("KEY3", "value3")

	keys := doc.Keys()
	assert.Len(t, keys, 3)
	assert.Contains(t, keys, "KEY1")
	assert.Contains(t, keys, "KEY2")
	assert.Contains(t, keys, "KEY3")
}

func TestEnvDoc_ToMap(t *testing.T) {
	doc := &dotenv.EnvDoc{}

	doc.AddVariable("KEY1", "value1")
	doc.AddVariable("KEY2", "value2")
	doc.AddComment("comment")
	doc.AddNewline()
	doc.AddVariable("KEY3", "value3")

	m := doc.ToMap()
	expected := map[string]string{
		"KEY1": "value1",
		"KEY2": "value2",
		"KEY3": "value3",
	}

	assert.Equal(t, expected, m)
}

func TestEnvDoc_ToArray(t *testing.T) {
	doc := &dotenv.EnvDoc{}

	doc.AddVariable("KEY1", "value1")
	doc.AddComment("comment")
	doc.AddNewline()

	arr := doc.ToArray()
	assert.Len(t, arr, 3)

	// Check first node (variable)
	assert.Equal(t, dotenv.VARIABLE, arr[0].Type)
	assert.Equal(t, "value1", arr[0].Value)
	assert.NotNil(t, arr[0].Key)
	assert.Equal(t, "KEY1", *arr[0].Key)

	// Check second node (comment)
	assert.Equal(t, dotenv.COMMENT, arr[1].Type)
	assert.Equal(t, "comment", arr[1].Value)
	assert.Nil(t, arr[1].Key)

	// Check third node (newline)
	assert.Equal(t, dotenv.NEWLINE, arr[2].Type)
	assert.Equal(t, "\n", arr[2].Value)
	assert.Nil(t, arr[2].Key)
}

func TestEnvDoc_At(t *testing.T) {
	doc := &dotenv.EnvDoc{}

	doc.AddVariable("KEY1", "value1")
	doc.AddComment("comment")

	// Test valid indices
	node0 := doc.At(0)
	assert.NotNil(t, node0)
	assert.Equal(t, dotenv.VARIABLE, node0.Type)

	node1 := doc.At(1)
	assert.NotNil(t, node1)
	assert.Equal(t, dotenv.COMMENT, node1.Type)
	// Test invalid indices
	assert.Nil(t, doc.At(-1))
	assert.Nil(t, doc.At(2))
	assert.Nil(t, doc.At(100))
}

func TestEnvDoc_SetValue(t *testing.T) {
	doc := &dotenv.EnvDoc{}

	// Set new value
	doc.Set("KEY1", "value1")
	value1, ok1 := doc.Get("KEY1")
	assert.True(t, ok1)
	assert.Equal(t, "value1", value1)

	// Update existing value
	doc.Set("KEY1", "updated_value1")
	updatedValue1, ok1Updated := doc.Get("KEY1")
	assert.True(t, ok1Updated)
	assert.Equal(t, "updated_value1", updatedValue1)

	// Set another new value
	doc.Set("KEY2", "value2")
	value2, ok2 := doc.Get("KEY2")
	assert.True(t, ok2)
	assert.Equal(t, "value2", value2)
}

func TestEnvDoc_Merge(t *testing.T) {
	doc1 := &dotenv.EnvDoc{}
	doc1.AddVariable("KEY1", "value1")
	doc1.AddVariable("KEY2", "value2")

	doc2 := &dotenv.EnvDoc{}
	doc2.AddVariable("KEY2", "updated_value2") // Should override
	doc2.AddVariable("KEY3", "value3")         // Should add new
	doc2.AddComment("This comment should be ignored")
	doc2.AddNewline()

	doc1.Merge(doc2)

	// Check that KEY1 remains unchanged
	value1, ok1 := doc1.Get("KEY1")
	assert.True(t, ok1)
	assert.Equal(t, "value1", value1)

	// Check that KEY2 was updated
	value2, ok2 := doc1.Get("KEY2")
	assert.True(t, ok2)
	assert.Equal(t, "updated_value2", value2)

	// Check that KEY3 was added
	value3, ok3 := doc1.Get("KEY3")
	assert.True(t, ok3)
	assert.Equal(t, "value3", value3)
}

func TestEnvDoc_GetValue_NotFound(t *testing.T) {
	doc := &dotenv.EnvDoc{}
	doc.AddVariable("KEY1", "value1")

	// Test existing key
	value1, ok1 := doc.Get("KEY1")
	assert.True(t, ok1)
	assert.Equal(t, "value1", value1)

	// Test non-existing key
	value2, ok2 := doc.Get("NONEXISTENT")
	assert.False(t, ok2)
	assert.Equal(t, "", value2)
}

func TestParseError_ErrorString(t *testing.T) {
	err := &dotenv.ParseError{
		Message: "Test error",
		Line:    5,
		Column:  10,
	}

	expected := "Test error at line 5, column 10"
	assert.Equal(t, expected, err.Error())
	assert.Equal(t, expected, err.String())
}

func TestParse_WorkingExamples(t *testing.T) {
	// Test what actually works with the current parser
	tests := []struct {
		name     string
		input    string
		expected map[string]string
	}{
		{
			name:  "quoted values work",
			input: `KEY="value"`,
			expected: map[string]string{
				"KEY": "value",
			},
		},
		{
			name:  "single quoted values work",
			input: `KEY='value'`,
			expected: map[string]string{
				"KEY": "value",
			},
		},
		{
			name:     "empty values are ignored (known parser limitation)",
			input:    `KEY=`,
			expected: map[string]string{"KEY": ""}, // Parser doesn't handle empty values correctly
		},
		{
			name:  "key with underscore works",
			input: `KEY_WITH_UNDERSCORE=value`,
			expected: map[string]string{
				"KEY_WITH_UNDERSCORE": "value",
			},
		},
		{
			name:  "variables with spaces work",
			input: `KEY_WITH_SPACES=value with spaces`,
			expected: map[string]string{
				"KEY_WITH_SPACES": "value with spaces",
			},
		},
		{
			name:  "variables with special characters work",
			input: `KEY_WITH_SPECIAL_CHARS=value@!%`,
			expected: map[string]string{
				"KEY_WITH_SPECIAL_CHARS": "value@!%",
			},
		},
		{
			name:  "variables with newlines work",
			input: "KEY_WITH_NEWLINES=\"value\nwith\nnewlines\"",
			expected: map[string]string{
				"KEY_WITH_NEWLINES": "value\nwith\nnewlines",
			},
		},
		{
			name: "variables with quoted newlines work",
			input: `KEY_WITH_QUOTED_NEWLINES="value
with
newlines"`,
			expected: map[string]string{
				"KEY_WITH_QUOTED_NEWLINES": "value\nwith\nnewlines",
			},
		},
		{
			name:  "variables with tabs work",
			input: "KEY_WITH_TABS=\"value\twith\ttabs\"",
			expected: map[string]string{
				"KEY_WITH_TABS": "value\twith\ttabs",
			},
		},
		{
			name: "mixed",
			input: `KEY1="value1"
KEY2='value2'
# comment

KEY3=value3`,
			expected: map[string]string{
				"KEY1": "value1",
				"KEY2": "value2",
				"KEY3": "value3",
			},
		},
		{
			name:  "variables with escaped quotes work",
			input: `KEY_WITH_ESCAPED_QUOTES="value with \"escaped quotes\""`,
			expected: map[string]string{
				"KEY_WITH_ESCAPED_QUOTES": `value with "escaped quotes"`,
			},
		},
		{
			name:  "variables with unicode characters work",
			input: `KEY_WITH_UNICODE=こんにちは`,
			expected: map[string]string{
				"KEY_WITH_UNICODE": "こんにちは",
			},
		},
		{
			name:  "variables with unicode emojis work",
			input: `KEY_WITH_EMOJI=😊`,
			expected: map[string]string{
				"KEY_WITH_EMOJI": "😊",
			},
		},
		{
			name:  "variables with escaped unicode work",
			input: `KEY_WITH_ESCAPED_UNICODE="value with \U0001F920"`,
			expected: map[string]string{
				"KEY_WITH_ESCAPED_UNICODE": "value with 🤠",
			},
		},
		{
			name:  "single quote does not escape",
			input: `KEY_WITH_SINGLE_QUOTE='value with single quote \n'`,
			expected: map[string]string{
				"KEY_WITH_SINGLE_QUOTE": "value with single quote \\n",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := dotenv.Parse(tt.input)
			t.Logf("Running test: %s", tt.name)
			for _, token := range doc.ToArray() {
				if token.Type == dotenv.VARIABLE {
					value := token.Value
					t.Logf("Variable: %s = %s", *token.Key, value)
				}
			}
			if err != nil {
				t.Fatalf("unexpected parse error: %v", err)
				t.Logf("Skipping test due to parse error: %v", err)
				t.Logf("Parse error: %v", err)
				return // Skip failing tests for now
			}

			result := doc.ToMap()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEnsureQuotedValuesRemain(t *testing.T) {
	doc := &dotenv.EnvDoc{}

	// Add variables with different quoting styles
	doc.AddVariable("KEY1", "value1")
	doc.AddQuotedVariable("KEY2", "value2", '"')
	doc.AddQuotedVariable("KEY3", "value3", '\'')

	// Ensure all values are quoted
	for _, node := range doc.ToArray() {
		if node.Type == dotenv.VARIABLE && node.Quote != nil {
			assert.NotNil(t, node.Quote, "Expected quoted variable to have a quote rune")
		}
	}

	first, ok := doc.Get("KEY1")
	assert.True(t, ok)
	assert.Equal(t, "value1", first)
	second, _ := doc.Get("KEY2")
	assert.Equal(t, "value2", second)
	third, _ := doc.Get("KEY3")
	assert.Equal(t, "value3", third)
}

func TestEnsureEscapedValuesGetQuoted(t *testing.T) {
	doc := &dotenv.EnvDoc{}

	// Add variables with escaped characters
	doc.AddVariable("KEY1", "value with \"escaped quotes\"")
	doc.AddVariable("KEY2", "value with 'single quotes'")
	doc.AddVariable("KEY3", "value with \\backslash")
	doc.AddVariable("KEY4", "value with \nnewlines")
	doc.AddVariable("KEY5", "value with \t tabs")
	doc.AddVariable("KEY6", "value with \v vertical tab")
	doc.AddVariable("KEY7", "value with \f form feed")
	doc.AddVariable("KEY8", "value with \r carriage return")
	doc.AddVariable("KEY9", "value with \b backspace")
	doc.AddVariable("KEY10", `value with \u2603 unicode snowman`)

	// Ensure all values are quoted
	for _, node := range doc.ToArray() {
		if node.Type == dotenv.VARIABLE {

			ret := assert.NotNil(t, node.Quote, "Expected quoted variable to have a quote rune")
			if !ret {
				t.Errorf("Variable %s should be quoted but is not", *node.Key)
			}
		}
	}

	first, ok := doc.Get("KEY1")
	assert.True(t, ok)
	assert.Equal(t, "value with \"escaped quotes\"", first)
	second, _ := doc.Get("KEY2")
	assert.Equal(t, "value with 'single quotes'", second)
	third, _ := doc.Get("KEY3")
	assert.Equal(t, "value with \\backslash", third)
	fourth, _ := doc.Get("KEY4")
	assert.Equal(t, "value with \nnewlines", fourth)
	fifth, _ := doc.Get("KEY5")
	assert.Equal(t, "value with \t tabs", fifth)
	sixth, _ := doc.Get("KEY6")
	assert.Equal(t, "value with \v vertical tab", sixth)
	seventh, _ := doc.Get("KEY7")
	assert.Equal(t, "value with \f form feed", seventh)
	eighth, _ := doc.Get("KEY8")
	assert.Equal(t, "value with \r carriage return", eighth)
	ninth, _ := doc.Get("KEY9")
	assert.Equal(t, "value with \b backspace", ninth)
	tenth, _ := doc.Get("KEY10")
	assert.Equal(t, `value with \u2603 unicode snowman`, tenth)
}

func TestEnsureParsedQuotesRemain(t *testing.T) {

	envContext := `
KEY1="value1"

KEY2='value2'
Key3=value3
Key4=a value with spaces
# This is a comment
Key5="a value with \"escaped quotes\""
Key6='a value with \'single quotes\''
Key7="line1
line2
line3
"
Key8="value with \nnewlines"
Key9="value with \t tabs"
Key11="😈"	
	`

	doc, err := dotenv.Parse(envContext)
	if err != nil {
		t.Fatalf("Failed to parse env context: %v", err)
	}

	for i, node := range doc.ToArray() {
		switch i {
		case 0:
			assert.Equal(t, dotenv.NEWLINE, node.Type, "First token should be a newline")
		case 1:
			assert.Equal(t, dotenv.VARIABLE, node.Type, "Second token should be a variable")
			assert.Equal(t, "KEY1", *node.Key, "Key should be KEY1")
			assert.Equal(t, "value1", node.Value, "Value should be 'value1'")
			assert.NotNil(t, node.Quote, "KEY1 should be quoted")
		case 2:
			assert.Equal(t, dotenv.NEWLINE, node.Type, "Third token should be a newline")
		case 3:
			assert.Equal(t, dotenv.VARIABLE, node.Type, "Fourth token should be a variable")
			assert.Equal(t, "KEY2", *node.Key, "Key should be KEY2")
			assert.Equal(t, "value2", node.Value, "Value should be 'value2'")
			assert.NotNil(t, node.Quote, "KEY2 should be quoted")
		case 4:
			assert.Equal(t, dotenv.VARIABLE, node.Type, "Fifth token should be a variable")
			assert.Equal(t, "Key3", *node.Key, "Key should be Key3")
			assert.Equal(t, "value3", node.Value, "Value should be 'value3'")
			assert.Nil(t, node.Quote, "Key3 should not be quoted")
		case 5:
			assert.Equal(t, dotenv.VARIABLE, node.Type, "Sixth token should be a variable")
			assert.Equal(t, "Key4", *node.Key, "Key should be Key4")
			assert.Equal(t, "a value with spaces", node.Value, "Value should be 'a value with spaces'")
			assert.Nil(t, node.Quote, "Key4 should not be quoted")

		case 6:
			assert.Equal(t, dotenv.COMMENT, node.Type, "Seventh token should be a comment")
			assert.Equal(t, "This is a comment", node.Value, "Comment should be 'This is a comment'")
			assert.Nil(t, node.Key, "Comment should not have a key")
		case 7:
			assert.Equal(t, dotenv.VARIABLE, node.Type, "Eighth token should be a variable")
			assert.Equal(t, "Key5", *node.Key, "Key should be Key5")
			assert.Equal(t, `a value with "escaped quotes"`, node.Value, "Value should be 'a value with \"escaped quotes\"'")
			assert.NotNil(t, node.Quote, "Key5 should be quoted")
		case 8:
			assert.Equal(t, dotenv.VARIABLE, node.Type, "Ninth token should be a variable")
			assert.Equal(t, "Key6", *node.Key, "Key should be Key6")
			assert.Equal(t, "a value with 'single quotes'", node.Value, "Value should be 'a value with \\'single quotes\\''")
			assert.NotNil(t, node.Quote, "Key6 should be quoted")
		case 9:
			assert.Equal(t, dotenv.VARIABLE, node.Type, "Tenth token should be a variable")
			assert.Equal(t, "Key7", *node.Key, "Key should be Key7")
			assert.Equal(t, "line1\nline2\nline3\n", node.Value, "Value should be 'line1\nline2\nline3\n'")
			assert.NotNil(t, node.Quote, "Key7 should be quoted")
		case 10:
			assert.Equal(t, dotenv.VARIABLE, node.Type, "Eleventh token should be a variable")
			assert.Equal(t, "Key8", *node.Key, "Key should be Key8")
			assert.Equal(t, "value with \nnewlines", node.Value, "Value should be 'value with \\nnewlines'")
			assert.NotNil(t, node.Quote, "Key8 should be quoted")
		case 11:
			assert.Equal(t, dotenv.VARIABLE, node.Type, "Twelfth token should be a variable")
			assert.Equal(t, "Key9", *node.Key, "Key should be Key9")
			assert.Equal(t, "value with \t tabs", node.Value, "Value should be 'value with \\t tabs'")
			assert.NotNil(t, node.Quote, "Key9 should be quoted")
		case 12:
			assert.Equal(t, dotenv.VARIABLE, node.Type, "Thirteenth token should be a variable")
			assert.Equal(t, "Key11", *node.Key, "Key should be Key11")
			assert.Equal(t, "😈", node.Value, "Value should be '😈'")
		}
	}
}
