package dispatcher

import (
	"io"
	"os"
	"os/exec"
)

type SilentStreamer struct {
	stdin io.Reader
}

func NewSilentStreamer(stdin io.Reader) Streamer {
	return &SilentStreamer{stdin: stdin}
}

func DefaultSilentStreamer() Streamer {
	return NewSilentStreamer(os.Stdin)
}

func (s *SilentStreamer) Prepare(cmd *exec.Cmd) error {
	cmd.Stdin = s.stdin
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	return nil
}

func (s *SilentStreamer) Finish(error) error {
	return nil
}
