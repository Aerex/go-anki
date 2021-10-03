package io

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
)

type IO struct {
  Input io.ReadCloser
  Output io.Writer 
  Error io.Writer
  Log *os.File
}

func NewSystemIO() *IO {
  return &IO{
    Input: os.Stdin,
    Output: os.Stdout,
    Error: os.Stdout,
  }
}

func NewTestIO(in *bytes.Buffer, out *bytes.Buffer, err *bytes.Buffer) *IO {
  return &IO {
    Input: ioutil.NopCloser(in),
    Output: out,
    Error: err,
  }
}

// Evalute cmd and return the output 
func Eval(cmdString string) (string, error) {
  buf := bytes.NewBufferString("");
  cmd := exec.Command(cmdString)
  // Set stdout to write buffer
  cmd.Stdout = buf
  cmd.Stderr = os.Stderr 
  if err := cmd.Run(); err != nil {
    return "", err
  }

  return buf.String(), nil
}
