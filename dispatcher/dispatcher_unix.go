//go:build unix

package dispatcher

import (
	"context"
	"os"
	"os/exec"
)

type UnixDispatcher struct {
	ctx    context.Context
	stream Streamer
}

func NewUnixDispatcher(ctx context.Context) *UnixDispatcher {
	stream := Streamer{
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
	return &UnixDispatcher{ctx: ctx, stream: stream}
}

func (u *UnixDispatcher) Run(cmd string) error {
	if cmd == "" {
		return nil
	}

	c := exec.CommandContext(u.ctx, "sh", "-c", cmd)
	c.Stdin = u.stream.Stdin
	c.Stdout = u.stream.Stdout
	c.Stderr = u.stream.Stderr
	return c.Run()
}

func (u *UnixDispatcher) RunAsPrivileged(cmd string) error {
	if cmd == "" {
		return nil
	}

	if os.Geteuid() == 0 {
		return u.Run(cmd)
	}

	for _, helper := range []string{"doas", "sudo"} {
		if _, err := exec.LookPath(helper); err == nil {
			c := exec.CommandContext(u.ctx, helper, "sh", "-c", cmd)
			c.Stdin = u.stream.Stdin
			c.Stdout = u.stream.Stdout
			c.Stderr = u.stream.Stderr
			return c.Run()
		}
	}

	return u.Run(cmd)
}

func (u *UnixDispatcher) WithStream(stream Streamer) Dispatcher {
	u.stream = stream
	return u
}
