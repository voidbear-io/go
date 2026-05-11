package exec

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"
)

type Result struct {
	Stdout    []byte
	Stderr    []byte
	Code      int
	FileName  string
	Args      []string
	StartedAt time.Time
	EndedAt   time.Time
	TempFile  *string
}

func (o *Result) Text() string {
	return string(o.Stdout)
}

func (o *Result) IsOk() bool {
	return o.Code == 0
}

func (o *Result) ToError() error {
	if o.IsOk() {
		return nil
	}
	return fmt.Errorf("command %s failed with code %d: %s", o.FileName, o.Code, o.ErrorText())
}

func (o *Result) ToErrorIf(f func(o *Result) bool) error {
	if f == nil {
		f = func(o *Result) bool {
			return o.Code != 0
		}
	}

	if f(o) {
		return o.ToError()
	}
	return nil
}

func (o *Result) Lines() []string {
	r := bytes.Split(o.Stdout, []byte(EOL))
	lines := []string{}
	for _, line := range r {
		lines = append(lines, string(line))
	}
	return lines
}

func (o *Result) ErrorText() string {
	return string(o.Stderr)
}

func (o *Result) ErrorLines() []string {
	r := bytes.Split(o.Stderr, []byte(EOL))
	lines := []string{}
	for _, line := range r {
		lines = append(lines, string(line))
	}
	return lines
}

func (o *Result) Json() (interface{}, error) {
	var out interface{}
	err := json.Unmarshal([]byte(o.Stdout), &out)
	return out, err
}

func (o *Result) ErrorJson() (interface{}, error) {
	var out interface{}
	err := json.Unmarshal([]byte(o.Stderr), &out)
	return out, err
}

func (o *Result) Validate() (bool, error) {
	return o.ValidateWith(nil)
}

func (o *Result) ValidateWith(cb func(o *Result) (bool, error)) (bool, error) {
	if cb == nil {
		cb = func(o *Result) (bool, error) {
			if o.Code != 0 {
				return false, fmt.Errorf("command %s failed with code %d", o.FileName, o.Code)
			}

			return true, nil
		}
	}

	return cb(o)
}
