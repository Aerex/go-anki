package io

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"

	shellQuote "github.com/kballard/go-shellquote"
	"github.com/rs/zerolog"

	"github.com/spf13/viper"
)

type ExecContext = func(name string, arg ...string) *exec.Cmd

type IO struct {
	Input       io.ReadCloser
	Output      io.Writer
	Error       io.Writer
	Log         *zerolog.Logger
	ExecContext ExecContext
}

func NewSystemIO() *IO {
	return &IO{
		Input:       os.Stdin,
		Output:      os.Stdout,
		Error:       os.Stdout,
		ExecContext: exec.Command,
	}
}

func NewTestIO(in *bytes.Buffer, out *bytes.Buffer, err *bytes.Buffer, execCtx ExecContext) *IO {
	return &IO{
		Input:       ioutil.NopCloser(in),
		Output:      out,
		Error:       err,
		ExecContext: execCtx,
	}
}

// Eval will evalute cmd and pipe the stdout to a provided buffer
func (i *IO) Eval(cmdString string, buf *bytes.Buffer) error {
	cmdSplit, err := shellQuote.Split(cmdString)
	if err != nil {
		return err
	}
	//cmd := i.ExecContext(cmdSplit[0], cmdSplit[1:]...)
	cmd := exec.Command(cmdSplit[0], cmdSplit[1:]...)
	// Set stdout to write buffer
	cmd.Stdout, cmd.Stderr, cmd.Stdin = os.Stdout, os.Stderr, os.Stdin
	if buf != nil {
		cmd.Stdout = buf
	}
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

// Get available editor program to use edit files
// If on Windows, default editor will be notepad.exe
// otherwise vim is used
func GetEditor() string {
	editor := viper.GetString("EDITOR")
	if editor == "" {
		if runtime.GOOS == "windows" {
			return "notepad.exe"
		}
		return "vim"
	}
	return editor
}
