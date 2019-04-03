package slave

import (
	"fmt"

	log "github.com/Sirupsen/logrus"

	"github.com/gregzuro/service/switch/cmd/common"
	"github.com/gregzuro/service/switch/protobuf"
	"golang.org/x/net/context"
)

// Node combines common state elements with type-specific ones
type Node struct {
	*common.NodeCommon
}

// New returns a new Node object
func New(commandArgs common.CommandArgs) *Node {

	node := Node{NodeCommon: common.NewNodeCommon(commandArgs, "slave")}

	// gs := ragg.NewGeoSearch()
	// data, err := pip.Asset("bindata/geodata")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// err = gs.ImportGeoData(data)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// s := &common.RAGGServer{GeoSearch: gs}
	// node.RAGGServer = *s

	return &node
}

// Go starts all comms for the node
func (n *Node) Go() error {

	var err error

	// always listen
	listenConn, listenSrv, err := n.SetUpListener()
	if err != nil {
		log.WithFields(log.Fields{
			"context": "SetUpListener()"}).Error(err)
		return err
	}

	// slaves always dial a master, expecting to stay connected to that master
	_, n.DialClients, err = common.SetUpDialer(n.CommandArgs.MasterAddress, n.CommandArgs.MasterPort)
	if err != nil {
		log.WithFields(log.Fields{
			"context": "SetUpDialer()"}).Error(err)
		return err
	}

	// contact master
	_goto, err := protobuf.InitialContactServiceClient.InitialContact(n.DialClients.Contact,
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
		})
	if err != nil {
		log.WithFields(log.Fields{
			"context": "InitialContact(1of1)"}).Error(err)
		return err
	}

	if _goto.Code != 1 { // 1= stay with me, you're cool.
		log.WithFields(log.Fields{
			"context": "InitialContact(1of1)",
			"_goto1":  _goto}).Error(err)
		return fmt.Errorf("InitialContact(1of1): %v", _goto)
	}

	// Entities that create client connections as needed.
	n.Parent = protobuf.Entity{Id: _goto.Id, Kind: _goto.Kind, Address: _goto.Address, Port: _goto.Port}

	// get GeoAffinities assigned by the master
	n.Entity.GeoAffinities = _goto.GeoAffinities
	fmt.Println("n.Entity.GeoAffinities:", n.Entity.GeoAffinities)

	// announce our identity
	log.WithFields(log.Fields{
		"context": "announce"}).Infof("I'm a '%s' switch called '%s' dialing '%s' on %d and listening on %d\n",
		n.Entity.Kind,
		n.CommandArgs.ShortName,
		n.Parent.Address,
		n.Parent.Port,
		n.CommandArgs.ListenPort)

	err = n.NodeCommon.Go()
	if err != nil {
		log.WithFields(log.Fields{
			"context": "Go()"}).Error(err)
		return err
	}

	// listen for connections
	err = listenSrv.Serve(listenConn)
	if err != nil {
		log.WithFields(log.Fields{
			"context": "Serve()"}).Error(err)
		return err
	}

	return nil
}

// func sendGeneric(client protobuf.SGMSServiceClient, entityId string) {

// 	r := rand.New(rand.NewSource(time.Now().UTC().UnixNano()))

// 	var rw float64
// 	rw = 0
// 	for {
// 		rw += r.NormFloat64()
// 		//		s := fmt.Sprintf(`"key": %f`, rw)
// 		common.SendGeneric(client, entityId, time.Now().UTC(), "random", map[string]string{"EntityId": entityId}, map[string]int64{}, map[string]float64{"rw": rw})
// 		time.Sleep(10 * time.Millisecond)
// 	}
// }
