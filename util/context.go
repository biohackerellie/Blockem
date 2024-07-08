package util

import "context"

func CtxSend[T any](ctx context.Context, ch chan T, val T) (ok bool) {
	if ctx == nil || ch == nil || ctx.Err() != nil {
		ok = false
		return
	}

	defer func() {
		if err := recover(); err != nil {
			ok = false
		}
	}()

	select {
	case <-ctx.Done():
		ok = false
	case ch <- val:
		ok = true
	}
	return
}
