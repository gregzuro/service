package mobile

import (
	"context"
	"runtime/debug"

	log "github.com/Sirupsen/logrus"
	"github.com/golang/protobuf/proto"
	"github.com/gregzuro/service/switch/cmd/common"
	"github.com/gregzuro/service/switch/cmd/device"
	pb "github.com/gregzuro/service/switch/protobuf"
)

var (
	node *device.Node
)

// StartSwitch starts the device switch.
func StartSwitch(masterAddress string, masterPort int64, listenPort int64, deviceID string, locBytes []byte) {
	// if the switch crashes, log stack trace and restart
	defer func() {
		if e := recover(); e != nil {
			log.WithFields(log.Fields{"recover": e, "stack": debug.Stack()})
			StartSwitch(masterAddress, masterPort, listenPort, deviceID, locBytes)
		}
	}()
	node = device.New(common.CommandArgs{
		MasterAddress:     masterAddress,
		MasterPort:        masterPort,
		ListenPort:        listenPort,
		ShortName:         "sgms_device_" + deviceID,
		StreamMessages:    false,
		PortMappedCluster: true,
	})
	// assign initial location if given
	loc := &pb.Location{}
	if err := proto.Unmarshal(locBytes, loc); err != nil {
		loc = nil
	}
	node.InitialLocation = loc
	// start and listen
	if err := node.Go(); err != nil {
		log.Fatal(err)
	}
}

// SendSGMS takes a SGMS serialized byte array and calls the SendSGMS method for the device node.  This method's purpose is to avoid TLS overhead and complexity.
func SendSGMS(b []byte) error {
	sgms := &pb.SGMS{}
	if err := proto.Unmarshal(b, sgms); err != nil {
		return err
	}
	_, err := node.SendSGMS(context.Background(), sgms)
	return err
}
