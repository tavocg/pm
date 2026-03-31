//go:build !unix

package dispatcher

import (
	"context"
	"fmt"
	"runtime"
)

type unsupportedDispatcher struct {
	err error
}

func DefaultDispatcher(context.Context) Dispatcher {
	return unsupportedDispatcher{
		err: fmt.Errorf("no default dispatcher for %s/%s", runtime.GOOS, runtime.GOARCH),
	}
}

func (u unsupportedDispatcher) Run(string) error {
	return u.err
}

func (u unsupportedDispatcher) WithStream(Streamer) Dispatcher {
	return u
}
