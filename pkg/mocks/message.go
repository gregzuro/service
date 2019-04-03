package mocks

import (
	"github.com/gregzuro/service/pkg/message"
	"github.com/gregzuro/service/switch/protobuf"
)

// IntMsg implements message.Message using an int.
type IntMsg uint64

func (i IntMsg) SGMS() *protobuf.SGMS {
	return &protobuf.SGMS{
		HeartBeat: &protobuf.HeartBeat{
			// cast pointer, then take value
			Counter: (uint64)(i),
		},
	}
}

func (m IntMsg) Source() message.Source {
	return ""
}

func (m IntMsg) SetSource(s message.Source) {
}
