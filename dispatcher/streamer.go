package dispatcher

import "os"

type Streamer struct {
	Stdin  *os.File
	Stdout *os.File
	Stderr *os.File
}
