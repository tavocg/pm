//go:build unix

package dispatcher

import (
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type UnixDispatcher struct {
	ctx    context.Context
	stream Streamer
}

type privilegeHelper struct {
	path string
	kind string
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
	for _, helper := range findPrivilegeHelpers() {
		return u.runCommand(nil, helper.path, "sh", "-c", command)
	}

	return u.runShell(command)
}

func (u *UnixDispatcher) runNonInteractivePrivileged(command string) error {
	if _, err := exec.LookPath("pkexec"); err == nil {
		return u.runCommand(nil, "pkexec", "sh", "-c", command)
	}

	helpers := findPrivilegeHelpers()
	if _, err := exec.LookPath("pinentry"); err == nil {
		sudoHelper, ok := findPinentryCapableHelper(helpers)
		if !ok {
			return u.runShell(command)
		}

		pin, pinErr := u.readPin()
		if pinErr != nil {
			return pinErr
		}

		return u.runCommand(strings.NewReader(pin+"\n"), sudoHelper.path, "-S", "-p", "", "-k", "sh", "-c", command)
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

func findPrivilegeHelpers() []privilegeHelper {
	helpers := make([]privilegeHelper, 0, 2)
	seen := map[string]struct{}{}

	for _, name := range []string{"doas", "sudo"} {
		path, err := exec.LookPath(name)
		if err != nil {
			continue
		}

		resolved := path
		if realPath, err := filepath.EvalSymlinks(path); err == nil && realPath != "" {
			resolved = realPath
		}

		if _, ok := seen[resolved]; ok {
			continue
		}
		seen[resolved] = struct{}{}

		helpers = append(helpers, privilegeHelper{
			path: resolved,
			kind: privilegeHelperKind(resolved),
		})
	}

	return helpers
}

func privilegeHelperKind(path string) string {
	switch filepath.Base(path) {
	case "sudo":
		return "sudo"
	case "doas", "opendoas":
		return "doas"
	default:
		return filepath.Base(path)
	}
}

func (h privilegeHelper) supportsStdinPassword() bool {
	return h.kind == "sudo"
}

func findPinentryCapableHelper(helpers []privilegeHelper) (privilegeHelper, bool) {
	for _, helper := range helpers {
		if helper.supportsStdinPassword() {
			return helper, true
		}
	}

	return privilegeHelper{}, false
}

func (u *UnixDispatcher) WithStream(streamer Streamer) Dispatcher {
	u.stream = streamer
	return u
}
