// Package dispatcher
package dispatcher

import (
	"os"
)

type Dispatcher interface {
	Run(cmd string) error
	RunAsPrivileged(cmd string) error
	WithStream(stream Streamer) Dispatcher
}

type Streamer struct {
	Stdin  *os.File
	Stdout *os.File
	Stderr *os.File
}
