package dispatcher

import (
	"context"
	"os"
	"os/exec"
	"strings"
)

type UnixDispatcher struct {
	ctx    context.Context
	stream Streamer
}

func NewUnixDispatcher(ctx context.Context) *UnixDispatcher {
	stream := Streamer{
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}
	return &UnixDispatcher{ctx: ctx, stream: stream}
}

func (u *UnixDispatcher) Run(cmd string) error {
	name, arg, _ := strings.Cut(cmd, " ")
	c := exec.CommandContext(u.ctx, name, arg)
	c.Stdout = u.stream.Stdout
	c.Stderr = u.stream.Stderr
	return c.Run()
}

func (u *UnixDispatcher) WithStream(stream Streamer) Dispatcher {
	u.stream = stream
	return u
}
