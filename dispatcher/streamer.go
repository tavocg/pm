package dispatcher

import "os/exec"

type Streamer interface {
	Prepare(cmd *exec.Cmd) error
	Finish(runErr error) error
}
