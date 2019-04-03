package message_test

import (
	"reflect"
	"sync"
	"testing"

	"github.com/gregzuro/service/pkg/message"
	"github.com/gregzuro/service/pkg/mocks"
	"golang.org/x/net/context"
)

func sliceAppender(c <-chan message.Message, ms *[]message.Message, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		for m := range c {
			*ms = append(*ms, m)
		}
		wg.Done()
	}()
}

func TestTee(t *testing.T) {
	in := make(chan message.Message, 1)
	Copy := message.Tee(in)
	wg := sync.WaitGroup{}

	x := []message.Message{mocks.IntMsg(1), mocks.IntMsg(2), mocks.IntMsg(3)}
	a := []message.Message{}
	b := []message.Message{}
	c := []message.Message{}
	ctx, cancel := context.WithCancel(context.Background())
	sliceAppender(Copy(ctx), &a, &wg)
	sliceAppender(Copy(context.Background()), &b, &wg)
	in <- x[0]
	cancel()
	in <- x[1]
	sliceAppender(Copy(context.Background()), &c, &wg)
	in <- x[2]
	close(in)

	wg.Wait()
	// a may have the first 0, 1, or 2 messages (!)
	// TODO: figure out a better way to test cancellation.
	if !reflect.DeepEqual(a, x[:len(a)]) {
		t.Fail()
	}
	if !reflect.DeepEqual(b, x) {
		t.Fail()
	}
	// Similarly, c may have the last 0, 1, or 2 messages.
	if !reflect.DeepEqual(c, x[len(x)-len(c):]) {
		t.Fail()
	}
}

func TestRingBuffer(t *testing.T) {
	in := make(chan message.Message)
	out := message.RingBuffer(in, 2)

	defer func() { close(in) }()

	in <- mocks.IntMsg(1)
	in <- mocks.IntMsg(2)
	in <- mocks.IntMsg(3)
	<-out
	<-out
	l := len(out)
	if l != 0 {
		t.Error("len(out): expected 0, got ", l)
		t.Fail()
	}

}
