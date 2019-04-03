// Copyright © 2016 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/gregzuro/service/switch/cmd/common"
	sglog "github.com/gregzuro/service/switch/pkg/log"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/grpclog"
)

var commandArgs common.CommandArgs

var cfgFile string

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "switch",
	Short: "gregzuro Message Switch (SGMS)",
	Long: `
A message switch that can be configured in several ways:

    switch master

    switch slave

    switch device

for master, slave, or device switches, or 

    switch health -a=<other-switch-address> -p=<other-switch-port>

to get a message from some other switch.

You can also get health over HTTP 1.1 with curl:

    curl -X POST -k https://<other-switch-address>:7467/v1/health
`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the RootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	// This can be customized per-command. Setting it globally for the moment.
	log.AddHook(sglog.ContextHook{})
	// Tell grpc to use the logrus logger
	grpclog.SetLogger(log.StandardLogger())

	RootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file")
}
