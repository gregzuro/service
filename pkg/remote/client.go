package remote

import (
	log "github.com/Sirupsen/logrus"

	"io"

	"github.com/gregzuro/service/pkg/message"
	"github.com/gregzuro/service/switch/protobuf"
	"golang.org/x/net/context"
)

// ClientStreamer streams messages in both directions using the supplied SGMSServiceClient.
func ClientStreamer(ctx context.Context, c protobuf.SGMSServiceClient) (message.Receiver, <-chan message.Message) {

	// passing ctx makes the Stream cancellable
	stream, err := c.StreamSGMS(ctx)
	if err != nil {
		log.Fatalf("%v.StreamSGMS(_), %v", c, err)
	}

	// used to uniquely identify this source
	source := &stream
	recv := make(chan message.Message)
	go func() {
		defer func() {
			close(recv)
		}()
		for {
			m, err := stream.Recv()
			if err == io.EOF {
				return
			}
			if err != nil {
				// TODO: check for fatal, handle transient
				log.Fatalf("Failed to receive a message : %v", err)
			} else {
				mw := message.Wrap(m)
				// We set the source before sending it to the local message switch
				// so we can filter these messages on the other side.
				mw.SetSource(source)
				recv <- mw
			}
		}
	}()
	send := message.FuncReceiver(func(m message.Message) error {
		// Avoid message reflection.
		if m.Source() == source {
			//			log.Info("Skipping message to avoid reflection")
			return nil
		}
		if err := stream.Send(m.SGMS()); err != nil {
			// TODO: check for fatal, handle transient
			log.Fatalf("Failed to send a message: %v", err)
			return err
		}
		return nil
	})
	return send, recv
}

// ClientSendReceiver uses SGMSServiceClient to send messages.
// func ClientSendReceiver(c protobuf.SGMSServiceClient) message.Receiver {
// 	return message.FuncReceiver(func(m *message.Message) {

// 		c.SendSGMS(context.Background(), protobuf.SGMS(m))
// 	})
// }

// // ClientStreamReceiver uses SGMSServiceClient to stream messages.
// func ClientStreamReceiver(c protobuf.SGMSServiceClient) message.Receiver {
// 	stream, err := c.StreamSGMS(context.Background())
// 	if err != nil {
// 		// log.Fatalf("%v.RecordRoute(_) = _, %v", client, err)
// 	}
// 	return message.FuncReceiver(func(m *message.Message) {
// 		if err := stream.Send(m.SGMS); err != nil {
// 			// log.Fatalf("%v.Send(%v) = %v", stream, point, err)
// 		}
// 		reply, err := stream.CloseAndRecv()
// 		c.SendSGMS(context.Background(), m.SGMS)
// 	})
// }

// func ResponseStreamReceiver(stream protobuf.SGMSService_StreamSGMSServer) (context.Context, message.Receiver) {
// 	n.ListenMessages(ctx, message.FuncReceiver(func(m *message.Message) {
// 		err := stream.Send(m.SGMS)
// 		// TODO: cancel ctx if stream.Send returns an err indicating the connection has closed.
// 	}))

// }

// // SendReceiver returns a Receiver which makes a new connection for each message.
// func (c *ClientStreamReceiver) ReceiveMessage(m *protobuf.SGMS) {
// 	c.SendSGMS(context.Background(), m)
// }

// // ResponseStreamReceiver receives messages to the response stream of a StreamSGMS request.
// type ResponseStreamReceiver struct {
// }
