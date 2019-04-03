package common

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	google_protobuf "github.com/golang/protobuf/ptypes/empty"
	"github.com/gregzuro/service/pkg/message"
	"github.com/gregzuro/service/pkg/remote"
	"github.com/gregzuro/service/switch/db"
	"github.com/gregzuro/service/switch/protobuf"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/peterh/liner"
	"golang.org/x/net/context"
	"golang.org/x/net/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
)

// NodeCommon contains all the state items that are common to all node types
type NodeCommon struct {
	// ReceiveMessage is how messages get into the switch.
	message.Receiver `json:"-"`

	// Tee returns a new (synchronous) copy of the node's message queue.
	Tee func(context.Context) <-chan message.Message `json:"-"`

	Entity            protobuf.Entity
	CommandArgs       CommandArgs
	DialClients       *RPCClients `json:"-"`
	Master            protobuf.Entity
	Parent            protobuf.Entity
	Masters           map[string]protobuf.Entity
	MastersLock       sync.RWMutex `json:"-"`
	Slaves            map[string]protobuf.Entity
	SlavesLock        sync.RWMutex `json:"-"`
	PendingSlaves     map[string]protobuf.Entity
	PendingSlavesLock sync.RWMutex `json:"-"`
	Devices           map[string]protobuf.Entity
	DevicesLock       sync.RWMutex `json:"-"`
	Plugins           map[string]protobuf.Entity
	PluginsLock       sync.RWMutex `json:"-"`

	TSDB                         db.TSDB `json:"-"`
	InboundMessageBatchCount     uint64
	InboundMessageCount          uint64
	StartTime                    time.Time
	LastInboundMessageBatchCount uint64
	LastInboundMessageCount      uint64
	LastStatsPeriodInSeconds     float64
	LastStatsTime                time.Time
	MessagesPerSecond            float64
	MessageBatchesPerSecond      float64

	CommandLine *liner.State

	Need     protobuf.NeedsResponse
	NeedLock sync.RWMutex `json:"-"`

	RAGGServer RAGGServer
}

// NewNodeCommon returns a new NodeCommon.
func NewNodeCommon(commandArgs CommandArgs, kind string) *NodeCommon {
	var err error
	node := NodeCommon{
		CommandArgs: commandArgs,
		Entity: protobuf.Entity{
			Kind: kind,
			Port: commandArgs.ListenPort,
		},
		StartTime: time.Now().UTC(),
	}
	// get a unique id for this node
	node.Entity.Id, err = commandArgs.GetEntityID()
	if err != nil {
		log.WithFields(log.Fields{"context": "GetEntityId"}).Error(err)
	}

	// initialize maps
	node.Masters = make(map[string]protobuf.Entity)
	node.Slaves = make(map[string]protobuf.Entity)
	node.Devices = make(map[string]protobuf.Entity)
	node.Plugins = make(map[string]protobuf.Entity)

	// set up message processing queues
	in := make(chan message.Message)
	node.Tee = message.Tee(in)
	node.Receiver = message.FuncReceiver(func(m message.Message) error {
		in <- m
		// TODO: Right now this channel isn't ever closed.
		// When it is, make sure this func returns an error
		// rather than panicking.
		return nil
	})

	// Catch-all handler.
	go func() {
		for m := range node.Tee(context.TODO()) {
			node.processIncomingSGMS(context.TODO(), m)
		}
	}()

	return &node
}

// Go starts outbound connections.
func (n *NodeCommon) Go() error {
	ctx, cancel := context.WithCancel(context.Background())

	// Open a stream connection to parent.

	// receive messages from parent
	send, in := remote.ClientStreamer(ctx, n.DialClients.SGMS)
	go message.ReceiveFromChan(in, n, cancel)

	// send messages to parent
	out := n.Tee(ctx)
	go message.ReceiveFromChan(out, send, cancel)

	// send heartbeat to parent
	heart := message.HeartBeat(ctx, n.Entity.Id)
	go message.ReceiveFromChan(heart, send, cancel)

	return nil

}

// SetUpListener sets up a Listener and Server with common SGMS handlers.
// This default implementation may be overridden by different node types to supply an
// alternate mux.
func (n *NodeCommon) SetUpListener() (net.Listener, *http.Server, error) {
	// This is funky because grpc didn't expose ServeHTTP
	// https://github.com/grpc/grpc-go/issues/75
	// Now it does, but there are performance issues:
	// https://github.com/grpc/grpc-go/issues/586

	//using the default mux because that's where the x/net/trace handlers are registered
	mux := http.DefaultServeMux
	gwmux := runtime.NewServeMux()

	serverName := fmt.Sprintf("%s:%d",
		"localhost",
		n.CommandArgs.ListenPort)

	fmt.Printf("SetUpListener: serverName (of self); %v\n", serverName)

	CertPool := GetCertPool()
	opts := []grpc.ServerOption{grpc.Creds(credentials.NewClientTLSFromCert(CertPool, "dev"))}

	grpcServer := grpc.NewServer(opts...)

	ctx := context.Background()

	dcreds := credentials.NewTLS(&tls.Config{
		ServerName: "dev", //TODO(greg) change this to come from env?
		RootCAs:    CertPool,
	})
	dopts := []grpc.DialOption{grpc.WithTransportCredentials(dcreds)}

	// allow trace connections from anywhere
	trace.AuthRequest = func(req *http.Request) (any, sensitive bool) {
		return true, true
	}

	mux.HandleFunc("/swagger.json", func(w http.ResponseWriter, req *http.Request) {
		io.Copy(w, strings.NewReader(protobuf.Swagger))
	})

	mux.Handle("/ui/", http.StripPrefix("/ui/", http.FileServer(http.Dir("ui"))))
	mux.Handle("/swagger-ui/", http.StripPrefix("/swagger-ui/", http.FileServer(http.Dir("third_party/swagger-ui"))))

	var srv *http.Server
	var conn net.Listener

	// health
	protobuf.RegisterStatusServiceServer(grpcServer, n)
	err := protobuf.RegisterStatusServiceHandlerFromEndpoint(ctx, gwmux, serverName, dopts)
	if err != nil {
		fmt.Printf("StatusService: %v\n", err)
		return conn, srv, err
	}

	// messages
	protobuf.RegisterSGMSServiceServer(grpcServer, n)
	err = protobuf.RegisterSGMSServiceHandlerFromEndpoint(ctx, gwmux, serverName, dopts)
	if err != nil {
		fmt.Printf("SGMSMessageService: %v\n", err)
		return conn, srv, err
	}

	// contact
	protobuf.RegisterInitialContactServiceServer(grpcServer, n)
	err = protobuf.RegisterInitialContactServiceHandlerFromEndpoint(ctx, gwmux, serverName, dopts)
	if err != nil {
		fmt.Printf("InitialContactService: %v\n", err)
		return conn, srv, err
	}

	mux.Handle("/", AllowCORS(gwmux))

	conn, err = net.Listen("tcp", fmt.Sprintf(":%d", n.CommandArgs.ListenPort))
	if err != nil {
		return conn, srv, err
	}

	KeyPair := GetKeyPair()
	srv = &http.Server{
		Addr:    serverName,
		Handler: GrpcHandlerFunc(grpcServer, mux),
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{*KeyPair},
			NextProtos:   []string{"h2"},
		},
	}

	conn = tls.NewListener(conn, srv.TLSConfig)
	return conn, srv, nil
}

func (n *NodeCommon) logSGMSMetrics(c context.Context, s *protobuf.SGMS) {
	n.InboundMessageBatchCount++

	// print a status message
	if time.Since(n.LastStatsTime).Seconds() < .45 {
		return
	}

	n.LastStatsPeriodInSeconds = time.Now().UTC().Sub(n.LastStatsTime).Seconds()
	n.MessagesPerSecond = float64(n.InboundMessageCount-n.LastInboundMessageCount) / n.LastStatsPeriodInSeconds
	n.MessageBatchesPerSecond = float64(n.InboundMessageBatchCount-n.LastInboundMessageBatchCount) / n.LastStatsPeriodInSeconds
	n.LastStatsTime = time.Now().UTC()
	n.LastInboundMessageCount = n.InboundMessageCount
	n.LastInboundMessageBatchCount = n.InboundMessageBatchCount

	// log.WithFields(log.Fields{
	// 	"context":                 "update==true",
	// 	"MessagesPerSecond":       n.MessagesPerSecond,
	// 	"MessageBatchesPerSecond": n.MessageBatchesPerSecond}).Info()
}

func (n *NodeCommon) updateLastSeen(h *protobuf.HeartBeat) {

	switch n.Entity.Kind {

	case "master":
		// if it's *my* child, then update the last seen time
		if tmp, ok := n.Masters[h.Common.EntityId]; ok {
			tmp.LastSeen = ToPbTime(time.Now().UTC())
			n.MastersLock.Lock()
			n.Masters[h.Common.EntityId] = tmp
			n.MastersLock.Unlock()
			return
		}

		// if it's *my* child, then update the last seen time
		if tmp, ok := n.Slaves[h.Common.EntityId]; ok {
			tmp.LastSeen = ToPbTime(time.Now().UTC())
			n.SlavesLock.Lock()
			n.Slaves[h.Common.EntityId] = tmp
			n.SlavesLock.Unlock()
			return
		}

		// if it's *my* child, then update the last seen time
		if tmp, ok := n.Plugins[h.Common.EntityId]; ok {
			tmp.LastSeen = ToPbTime(time.Now().UTC())
			n.PluginsLock.Lock()
			n.Plugins[h.Common.EntityId] = tmp
			n.PluginsLock.Unlock()
			return
		}

	case "slave":
		// if it's *my* child, then update the last seen time
		if tmp, ok := n.Devices[h.Common.EntityId]; ok {
			tmp.LastSeen = ToPbTime(time.Now().UTC())
			n.DevicesLock.Lock()
			n.Devices[h.Common.EntityId] = tmp
			n.DevicesLock.Unlock()
			return
		}

		// if it's *my* child, then update the last seen time
		if tmp, ok := n.Slaves[h.Common.EntityId]; ok {
			tmp.LastSeen = ToPbTime(time.Now().UTC())
			n.SlavesLock.Lock()
			n.Slaves[h.Common.EntityId] = tmp
			n.SlavesLock.Unlock()
			return
		}

		// if it's *my* child, then update the last seen time
		if tmp, ok := n.Plugins[h.Common.EntityId]; ok {
			tmp.LastSeen = ToPbTime(time.Now().UTC())
			n.PluginsLock.Lock()
			n.Plugins[h.Common.EntityId] = tmp
			n.PluginsLock.Unlock()
			return
		}

	case "device":
		// if it's *my* child, then update the last seen time
		if tmp, ok := n.Plugins[h.Common.EntityId]; ok {
			tmp.LastSeen = ToPbTime(time.Now().UTC())
			n.PluginsLock.Lock()
			n.Plugins[h.Common.EntityId] = tmp
			n.PluginsLock.Unlock()
			return
		}
	}
}

func (n *NodeCommon) processIncomingSGMS(c context.Context, m message.Message) {
	s := m.SGMS()

	n.logSGMSMetrics(c, s)

	if s.HeartBeat != nil {
		n.InboundMessageCount++
		n.updateLastSeen(s.HeartBeat)
	}

	if s.Generic != nil {
		n.InboundMessageCount += uint64(len(s.Generic))

		if n.TSDB != nil {
			n.TSDB.SaveGeneric(s.Generic)
		}
	}

	if s.Location != nil {
		n.InboundMessageCount += uint64(len(s.Location))

		// if n.MessageBatchesPerSecond < 10 {
		// 	if n.Entity.Kind == "slave" {
		// 		for _, l := range s.Location {
		// 			log.WithFields(log.Fields{
		// 				"context":   "Kind == `slave`",
		// 				"location":  n.RAGGServer.countryHandler(*l),
		// 				"latitude":  s.Location[0].Latitude,
		// 				"longitude": s.Location[0].Longitude,
		// 			}).Info()
		// 		}
		// 	} else {
		// 		log.WithFields(log.Fields{
		// 			"context":   "Kind != `slave`",
		// 			"latitude":  s.Location[0].Latitude,
		// 			"longitude": s.Location[0].Longitude,
		// 		}).Info()
		// 	}
		// }

		if n.TSDB != nil {
			n.TSDB.SaveLocation(s.Location)
		}
	}
}

// SendSGMS is actually the handler function for receiving a single message.
func (n *NodeCommon) SendSGMS(c context.Context, s *protobuf.SGMS) (*google_protobuf.Empty, error) {
	n.ReceiveMessage(message.Wrap(s))
	return &google_protobuf.Empty{}, nil
}

// StreamSGMS handles streaming connections
func (n *NodeCommon) StreamSGMS(stream protobuf.SGMSService_StreamSGMSServer) error {
	// use stream pointer to uniquely identify this source
	source := &stream
	go func() {
		in := n.Tee(stream.Context())
		for m := range in {
			if m.Source() == source {
				// log.Info("Skipping message to avoid reflection")
				continue
			}
			if err := stream.Send(m.SGMS()); err != nil {
				log.Error(err)
				return
			}
		}
	}()
	for {
		s, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		m := message.Wrap(s)
		m.SetSource(source)
		n.ReceiveMessage(m)
	}
}

// Health returns the Health object to the caller // just the Node object for now
func (n *NodeCommon) Health(c context.Context, s *protobuf.HealthRequest) (*protobuf.HealthResponse, error) {
	json, err := json.MarshalIndent(&n, "", "   ")
	if err != nil {
		log.WithFields(log.Fields{
			"context": "json.MarshalIndent"}).Error(err)
	}

	response := protobuf.HealthResponse{Value: "\n" + string(json) + "\n"} // TODO(greg): move NodeCommon to protobuf so we can stop wrapping this in a string
	return &response, nil
}

// Needs returns the Need object to the caller
func (n *NodeCommon) Needs(c context.Context, s *protobuf.NeedsRequest) (*protobuf.NeedsResponse, error) {
	return &n.Need, nil
}

// InitialContact determines to which slave the caller should connect, then returns that info to the caller
func (n *NodeCommon) InitialContact(c context.Context, h *protobuf.Hello) (*protobuf.Goto, error) {

	var _goto protobuf.Goto
	var msg string

	s := h.CallerEntity // caller
	pr, ok := peer.FromContext(c)
	if !ok {
		return &_goto, errors.New("failed to get peer from ctx")
	}
	if pr.Addr == net.Addr(nil) {
		return &_goto, errors.New("failed to get peer address")
	}

	// receiver's kind
	switch n.Entity.Kind {
	case "master":

		// caller's kind
		switch s.Kind {

		case "master":
			msg = fmt.Sprintf("I'm your master.")
			_goto = protobuf.Goto{Code: 1, Id: n.Entity.Id, Kind: n.Entity.Kind, Address: h.CalledEntity.Address, Port: h.CalledEntity.Port}
			// TODO(greg) add master to this node's children
			n.MastersLock.Lock()
			n.Masters[s.Id] = protobuf.Entity{Id: s.Id, Kind: s.Kind, Address: s.Address, Port: s.Port, LastSeen: ToPbTime(time.Now().UTC())}
			n.MastersLock.Unlock()

		case "slave":
			msg = fmt.Sprintf("I'm your master.")
			gAk, err := determineGeoAffinityForSlave(n.Entity.GeoAffinities, n.Slaves)
			if err != nil {
				msg = msg + " " + err.Error()
				_goto = protobuf.Goto{Code: 4}
			} else {
				gAs := []*protobuf.GeoAffinity{n.Entity.GeoAffinities[gAk]}
				c := protobuf.Coverer{EntityId: s.Id, CovererStartedTime: ToPbTime(time.Now().UTC()), CovererRunning: false}
				n.Entity.GeoAffinities[gAk].Coverers = append(n.Entity.GeoAffinities[gAk].Coverers, &c)
				_goto = protobuf.Goto{GeoAffinities: gAs, Code: 1, Id: n.Entity.Id, Kind: n.Entity.Kind, Address: h.CalledEntity.Address, Port: h.CalledEntity.Port}
			}
			n.SlavesLock.Lock()
			n.Slaves[s.Id] = protobuf.Entity{Id: s.Id, Kind: s.Kind, Address: s.Address, Port: s.Port, LastSeen: ToPbTime(time.Now().UTC())}
			n.SlavesLock.Unlock()

		case "device":
			// find a good slave
			fmt.Println(pr.Addr.String())
			gS, err := findGoodSlave(n.Slaves)
			if err != nil {
				msg = fmt.Sprintf("I don't have any slaves for you.")
				_goto = protobuf.Goto{Code: 4}
			} else {
				msg = fmt.Sprintf("Contact the given slave.")
				_goto = protobuf.Goto{Code: 2, Id: gS.Id, Kind: "slave", Address: gS.Address, Port: gS.Port}
			}

		case "plugin":
			msg = fmt.Sprintf("Welcome.")
			_goto = protobuf.Goto{Code: 1, Id: n.Entity.Id, Kind: n.Entity.Kind, Address: h.CalledEntity.Address, Port: h.CalledEntity.Port}
			n.PluginsLock.Lock()
			n.Plugins[s.Id] = protobuf.Entity{Id: s.Id, Kind: s.Kind, Address: s.Address, Port: s.Port, LastSeen: ToPbTime(time.Now().UTC())}
			n.PluginsLock.Unlock()

		case "health":
			msg = fmt.Sprintf("I don't have anything for you.")
			_goto = protobuf.Goto{Code: 3}

		}

	case "slave":

		// caller's kind
		switch s.Kind {

		case "master":
			msg = fmt.Sprintf("Can't do that.")
			_goto = protobuf.Goto{Code: 4}

		case "slave":
			msg = fmt.Sprintf("Can't do that.")
			_goto = protobuf.Goto{Code: 4}

		case "device":
			msg = fmt.Sprintf("Welcome.")
			_goto = protobuf.Goto{Code: 1, Id: n.Entity.Id, Kind: n.Entity.Kind, Address: h.CalledEntity.Address, Port: h.CalledEntity.Port}
			n.DevicesLock.Lock()
			n.Devices[s.Id] = protobuf.Entity{Id: s.Id, Kind: s.Kind, Address: s.Address, Port: s.Port, LastSeen: ToPbTime(time.Now().UTC())}
			n.DevicesLock.Unlock()

		case "plugin":
			msg = fmt.Sprintf("Welcome.")
			_goto = protobuf.Goto{Code: 1, Id: n.Entity.Id, Kind: n.Entity.Kind, Address: h.CalledEntity.Address, Port: h.CalledEntity.Port}
			n.PluginsLock.Lock()
			n.Plugins[s.Id] = protobuf.Entity{Id: s.Id, Kind: s.Kind, Address: s.Address, Port: s.Port, LastSeen: ToPbTime(time.Now().UTC())}
			n.PluginsLock.Unlock()

		case "health":
			msg = fmt.Sprintf("I don't have anything for you.")
			_goto = protobuf.Goto{Code: 3}

		}

	case "device":

		// caller's kind
		switch s.Kind {

		case "master":
			msg = fmt.Sprintf("Can't do that.")
			_goto = protobuf.Goto{Code: 4}

		case "slave":
			msg = fmt.Sprintf("Can't do that.")
			_goto = protobuf.Goto{Code: 4}

		case "device":
			msg = fmt.Sprintf("Can't do that.")
			_goto = protobuf.Goto{Code: 4}

		case "plugin":
			msg = fmt.Sprintf("Welcome.")
			_goto = protobuf.Goto{Code: 1, Id: n.Entity.Id, Kind: n.Entity.Kind, Address: h.CalledEntity.Address, Port: h.CalledEntity.Port}
			// TODO(greg) add plugin to this node's children
			n.PluginsLock.Lock()
			n.Plugins[s.Id] = protobuf.Entity{Id: s.Id, Kind: s.Kind, Address: s.Address, Port: s.Port, LastSeen: ToPbTime(time.Now().UTC())}
			n.PluginsLock.Unlock()

		case "health":
			msg = fmt.Sprintf("I don't have anything for you.")
			_goto = protobuf.Goto{Code: 3}

		}

	}

	_goto.Message = msg + fmt.Sprintf(" You: %s (%s). Me: %s (%s). My Master: %s (%s). My Parent: %s (%s).",
		s.Id, s.Kind, n.Entity.Id, n.Entity.Kind, n.Master.Id, n.Master.Kind, n.Parent.Id, n.Parent.Kind)

	return &_goto, nil
}

// CountSlaves returns the number of good an dbad slaves
func (n *NodeCommon) CountSlaves() (int, int) {

	good := 0
	bad := 0

	for _, v := range n.Slaves {
		if time.Since(ToGoTime(v.LastSeen)) < time.Second*10 {
			good++
		} else {
			bad++
		}

	}

	return good, bad
}

func (n *NodeCommon) UpdateSlaveNeeds() (int, error) {

	for {
		time.Sleep(10 * time.Second)
		good, _ := n.CountSlaves()

		n.Need.NeedStatus["slave"].Have = int32(good)
	}

}
