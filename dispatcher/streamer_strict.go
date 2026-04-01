package dispatcher

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	ansiGray  = "\x1b[90m"
	ansiReset = "\x1b[0m"
)

type StrictStreamer struct {
	stdin  io.Reader
	stderr *os.File
	label  string

	stdoutBuf bytes.Buffer
	stderrBuf bytes.Buffer
}

func NewStrictStreamer(stdin io.Reader, stderr *os.File) Streamer {
	return &StrictStreamer{
		stdin:  stdin,
		stderr: stderr,
	}
}

func DefaultStrictStreamer() Streamer {
	return NewStrictStreamer(os.Stdin, os.Stderr)
}

func (s *StrictStreamer) Prepare(cmd *exec.Cmd) error {
	s.stdoutBuf.Reset()
	s.stderrBuf.Reset()
	s.label = commandLabel(cmd)

	cmd.Stdin = s.stdin
	cmd.Stdout = &s.stdoutBuf
	cmd.Stderr = &s.stderrBuf
	return nil
}

func (s *StrictStreamer) Finish(runErr error) error {
	if runErr == nil || s.stderrBuf.Len() == 0 {
		return nil
	}

	target := s.stderr
	if target == nil {
		target = os.Stderr
	}

	if !isTerminal(target) {
		return writePrefixed(target, s.label, s.stderrBuf.Bytes())
	}

	if _, err := fmt.Fprint(target, ansiGray); err != nil {
		return err
	}
	if err := writePrefixed(target, s.label, s.stderrBuf.Bytes()); err != nil {
		return err
	}
	_, err := fmt.Fprint(target, ansiReset)
	return err
}

func commandLabel(cmd *exec.Cmd) string {
	if cmd == nil || len(cmd.Args) == 0 {
		return ""
	}

	raw := cmd.Args[len(cmd.Args)-1]
	fields := strings.Fields(raw)
	if len(fields) == 0 {
		return ""
	}

	return filepath.Base(fields[0])
}

func writePrefixed(w io.Writer, label string, data []byte) error {
	prefix := ""
	if label != "" {
		prefix = "[" + label + "] "
	}

	if prefix == "" {
		_, err := w.Write(data)
		return err
	}

	for len(data) > 0 {
		line := data
		hasNewline := false

		if idx := bytes.IndexByte(data, '\n'); idx >= 0 {
			line = data[:idx]
			data = data[idx+1:]
			hasNewline = true
		} else {
			data = nil
		}

		if _, err := io.WriteString(w, prefix); err != nil {
			return err
		}
		if len(line) > 0 {
			if _, err := w.Write(line); err != nil {
				return err
			}
		}
		if hasNewline {
			if _, err := io.WriteString(w, "\n"); err != nil {
				return err
			}
		}
	}

	return nil
}

func isTerminal(file *os.File) bool {
	if file == nil {
		return false
	}

	info, err := file.Stat()
	if err != nil {
		return false
	}

	return info.Mode()&os.ModeCharDevice != 0
}
