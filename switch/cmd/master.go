package cmd

import (
	"io/ioutil"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/gregzuro/service/switch/cmd/common"
	service "github.com/gregzuro/service/switch/cmd/master"
	"github.com/gregzuro/service/switch/protobuf"
)

// masterCmd starts a master node
var masterCmd = &cobra.Command{
	Use:     "master",
	Aliases: []string{"m", "ma"},
	Short:   "start a master node",
	Run: func(cmd *cobra.Command, args []string) {
		masterRun(cmd, args)
	},
}

func masterRun(cmd *cobra.Command, args []string) error {

	// set up logging first thing!
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.InfoLevel)
	log.SetOutput(os.Stdout)

	log.WithFields(log.Fields{
		"context": "starting"}).Info()

	// set up a master node
	if err := common.ReadConfigFile(cfgFile, cmd, &commandArgs); err == nil {
		log.WithFields(log.Fields{
			"context": "readConfigFile"}).Info("Using config file:", viper.ConfigFileUsed())
	} else {
		log.WithFields(log.Fields{
			"context": "readConfigFile"}).Error(err)
		return nil
	}

	node := service.New(commandArgs)

	if err := handleGeoAffinities(node); err != nil {
		if err != nil {
			log.WithFields(log.Fields{
				"context": "handleGeoAffinities()",
			}).Error(err)
		}
	}

	// master is the only switch type that has GAs, so do this here (master.go)

	if err := node.Go(); err != nil {
		log.WithFields(log.Fields{
			"context": "Go"}).Error(err)
	}

	return nil
}

func handleGeoAffinities(node *service.Node) error {
	wantCoveringCL, err := commandArgs.GeoAffinityCL.ParseCLGeoAffinities(&node.Entity.GeoAffinities)
	if err != nil {
		log.WithFields(log.Fields{"context": "ParseCLGeoAffinities()"}).Error(err)
	}

	// read static slave geoaffinities from the file specified in config
	commandArgs.SlaveGeoAffinitiesFile = viper.Get("slave-geoaffinities-file").(string)
	if commandArgs.SlaveGeoAffinitiesFile != "" {
		staticGeoaffinities, err := ioutil.ReadFile(commandArgs.SlaveGeoAffinitiesFile)
		if err != nil {
			log.WithFields(log.Fields{
				"context": "ReadFile",
			}).Error(err)
			return err
		}

		wantCoveringConfig, err := common.ParseJSONGeoAffinities(&node.Entity.GeoAffinities, staticGeoaffinities)
		if err != nil {
			log.WithFields(log.Fields{
				"context": "ParseJsonGeoAffinities()",
			}).Error(err)
			return err
		}

		if node.Need.NeedStatus == nil {
			node.Need.NeedStatus = make(map[string]*protobuf.NeedStatus)
		}
		node.Need.NeedStatus["slave"] = &protobuf.NeedStatus{
			Need: int32(wantCoveringCL + wantCoveringConfig),
			Have: 0,
		}

		go node.UpdateSlaveNeeds()

		// var g protobuf.GeoJson_FeatureCollection
		// err = jsonpb.UnmarshalString(string(staticGeoaffinities), &g)
		// if err != nil {
		// 	log.WithFields(log.Fields{
		// 		"context": "jsonpb.UnmarshalString",
		// 	}).Error(err)
		// }

		// fmt.Println("***************** staticGeoaffinities.DataType:", g)
	}
	return nil
}

func init() {
	common.AddCommFlags(masterCmd, &commandArgs)
	common.AddGeoAffinityFlags(masterCmd, &commandArgs)
	RootCmd.AddCommand(masterCmd)
}
