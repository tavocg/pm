// Package dispatcher
package dispatcher

type Dispatcher interface {
	Run(cmd string) error
	RunAsPrivileged(cmd string) error
	WithStream(stream Streamer) Dispatcher
}
