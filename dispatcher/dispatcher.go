// Package dispatcher
package dispatcher

import (
	"os"
)

type Dispatcher interface {
	Run(cmd string) error
	WithStream(stream Streamer) Dispatcher
}

type Streamer struct {
	Stdout *os.File
	Stderr *os.File
}
