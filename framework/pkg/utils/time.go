package utils

import (
	"context"
	"time"
)

// Shrink will decrease the duration by comparing with context's timeout duration
// and return new timeout\context\CancelFunc.
func Shrink(c context.Context, d time.Duration) (time.Duration, context.Context, context.CancelFunc) {
	if deadline, ok := c.Deadline(); ok {
		if ctimeout := time.Until(deadline); ctimeout < d {
			// deliver small timeout
			return ctimeout, c, func() {}
		}
	}
	ctx, cancel := context.WithTimeout(c, d)
	return d, ctx, cancel
}
