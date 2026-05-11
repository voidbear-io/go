package exec

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	osexec "os/exec"
	"path/filepath"
	"time"

	"github.com/voidbear-io/go/cmdargs"
)

const (
	STDIO_INHERIT = 0
	STDIO_PIPED   = 1
	STDIO_NULL    = 2
)

var (
	logger func(cmd *Cmd)
)

type Cmd struct {
	*osexec.Cmd
	ctx           *context.Context // if true, the command is a context command
	logger        func(cmd *Cmd)
	disableLogger bool
	TempFile      *string
}

func New(name string, args ...string) *Cmd {
	cmd := osexec.Command(name, args...)
	return &Cmd{Cmd: cmd}
}

func NewContext(ctx context.Context, name string, args ...string) *Cmd {
	cmd := osexec.CommandContext(ctx, name, args...)
	return &Cmd{Cmd: cmd, ctx: &ctx}
}

func SetLogger(f func(cmd *Cmd)) {
	logger = f
}

func (c *Cmd) SetLogger(f func(cmd *Cmd)) {
	c.logger = f
}

func (c *Cmd) DisableLogger() {
	c.disableLogger = true
}

func CommandContext(ctx context.Context, command string) *Cmd {
	exe := ""
	args := cmdargs.Split(command).ToArray()
	if len(args) > 0 {
		exe = args[0]
		args = args[1:]
	}
	return NewContext(ctx, exe, args...)
}

// Command parses the command and arguments and returns a new Cmd
// with the parsed command and arguments
// Example:
//
//	Command("echo hello world")
//	Command("echo 'hello world'")
func Command(command string) *Cmd {
	exe := ""
	args := cmdargs.Split(command).ToArray()
	if len(args) > 0 {
		exe = args[0]
		args = args[1:]
	}

	return New(exe, args...)
}

func Run(command string) (*Result, error) {
	args := cmdargs.Split(command).ToArray()
	if len(args) == 0 {
		return nil, errors.New("command cannot be empty")
	}
	if len(args) == 1 {
		return New(args[0]).Run()
	}

	command = args[0]
	args = args[1:]
	c := New(command, args...)
	return c.Run()
}

func Output(command string) (*Result, error) {
	args := cmdargs.Split(command).ToArray()
	if len(args) == 0 {
		return nil, errors.New("command cannot be empty")
	}
	if len(args) == 1 {
		return New(args[0]).Output()
	}

	command = args[0]
	args = args[1:]
	c := New(command, args...)
	return c.Output()
}

func (c *Cmd) AppendArgs(args ...string) *Cmd {
	c.Args = append(c.Args, args...)
	return c
}

func (c *Cmd) PrependArgs(args ...string) *Cmd {
	if len(args) == 0 {
		return c
	}
	c.Args = append(args, c.Args...)
	return c
}

func (c *Cmd) WithArgs(args ...string) *Cmd {
	c.Args = args
	return c
}

func (c *Cmd) AppendEnv(env ...string) *Cmd {
	c.Env = append(c.Env, env...)
	return c
}

func (c *Cmd) PrependEnv(env ...string) *Cmd {
	if len(env) == 0 {
		return c
	}
	c.Env = append(env, c.Env...)
	return c
}

func (c *Cmd) WithEnvMap(env map[string]string) *Cmd {
	data := make([]string, 0)
	for k, v := range env {
		data = append(data, k+"="+v)
	}
	return c.WithEnv(data...)
}

func (c *Cmd) WithEnv(env ...string) *Cmd {
	c.Env = env
	return c
}

func (c *Cmd) WithCwd(dir string) *Cmd {
	c.Dir = dir
	return c
}

func (c *Cmd) WithStdin(stdin io.Reader) *Cmd {
	c.Stdin = stdin
	return c
}

func (c *Cmd) WithStdout(stdout io.Writer) *Cmd {
	c.Stdout = stdout
	return c
}

func (c *Cmd) WithStderr(stderr io.Writer) *Cmd {
	c.Stderr = stderr
	return c
}

func (c *Cmd) WithStdio(stdin, stdout, stderr int) *Cmd {
	switch stdin {
	case STDIO_INHERIT:
		c.Stdin = os.Stdin
	case STDIO_PIPED:
		c.Stdin = bytes.NewBuffer(nil)
	case STDIO_NULL:
		c.Stdin = nil
	}

	switch stdout {
	case STDIO_INHERIT:
		c.Stdout = os.Stdout
	case STDIO_PIPED:
		c.Stdout = bytes.NewBuffer(nil)
	case STDIO_NULL:
		c.Stdout = nil
	}

	switch stderr {
	case STDIO_INHERIT:
		c.Stderr = os.Stderr
	case STDIO_PIPED:
		c.Stderr = bytes.NewBuffer(nil)
	case STDIO_NULL:
		c.Stderr = nil
	}

	return c
}

// Runs the command quietly, without any PsOutput
func (c *Cmd) Quiet() (*Result, error) {
	c.Stdout = nil
	c.Stderr = nil
	var out Result
	out.FileName = c.Path
	out.Args = c.Args
	out.Stdout = make([]byte, 0)
	out.Stderr = make([]byte, 0)
	// use utc time
	out.StartedAt = time.Now().UTC()

	err := c.Start()
	if err != nil {
		return nil, err
	}

	err = c.Wait()
	if err != nil {
		return nil, err
	}
	out.EndedAt = time.Now().UTC()
	out.Code = c.ProcessState.ExitCode()
	if c.TempFile != nil {
		out.TempFile = c.TempFile
	}

	return &out, nil
}

// Runs the command and waits for it to finish
// PsOutputs are inherited from the current process and
// are not captured
func (c *Cmd) Run() (*Result, error) {
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Stdin = os.Stdin
	var out Result
	out.FileName = c.Path
	out.Args = c.Args
	// use utc time
	out.StartedAt = time.Now().UTC()
	out.Stdout = make([]byte, 0)
	out.Stderr = make([]byte, 0)

	err := c.Start()
	if err != nil {
		out.EndedAt = time.Now().UTC()
		out.Code = 1
		return &out, err
	}

	err = c.Wait()
	if err != nil {
		out.EndedAt = time.Now().UTC()
		out.Code = 1
		return &out, err
	}

	out.EndedAt = time.Now().UTC()
	out.Code = c.ProcessState.ExitCode()
	if c.TempFile != nil {
		out.TempFile = c.TempFile
	}

	return &out, nil
}

// Runs the command and captures the PsOutput
// PsOutputs are captured from the current process and
// are not inherited
func (c *Cmd) Output() (*Result, error) {

	var out Result
	out.Stdout = make([]byte, 0)
	out.Stderr = make([]byte, 0)
	out.StartedAt = time.Now().UTC()
	out.FileName = c.Path
	out.Args = c.Args

	var outb, errb bytes.Buffer
	c.Stdout = &outb
	c.Stderr = &errb

	err := c.Start()
	if err != nil {
		out.EndedAt = time.Now().UTC()
		out.Code = 1
		return &out, err
	}

	err = c.Wait()
	if err != nil {
		out.EndedAt = time.Now().UTC()
		out.Code = 1
		return &out, err
	}

	out.EndedAt = time.Now().UTC()
	out.Code = c.ProcessState.ExitCode()
	out.Stdout = outb.Bytes()
	out.Stderr = errb.Bytes()
	if c.TempFile != nil {
		out.TempFile = c.TempFile
	}

	return &out, nil
}

func (c *Cmd) Start() error {
	if c.disableLogger {
		return c.Cmd.Start()
	}

	if c.logger != nil {
		c.logger(c)
	}

	if logger != nil {
		logger(c)
	}

	p := c.Path
	if p != "" && !filepath.IsAbs(p) {
		p2, err := Find(p, nil)
		if err == nil {
			c.Path = p2
		}
	}

	return c.Cmd.Start()
}

func (c *Cmd) Wait() error {
	return c.Cmd.Wait()
}
