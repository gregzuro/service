package device

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	log "github.com/Sirupsen/logrus"
	"golang.org/x/net/context"

	"github.com/gregzuro/service/switch/cmd/common"
	"github.com/gregzuro/service/switch/db"
	"github.com/gregzuro/service/switch/protobuf"
)

type Node struct {
	*common.NodeCommon
	TSDB            db.TSDB
	InitialLocation *protobuf.Location
}

func New(commandArgs common.CommandArgs) *Node {
	node := Node{NodeCommon: common.NewNodeCommon(commandArgs, "device")}
	return &node
}

func (n *Node) Go() error {

	var err error

	// find our parent
	log.WithFields(log.Fields{
		"context": "startup"}).Infof("asking master %s:%d for our parent...", n.CommandArgs.MasterAddress, n.CommandArgs.MasterPort)
	n.CommandArgs.ParentAddress, n.CommandArgs.ParentPort, err = common.GetParentForDevice(n.CommandArgs, n.Entity)
	log.WithFields(log.Fields{
		"context": "GetParentForDevice"}).Infof(" ... parent is: %s:%v\n", n.CommandArgs.ParentAddress, n.CommandArgs.ParentPort)

	// always listen
	listenConn, listenSrv, err := n.SetUpListener()
	if err != nil {
		log.WithFields(log.Fields{
			"context": "SetUpListener"}).Error(err)
		return err
	}

	// devices always dial a master to find a slave
	_, n.DialClients, err = common.SetUpDialer(n.CommandArgs.MasterAddress, n.CommandArgs.MasterPort)
	if err != nil {
		log.WithFields(log.Fields{
			"context": "GetEntityId"}).Error(err)
		return err
	}

	// contact master
	_goto1, err := protobuf.InitialContactServiceClient.InitialContact(n.DialClients.Contact,
		context.Background(),
		&protobuf.Hello{
			CallerEntity: &protobuf.Entity{
				Id:      n.Entity.Id,
				Kind:    n.Entity.Kind,
				Address: "sgms_" + n.CommandArgs.ShortName, // TODO(greg) this is a kludge for docker
				Port:    n.Entity.Port},
			CalledEntity: &protobuf.Entity{
				Id:      n.Entity.Id,
				Kind:    "master",
				Address: n.CommandArgs.MasterAddress,
				Port:    n.CommandArgs.MasterPort},
			Location: n.InitialLocation,
		})
	if err != nil {
		//	logxx.Prixxntln(err.Error())
		log.WithFields(log.Fields{
			"context": "InitialContact(1of2)"}).Error(err)
		return err
	}

	if _goto1.Code != 2 { // 2= connect to the specified server instead
		log.WithFields(log.Fields{
			"context": "InitialContact(1of2)",
			"_goto1":  _goto1}).Error(err)
		return fmt.Errorf("InitialContact(1of2): %v", _goto1)
	}
	log.WithFields(log.Fields{
		"context": "InitialContact(1of2)",
		"_goto1":  _goto1}).Info()

	n.CommandArgs.ParentAddress = _goto1.Address
	n.CommandArgs.ParentPort = _goto1.Port

	// if PortMappedCluster, reuse MasterAddress for parent connection
	if n.CommandArgs.PortMappedCluster {
		n.CommandArgs.ParentAddress = n.CommandArgs.MasterAddress

	}
	// now dial the slave
	_, n.DialClients, err = common.SetUpDialer(n.CommandArgs.ParentAddress, n.CommandArgs.ParentPort)
	if err != nil {
		log.WithFields(log.Fields{
			"context": "SetUpDialer()"}).Error(err)
		return err
	}

	_goto2, err := protobuf.InitialContactServiceClient.InitialContact(n.DialClients.Contact,
		context.Background(),
		&protobuf.Hello{
			CallerEntity: &protobuf.Entity{
				Id:      n.Entity.Id,
				Kind:    n.Entity.Kind,
				Address: n.CommandArgs.ShortName, // TODO(greg) this is a kludge for docker
				Port:    n.Entity.Port},
			CalledEntity: &protobuf.Entity{
				Id:      n.Entity.Id,
				Kind:    "master",
				Address: n.CommandArgs.MasterAddress,
				Port:    n.CommandArgs.MasterPort},
			Location: n.InitialLocation,
		})
	if err != nil {
		log.WithFields(log.Fields{
			"context": "InitialContact(2of2)"}).Error(err)
		return err
	}

	if _goto2.Code != 1 { // 1= stay with me, you're cool.
		log.WithFields(log.Fields{
			"context": "InitialContact(2of2)",
			"_goto2":  _goto2}).Error(err)
		return fmt.Errorf("InitialContact(2of2): %v", _goto2)
	}
	log.WithFields(log.Fields{
		"context": "InitialContact(2of2)",
		"_goto2":  _goto2}).Info()

	n.Parent = protobuf.Entity{Id: _goto2.Id, Kind: _goto2.Kind, Address: n.CommandArgs.ParentAddress, Port: n.CommandArgs.ParentPort}

	// announce our identity
	log.WithFields(log.Fields{
		"context": "announce"}).Infof("I'm a '%s' switch called '%s' dialing '%s' on %d and listening on %d\n",
		n.Entity.Kind,
		n.CommandArgs.ShortName,
		n.CommandArgs.ParentAddress,
		n.CommandArgs.ParentPort,
		n.CommandArgs.ListenPort)

	err = n.NodeCommon.Go()
	if err != nil {
		log.WithFields(log.Fields{
			"context": "Go()"}).Error(err)
		return err
	}

	// // kick off Generic-sending
	// go sendGeneric(n.DialClients.SGMS, n.Entity.Id)
	log.Infof("starting %d randomWalkers", n.CommandArgs.NumTestDevices)

	for rw := 0; rw < n.CommandArgs.NumTestDevices; rw++ {
		// kick off random walking
		rwString := fmt.Sprintf("-%03d", rw)
		go randomWalker(n.DialClients.SGMS,
			n.Entity.Id+rwString,
			protobuf.Location{Latitude: 45.523, Longitude: -122.6765, Velocity: 0.1666, Course: 0.},
			protobuf.Location{Latitude: 0, Longitude: 0, Velocity: 0.001666, Course: 0.25},
			protobuf.Location{Latitude: 0, Longitude: 0, Velocity: 1.666, Course: 999999.})
		// go randomWalker(n.DialClients.SGMS,
		// 	n.Entity.Id+rwString,
		// 	protobuf.Location{Latitude: 45.523, Longitude: -122.6765, Velocity: 0.0000833, Course: 0.},
		// 	protobuf.Location{Latitude: 0, Longitude: 0, Velocity: 0.000000833, Course: 0.25},
		// 	protobuf.Location{Latitude: 0, Longitude: 0, Velocity: 0.000833, Course: 999999.})
	}

	// listen for connections
	err = listenSrv.Serve(listenConn)
	if err != nil {
		log.WithFields(log.Fields{
			"context": "Serve()"}).Error(err)
		return err
	}

	// kick off the send loop
	// se
	// stream.StartSender()
	return nil
}

func sendGeneric(client protobuf.SGMSServiceClient, entityId string) {

	r := rand.New(rand.NewSource(time.Now().UTC().UnixNano()))

	var rw float64
	rw = 0
	for {
		rw += r.NormFloat64()
		common.SendGeneric(client, entityId, time.Now().UTC(), "random", map[string]string{"EntityId": entityId}, map[string]int64{}, map[string]float64{"rw": rw})
		time.Sleep(10 * time.Millisecond)
	}
}

// randomWalker moves around randomly
func randomWalker(client protobuf.SGMSServiceClient, entityId string, startLocation, stdDeltas, max protobuf.Location) {

	r := rand.New(rand.NewSource(time.Now().UTC().UnixNano()))

	var courseDelta float64
	courseDelta = r.NormFloat64()*0.05 + 0.1
	if r.NormFloat64() > 0.0 {
		courseDelta *= -1.0
	}

	location := startLocation

	for {
		// calculate new location based on last location plus some bias based on the initial velocity and direction

		location.Course = float64(math.Mod(float64(location.Course)+(r.NormFloat64()*float64(stdDeltas.Course))+courseDelta, 360.))
		if math.Abs(float64(location.Velocity)) > float64(max.Velocity) {
			location.Velocity += float64(math.Abs(r.NormFloat64()*float64(stdDeltas.Velocity)) * -1.)
		} else {
			location.Velocity += float64(r.NormFloat64()) * stdDeltas.Velocity

		}
		location.Latitude += math.Sin(float64(location.Course)) * float64(location.Velocity)
		location.Longitude += math.Cos(float64(location.Course)) * float64(location.Velocity)
		location.Accuracy = 10.0

		common.SendLocation(client, entityId, time.Now().UTC(), location)
		time.Sleep(100 * time.Millisecond)
	}
}
