//go:build unix

package dispatcher

import (
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"strings"
)

type UnixDispatcher struct {
	ctx    context.Context
	stream Streamer
}

func NewUnixDispatcher(ctx context.Context) *UnixDispatcher {
	stream := DefaultVerboseStreamer()
	return &UnixDispatcher{ctx: ctx, stream: stream}
}

func (u *UnixDispatcher) Cmd(command string) *DispatcherCmd {
	return NewDispatcherCmd(command)
}

func (u *UnixDispatcher) Run(cmd *DispatcherCmd) error {
	if cmd == nil || cmd.Command() == "" {
		return nil
	}

	if !cmd.Privileged() || os.Geteuid() == 0 {
		return u.runShell(cmd.Command())
	}

	if cmd.Interactive() {
		return u.runInteractivePrivileged(cmd.Command())
	}

	return u.runNonInteractivePrivileged(cmd.Command())
}

func (u *UnixDispatcher) runShell(command string) error {
	return u.runCommand(nil, "sh", "-c", command)
}

func (u *UnixDispatcher) runInteractivePrivileged(command string) error {
	for _, helper := range []string{"doas", "sudo"} {
		if _, err := exec.LookPath(helper); err == nil {
			return u.runCommand(nil, helper, "sh", "-c", command)
		}
	}

	return u.runShell(command)
}

func (u *UnixDispatcher) runNonInteractivePrivileged(command string) error {
	if _, err := exec.LookPath("pkexec"); err == nil {
		return u.runCommand(nil, "pkexec", "sh", "-c", command)
	}

	if _, err := exec.LookPath("pinentry"); err == nil {
		pin, pinErr := u.readPin()
		if pinErr != nil {
			return pinErr
		}

		if _, err := exec.LookPath("sudo"); err == nil {
			return u.runCommand(strings.NewReader(pin+"\n"), "sudo", "-S", "-p", "", "sh", "-c", command)
		}
	}

	return u.runShell(command)
}

func (u *UnixDispatcher) runCommand(stdin io.Reader, name string, args ...string) error {
	c := exec.CommandContext(u.ctx, name, args...)
	if err := u.stream.Prepare(c); err != nil {
		return err
	}
	if stdin != nil {
		c.Stdin = stdin
	}

	runErr := c.Run()
	if err := u.stream.Finish(runErr); err != nil {
		return err
	}
	return runErr
}

func (u *UnixDispatcher) readPin() (string, error) {
	c := exec.CommandContext(u.ctx, "pinentry")
	c.Stdin = strings.NewReader("GETPIN\n")

	out, err := c.Output()
	if err != nil {
		return "", err
	}

	pin := parsePinentryOutput(string(out))
	if pin == "" {
		return "", errors.New("pinentry did not return a secret")
	}

	return pin, nil
}

func parsePinentryOutput(output string) string {
	lines := strings.Split(strings.TrimRight(output, "\r\n"), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "D ") {
			return strings.TrimPrefix(line, "D ")
		}
	}

	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimRight(lines[i], "\r")
		if line == "" || strings.HasPrefix(line, "OK") || strings.HasPrefix(line, "ERR") {
			continue
		}
		return line
	}

	return ""
}

func (u *UnixDispatcher) WithStream(streamer Streamer) Dispatcher {
	u.stream = streamer
	return u
}
