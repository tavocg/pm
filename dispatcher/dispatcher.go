// Package dispatcher
package dispatcher

type Dispatcher interface {
	Cmd(command string) *DispatcherCmd
	Run(cmd *DispatcherCmd) error
	WithStream(streamer Streamer) Dispatcher
}

type DispatcherCmd struct {
	command     string
	privileged  bool
	interactive bool
}

func NewDispatcherCmd(command string) *DispatcherCmd {
	return &DispatcherCmd{command: command}
}

func (c *DispatcherCmd) WithPrivileged() *DispatcherCmd {
	if c != nil {
		c.privileged = true
	}

	return c
}

func (c *DispatcherCmd) WithInteractive() *DispatcherCmd {
	if c != nil {
		c.interactive = true
	}

	return c
}

func (c *DispatcherCmd) Command() string {
	if c == nil {
		return ""
	}

	return c.command
}

func (c *DispatcherCmd) Privileged() bool {
	return c != nil && c.privileged
}

func (c *DispatcherCmd) Interactive() bool {
	return c != nil && c.interactive
}
