package cmd

// TODO

import (
	"os"

	log "github.com/Sirupsen/logrus"

	"github.com/gregzuro/service/switch/cmd/common"
	service "github.com/gregzuro/service/switch/cmd/device"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// deviceCmd starts a device node
var deviceCmd = &cobra.Command{
	Use:     "device",
	Aliases: []string{"d", "de"},
	Short:   "start a device node",
	Run: func(cmd *cobra.Command, args []string) {
		deviceRun(cmd, args)
	},
}

func deviceRun(cmd *cobra.Command, args []string) error {

	// set up logging first thing!
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.DebugLevel)
	log.SetOutput(os.Stdout)

	log.WithFields(log.Fields{
		"context": "starting"}).Info()

	// set up a device node
	if err := common.ReadConfigFile(cfgFile, cmd, &commandArgs); err == nil {
		log.WithFields(log.Fields{
			"context": "readConfigFile"}).Info("Using config file:", viper.ConfigFileUsed())
	} else {
		log.WithFields(log.Fields{
			"context": "readConfigFile"}).Error(err)
		return nil
	}

	node := service.New(commandArgs)
	err := node.Go()
	if err != nil {
		// logxx.Prixxntln(err.Error())
		log.WithFields(log.Fields{
			"context": "Go"}).Error(err)
	}

	return nil
}

func init() {
	common.AddCommFlags(deviceCmd, &commandArgs)

	deviceCmd.Flags().IntVarP(&commandArgs.NumTestDevices, "num-test-devices", "n", 0, "the number of test devices to create")
	deviceCmd.Flags().BoolVarP(&commandArgs.PortMappedCluster, "port-mapped", "m", false, "Cluster is running port-mapped on a single host (Docker). Reuse the master address for subsequent cluster connections.")

	RootCmd.AddCommand(deviceCmd)
}
