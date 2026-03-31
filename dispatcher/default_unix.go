//go:build unix

package dispatcher

import "context"

func DefaultDispatcher(ctx context.Context) Dispatcher {
	return NewUnixDispatcher(ctx).WithStream(DefaultPrettyStreamer())
}
