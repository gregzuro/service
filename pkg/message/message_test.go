package message_test

import (
	"github.com/gregzuro/service/pkg/message"
	"github.com/gregzuro/service/pkg/mocks"
)

func ExampleSGMSMessage() {
	// a *SGMSMessage can be passed to a function expecting a Message
	f := func(message.Message) {}
	f(message.Wrap((mocks.IntMsg)(1).SGMS()))
}
