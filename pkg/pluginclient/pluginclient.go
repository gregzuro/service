package pluginclient

import (
	"fmt"

	log "github.com/Sirupsen/logrus"

	"github.com/gregzuro/service/pkg/message"
	"github.com/gregzuro/service/pkg/remote"
	"github.com/gregzuro/service/switch/cmd/common"
	"github.com/gregzuro/service/switch/protobuf"
	"golang.org/x/net/context"
)

// Connect performs InitialContact and opens a StreamSGMS to the given address and port, returning any error.
func Connect(myEntityID, address string, port int) (message.Receiver, <-chan message.Message, error) {
	_, DialClients, err := common.SetUpDialer(address, int64(port))
	if err != nil {
		log.WithFields(log.Fields{
			"context": "SetUpDialer()"}).Error(err)
		return nil, nil, err
	}
	// TODO: initial contact
	gotoo, err := protobuf.InitialContactServiceClient.InitialContact(DialClients.Contact,
		context.Background(),
		&protobuf.Hello{
			CallerEntity: &protobuf.Entity{
				Id:      myEntityID,
				Kind:    "plugin",
				Address: "",
				Port:    0},
			CalledEntity: &protobuf.Entity{
				Id:      "",
				Kind:    "",
				Address: "",
				Port:    0},
		})
	if err != nil {
		log.WithFields(log.Fields{
			"context": "InitialContact"}).Error(err)
		return nil, nil, err
	}

	if gotoo.Code != 1 { // 1= stay with me, you're cool.
		log.WithFields(log.Fields{
			"context": "InitialContact",
			"gotoo":   gotoo}).Error(err)
		return nil, nil, fmt.Errorf("InitialContact: %v", gotoo)
	}
	log.WithFields(log.Fields{
		"context": "InitialContact",
		"gotoo":   gotoo}).Info()

	// parent := protobuf.Entity{Id: gotoo.Id, Kind: gotoo.Kind, Address: address, Port: int64(port)}

	// announce our identity
	log.WithFields(log.Fields{
		"context": "announce"}).Infof("I'm a 'plugin' called '%s' dialing '%s' on %d\n",
		myEntityID,
		address,
		port)

	ctx, _ := context.WithCancel(context.Background())
	send, in := remote.ClientStreamer(ctx, DialClients.SGMS)
	return send, in, nil

}
