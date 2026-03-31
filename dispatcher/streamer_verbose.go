package dispatcher

import (
	"io"
	"os"
	"os/exec"
)

type VerboseStreamer struct {
	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer
}

func NewVerboseStreamer(stdin io.Reader, stdout io.Writer, stderr io.Writer) Streamer {
	return &VerboseStreamer{
		stdin:  stdin,
		stdout: stdout,
		stderr: stderr,
	}
}

func DefaultVerboseStreamer() Streamer {
	return NewVerboseStreamer(os.Stdin, os.Stdout, os.Stderr)
}

func (s *VerboseStreamer) Prepare(cmd *exec.Cmd) error {
	cmd.Stdin = s.stdin
	cmd.Stdout = s.stdout
	cmd.Stderr = s.stderr
	return nil
}

func (s *VerboseStreamer) Finish(error) error {
	return nil
}
