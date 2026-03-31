package dispatcher

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"sync"
)

const (
	prettyMaxLines = 5
	ansiGray       = "\x1b[90m"
	ansiReset      = "\x1b[0m"
)

var ansiPattern = regexp.MustCompile(`\x1b\[[0-9;?]*[ -/]*[@-~]`)

type PrettyStreamer struct {
	stdin       io.Reader
	output      io.Writer
	interactive bool

	mu       sync.Mutex
	lines    []string
	partial  bytes.Buffer
	rendered int
	writer   io.Writer
}

func NewPrettyStreamer(stdin io.Reader, output *os.File) Streamer {
	return newPrettyStreamer(stdin, output, isTerminal(output))
}

func DefaultPrettyStreamer() Streamer {
	return NewPrettyStreamer(os.Stdin, os.Stdout)
}

func newPrettyStreamer(stdin io.Reader, output io.Writer, interactive bool) *PrettyStreamer {
	s := &PrettyStreamer{
		stdin:       stdin,
		output:      output,
		interactive: interactive,
	}
	s.writer = &prettyWriter{streamer: s}
	return s
}

func (s *PrettyStreamer) Prepare(cmd *exec.Cmd) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.lines = nil
	s.partial.Reset()
	s.rendered = 0

	cmd.Stdin = s.stdin
	cmd.Stdout = s.writer
	cmd.Stderr = s.writer
	return nil
}

func (s *PrettyStreamer) Finish(runErr error) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.interactive {
		s.clearRenderedLocked()
	}

	if runErr != nil {
		for _, line := range s.visibleLinesLocked() {
			if _, err := fmt.Fprintln(s.output, line); err != nil {
				return err
			}
		}
	}

	s.lines = nil
	s.partial.Reset()
	s.rendered = 0
	return nil
}

func (s *PrettyStreamer) write(data []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, b := range data {
		switch b {
		case '\n':
			s.pushLineLocked(s.partial.String())
			s.partial.Reset()
		case '\r':
			s.partial.Reset()
		default:
			s.partial.WriteByte(b)
		}
	}

	if s.interactive {
		if err := s.renderLocked(); err != nil {
			return 0, err
		}
	}

	return len(data), nil
}

func (s *PrettyStreamer) pushLineLocked(line string) {
	s.lines = append(s.lines, sanitizeLine(line))
	if len(s.lines) > prettyMaxLines {
		s.lines = s.lines[len(s.lines)-prettyMaxLines:]
	}
}

func (s *PrettyStreamer) visibleLinesLocked() []string {
	lines := append([]string(nil), s.lines...)
	if s.partial.Len() > 0 {
		lines = append(lines, sanitizeLine(s.partial.String()))
	}
	if len(lines) > prettyMaxLines {
		lines = lines[len(lines)-prettyMaxLines:]
	}
	return lines
}

func (s *PrettyStreamer) renderLocked() error {
	lines := s.visibleLinesLocked()
	if err := s.clearRenderedLocked(); err != nil {
		return err
	}
	if len(lines) == 0 {
		return nil
	}

	for i, line := range lines {
		if _, err := fmt.Fprintf(s.output, "\r\x1b[2K%s%s%s", ansiGray, line, ansiReset); err != nil {
			return err
		}
		if i < len(lines)-1 {
			if _, err := fmt.Fprint(s.output, "\n"); err != nil {
				return err
			}
		}
	}

	s.rendered = len(lines)
	return nil
}

func (s *PrettyStreamer) clearRenderedLocked() error {
	if s.rendered == 0 {
		return nil
	}

	if s.rendered > 1 {
		if _, err := fmt.Fprintf(s.output, "\x1b[%dA", s.rendered-1); err != nil {
			return err
		}
	}

	for i := 0; i < s.rendered; i++ {
		if _, err := fmt.Fprint(s.output, "\r\x1b[2K\x1b[M"); err != nil {
			return err
		}
	}

	s.rendered = 0
	return nil
}

func sanitizeLine(line string) string {
	return ansiPattern.ReplaceAllString(line, "")
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

type prettyWriter struct {
	streamer *PrettyStreamer
}

func (w *prettyWriter) Write(data []byte) (int, error) {
	return w.streamer.write(data)
}
