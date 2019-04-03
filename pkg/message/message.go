package message

import (
	"github.com/gregzuro/service/switch/protobuf"
)

type Source interface{}

// Message represents a general message which is moving through our routing system.
type Message interface {
	SGMS() *protobuf.SGMS
	// Get the message source, to allow reflection prevention.
	Source() Source
	SetSource(Source)
}

// SGMSMessage is a wrapper type around protobuf.SGMS.
type SGMSMessage struct {
	msg    *protobuf.SGMS
	source Source
}

// Wrap a protobuf.SGMS in a SGMSMessage.
func Wrap(s *protobuf.SGMS) *SGMSMessage {
	return &SGMSMessage{
		msg: s,
	}
}

func (m *SGMSMessage) SGMS() *protobuf.SGMS {
	return m.msg
}

func (m *SGMSMessage) Source() Source {
	return m.source
}

func (m *SGMSMessage) SetSource(s Source) {
	m.source = s
}
