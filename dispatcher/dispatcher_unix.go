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
	stream := DefaultVerboseStreamer()
	return &UnixDispatcher{ctx: ctx, stream: stream}
}

func (u *UnixDispatcher) Run(cmd string) error {
	if cmd == "" {
		return nil
	}

	c := exec.CommandContext(u.ctx, "sh", "-c", cmd)
	if err := u.stream.Prepare(c); err != nil {
		return err
	}

	runErr := c.Run()
	if err := u.stream.Finish(runErr); err != nil {
		return err
	}
	return runErr
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
			if err := u.stream.Prepare(c); err != nil {
				return err
			}

			runErr := c.Run()
			if err := u.stream.Finish(runErr); err != nil {
				return err
			}
			return runErr
		}
	}

	return u.Run(cmd)
}

func (u *UnixDispatcher) WithStream(streamer Streamer) Dispatcher {
	u.stream = streamer
	return u
}
