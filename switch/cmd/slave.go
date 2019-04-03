package cmd

import (
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/gregzuro/service/switch/cmd/common"
	service "github.com/gregzuro/service/switch/cmd/slave"
)

// slaveCmd starts a slave node
var slaveCmd = &cobra.Command{
	Use:     "slave",
	Aliases: []string{"s", "sl"},
	Short:   "start a slave node",
	Run: func(cmd *cobra.Command, args []string) {
		slaveRun(cmd, args)
	},
}

func slaveRun(cmd *cobra.Command, args []string) error {

	// set up logging first thing!
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.DebugLevel)
	log.SetOutput(os.Stdout)

	log.WithFields(log.Fields{
		"context": "starting"}).Info()

	// set up a slave node
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
	common.AddCommFlags(slaveCmd, &commandArgs)
	RootCmd.AddCommand(slaveCmd)
}
