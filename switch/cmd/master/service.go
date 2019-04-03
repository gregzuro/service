package master

import (
	"fmt"

	log "github.com/Sirupsen/logrus"

	"github.com/gregzuro/service/switch/cmd/common"
	"github.com/gregzuro/service/switch/db"
	"github.com/gregzuro/service/switch/protobuf"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
)

// Node combines common state elements with type-specific ones
type Node struct {
	// FirstSlaveListenPort int64
	SlaveCount int64
	*common.NodeCommon
}

// New returns a new Node object
func New(commandArgs common.CommandArgs) *Node {

	node := Node{
		NodeCommon: common.NewNodeCommon(commandArgs, "master"),
		// FirstSlaveListenPort: 7500, // TODO(greg) make this an env parameter /TODO this is the *base* port that the new slave should try.  incrementing if failing to get the port
	}

	return &node
}

// Go starts all comms for the node
func (n *Node) Go() error {

	var err error

	// connect to time-series database
	n.TSDB, err = db.NewInfluxTSDB(
		viper.Get("influxdb-host").(string),
		viper.Get("influxdb-dbname").(string),
		viper.Get("influxdb-user").(string),
		viper.Get("influxdb-password").(string))

	if err != nil {
		fmt.Println(err.Error())
	}

	// always listen
	listenConn, listenSrv, err := n.SetUpListener()
	if err != nil {
		log.WithFields(log.Fields{
			"context": "SetUpListener()"}).Error(err)
		return err
	}

	// masters sometimes dial a parent, but not ATM
	if n.CommandArgs.MasterAddress != "" {
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
		//		n.Parent = common.Entity{Entity: protobuf.Entity{Id: _goto.Id, Kind: _goto.Kind, Address: _goto.Address, Port: _goto.Port}}
		n.Parent = protobuf.Entity{Id: _goto.Id, Kind: _goto.Kind, Address: _goto.Address, Port: _goto.Port}
	} else {
		n.Parent = protobuf.Entity{}
	}

	// announce our identity

	announce := fmt.Sprintf("I'm a '%s' switch called '%s'",
		n.Entity.Kind,
		n.CommandArgs.ShortName)

	if n.CommandArgs.ParentAddress != "" {
		announce += fmt.Sprintf("dialing '%s' on %d and ", n.Parent.Address, n.Parent.Port)

		err = n.NodeCommon.Go()
		if err != nil {
			log.WithFields(log.Fields{
				"context": "Go()"}).Error(err)
			return err
		}
	}

	log.WithFields(log.Fields{
		"context": "announce"}).Infof("%s listening on %d\n",
		announce,
		n.CommandArgs.ListenPort)

	// // start commandLine handler
	// go commandLine(n)

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

// func commandLine(n *Node) {
// 	if !terminal.IsTerminal(int(os.Stdout.Fd())) {
// 		log.WithFields(log.Fields{
// 			"context": "GetStdinMode"}).Info("Error with Stdin, so disabling console.")

// 		return
// 	}

// 	n.CommandLine = liner.NewLiner()
// 	defer n.CommandLine.Close()

// 	//	n.CommandLine.SetCtrlCAborts(true)

// 	for {
// 		if cmd, err := n.CommandLine.Prompt("amp> "); err == nil {

// 			cl := strings.Split(cmd, " ")

// 			switch cl[0] {

// 			case "quit", "exit", "q":
// 				fmt.Println("use ctrl-c to quit process")
// 				goto end

// 			case "start":
// 				if len(cl) == 1 {
// 					fmt.Println("What do you want to start?")
// 					break
// 				}

// 				switch cl[1] {
// 				case "slave":
// 					fmt.Println("Yay! Let's start a", cl[1])

// 					startSlave(n)

// 				default:
// 					fmt.Println("I don't know", cl[1])

// 				}

// 			case "get":

// 				switch cl[1] {
// 				case "dockeraddress":
// 					fmt.Println(viper.GetString("DockerAddress"))

// 				case "containers":
// 					fmt.Println(common.GetContainerList(viper.GetString("DockerAddress")))

// 				default:
// 					fmt.Println("I don't know", cl[1])

// 				}

// 			default:
// 			}

// 		} else if err == liner.ErrPromptAborted {
// 			fmt.Println("Switch Console Aborted")
// 			break
// 		} else {
// 			fmt.Println("Error reading line: ", err)
// 		}
// 	}

// end:
// }

// func startSlave(n *Node) {
// 	docker := viper.GetString("DockerAddress")
// 	fmt.Println("Connecting to docker:", docker)
// 	defaultHeaders := map[string]string{"User-Agent": "engine-api-cli-1.0"}
// 	cli, err := client.NewClient(docker, "v1.22", nil, defaultHeaders)
// 	if err != nil {
// 		panic(err)
// 	}

// 	listenPort := n.FirstSlaveListenPort + n.SlaveCount
// 	n.SlaveCount++
// 	containerName := fmt.Sprintf("sgms_slave%03d", n.SlaveCount)
// 	containerConfigCmdTemplate := "./slave run -a %s -p %d -l %d --short-name %s --stream"
// 	cmd := fmt.Sprintf(
// 		containerConfigCmdTemplate,
// 		n.CommandArgs.ShortName,
// 		n.Entity.Port,
// 		listenPort,
// 		containerName,
// 	)
// 	exposedPorts := make(map[nat.Port]struct{})
// 	stringListenPort := strconv.FormatInt(listenPort, 10)
// 	exposedPorts[nat.Port(stringListenPort)] = struct{}{}
// 	containerConfig := &container.Config{
// 		Image:        "gregzuro/switch/slave",
// 		Cmd:          strslice.StrSlice(strings.Split(cmd, " ")),
// 		Tty:          true,
// 		ExposedPorts: exposedPorts,
// 	}
// 	portBindings := make(map[nat.Port][]nat.PortBinding)
// 	portBindings[nat.Port(stringListenPort)] = []nat.PortBinding{
// 		nat.PortBinding{HostIP: "0.0.0.0", HostPort: stringListenPort},
// 	}
// 	cont, err := cli.ContainerCreate(context.Background(),
// 		containerConfig,
// 		&container.HostConfig{
// 			NetworkMode: "sgms",
// 			//PublishAllPorts: true,
// 			PortBindings: portBindings,
// 			LogConfig: container.LogConfig{
// 				Type: "fluentd",
// 				Config: map[string]string{
// 					"tag": containerName,
// 				},
// 			},
// 		},
// 		nil, //networkingConfig,
// 		containerName,
// 	)
// 	fmt.Println(cont, err)
// 	err = cli.ContainerStart(context.Background(), cont.ID, types.ContainerStartOptions{})
// 	fmt.Println(err)

// 	fmt.Println(cmd)

// }
