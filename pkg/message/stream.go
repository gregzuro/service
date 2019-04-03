package message

import (
	// Keep an eye on https://github.com/golang/go/issues/18130 (for Go 1.9)
	"golang.org/x/net/context"
)

// RingBuffer reads messages from in and writes them to out.
// Whenever out is full, it will remove the oldest message to make room.
// Adapted from https://blog.pivotal.io/labs/labs/a-concurrent-ring-buffer-for-go.
func RingBuffer(in <-chan Message, size int) <-chan Message {
	out := make(chan Message, size)
	go func() {
		defer close(out)
		for m := range in {
			select {
			case out <- m:
			default:
				// If out is full, read a message from the end to make room.
				select {
				case <-out:
				default:
					// Avoid a deadlock in case the buffer has since been drained.
				}
				out <- m
			}
		}
	}()
	return out
}

// Nothing to see here, move along..
// sink is an internal implementation detail of Tee.
type sink struct {
	ctx context.Context
	ch  chan<- Message
}

// Tee(in) kicks off a goroutine reading messages from in and copying them
// to each of an (initially empty) collection of out channels.

// Each call to the returned function adds a new out channel to the
// processing loop for subsequent messages. If the context associated with an
// out channel is cancelled, that channel is eventually closed.

// Closing in closes all outs and exits the goroutine.

// Note that Tee does not pass Context up: it is guaranteed to continue processing
// messages until in is closed, even if all outs are cancelled.

// Message delivery to out channels is:

//  - Synchronized (a blocked out channel halts processing)
//  - Order-preserving

func Tee(in <-chan Message) func(context.Context) <-chan Message {
	// outs could just be a map, which would make send order nondeterministic.
	var outs []sink
	closing := false

	go func() {
		// when in closes, close all outs
		defer func() {
			closing = true
			for _, c := range outs {
				close(c.ch)
			}
		}()

		for m := range in {
			offset := 0 // deletes offset
			// attempt to send to each sink
			for i, _ := range outs {
				// If both cases are available Go will randomly choose,
				// making the number of messages received after context
				// cancellation non-deterministic.
				// This could be worked around using a default case, but that
				// would introduce potential race conditions.
				select {
				case <-outs[i-offset].ctx.Done():
					// // Close the downstream channel.
					close(outs[i-offset].ch)
					// Delete current element.
					// https://github.com/golang/go/wiki/SliceTricks
					outs = append(outs[:i-offset], outs[i-offset+1:]...)
					offset++
				// ensure closing source succeeds eventually.
				case outs[i-offset].ch <- m:
				}
			}
		}
	}()
	return func(ctx context.Context) <-chan Message {
		// A small output buffer allows writes to in to succeed immediately when outs are
		// being drained quickly.
		ch := make(chan Message, 2)
		if closing {
			close(ch)
		} else {
			// this won't interfere with the main loop's sinks access
			// since it is append-only.
			outs = append(outs, sink{
				ctx: ctx,
				ch:  ch,
			})
		}
		return ch
	}
}
