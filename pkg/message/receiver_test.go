package message_test

import (
	"fmt"
	"testing"

	"github.com/gregzuro/service/pkg/message"
	"github.com/gregzuro/service/pkg/mocks"
	"golang.org/x/net/context"
)

func ExampleFuncReceiver() {
	fr := message.FuncReceiver(func(m message.Message) error {
		fmt.Println(m)
		return nil
	})
	fr.ReceiveMessage(mocks.IntMsg(1))
	// Output:
	// 1
}

func TestBufferedReceiver(t *testing.T) {
	c := make(chan message.Message, 0)

	// r blocks until something empties the channel.
	r := message.FuncReceiver(func(m message.Message) error {
		c <- m
		return nil
	})
	br := message.BufferedReceiver(r, 2)

	// This would block indefinitely with an unbuffered receiver.
	br.ReceiveMessage(mocks.IntMsg(1))
	br.ReceiveMessage(mocks.IntMsg(2))

	m1, m2 := <-c, <-c
	if m1 != mocks.IntMsg(1) || m2 != mocks.IntMsg(2) {
		t.Fatal(fmt.Sprintf("%v,%v != 1,2", m1, m2))
	}
}

func ExampleFuncListenable() {
	printReceiver := message.FuncReceiver(func(m message.Message) error {
		fmt.Println(m)
		return nil
	})

	fs := message.FuncListenable(func(ctx context.Context, r message.Receiver) {
		r.ReceiveMessage(mocks.IntMsg(1))
	})
	fs.ListenMessages(context.Background(), printReceiver)
	// Output:
	// 1
}

func TestBufferedListenable(t *testing.T) {
	c := make(chan message.Message, 0)

	rs := message.Receivers{}
	bl := message.BufferedListenable(&rs, 2)

	bl.ListenMessages(context.Background(), message.FuncReceiver(func(m message.Message) error {
		c <- m
		return nil
	}))
	// This would block indefinitely with an unbuffered receiver.
	rs.ReceiveMessage(mocks.IntMsg(1))
	rs.ReceiveMessage(mocks.IntMsg(2))

	m1, m2 := <-c, <-c
	if m1 != mocks.IntMsg(1) || m2 != mocks.IntMsg(2) {
		t.Fatal(fmt.Sprintf("%v,%v != 1,2", m1, m2))
	}
}

// type callbackReceiver func(m *protobuf.SGMS)

// func (cb callbackReceiver) ReceiveMessage(m *protobuf.SGMS) {
// 	cb(m)
// }

func printReceiver(label string) message.Receiver {
	return message.FuncReceiver(func(m message.Message) error {
		fmt.Println(label, m.SGMS().HeartBeat.Counter)
		return nil
	})
}

func ExampleReceivers() {
	// Create a new Receivers with one initial listener.
	rs := message.Receivers{printReceiver("A")}
	// Add a second listener.
	rs.ListenMessages(context.Background(), printReceiver("B"))
	rs.ReceiveMessage(mocks.IntMsg(1))
	// Output:
	// A 1
	// B 1
}

func ExampleNewReceiverListenablePair() {
	r, l := message.NewReceiverListenablePair()
	l.ListenMessages(context.Background(), printReceiver("A"))
	r.ReceiveMessage(mocks.IntMsg(1))
	// Output:
	// A 1
}
