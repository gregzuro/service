package message

import (
	"golang.org/x/net/context"
)

// Receiver specifies an interface for receiving messages.
type Receiver interface {
	// TODO: define an error enum type for message delivery
	ReceiveMessage(Message) error
}

// FuncReceiver uses a function to satisfy the Receiver interface.
type FuncReceiver func(Message) error

func (f FuncReceiver) ReceiveMessage(m Message) error { return f(m) }

// BufferedReceiver wraps a Receiver with a buffer.
func BufferedReceiver(r Receiver, size int) Receiver {
	buf := make(chan Message, size)
	br := FuncReceiver(func(m Message) error {
		buf <- m
		return nil
	})
	go func() {
		for n := range buf {
			r.ReceiveMessage(n)
		}
	}()
	return br
}

// Listenable sends messages to Receivers registered via the ListenMessages method until
// their context is cancelled.
// Note that the caller is responsible for cleaning up any Receiver-associate goroutines
// when cancelling ctx.
type Listenable interface {
	ListenMessages(ctx context.Context, r Receiver)
}

// FuncListenable implements Listenable with a function.
type FuncListenable func(context.Context, Receiver)

func (f FuncListenable) ListenMessages(ctx context.Context, r Receiver) { f(ctx, r) }

// BufferedListenable puts a buffer on each added receiver.
func BufferedListenable(ra Listenable, bufSize int) Listenable {
	return FuncListenable(func(ctx context.Context, r Receiver) {
		ra.ListenMessages(ctx, BufferedReceiver(r, bufSize))
	})
}

// Receivers implements both the Receiver and Listenable interfaces.
type Receivers []Receiver

func (rs *Receivers) ListenMessages(ctx context.Context, r Receiver) {
	*rs = append(*rs, r)
	// TODO: remove r when ctx is cancelled
}

func (rs *Receivers) ReceiveMessage(m Message) {
	for _, r := range *rs {
		r.ReceiveMessage(m)
	}
}

// NewReceiverListenablePair returns a Receiver with a linked Listenable.
func NewReceiverListenablePair() (Receiver, Listenable) {
	in := make(chan Message)
	tee := Tee(in)
	var r FuncReceiver = func(m Message) error {
		in <- m
		return nil
	}
	var l FuncListenable = func(ctx context.Context, r Receiver) {
		q := tee(ctx)
		go func() {
			for m := range q {
				r.ReceiveMessage(m)
			}
			// We're ignoring the channel closure.
			// ListenMessages could take a Done func or something.
		}()
	}
	return r, l
}

func ReceiveFromChan(in <-chan Message, r Receiver, cancel context.CancelFunc) {
	defer cancel()
	for m := range in {
		r.ReceiveMessage(m)
	}
}
