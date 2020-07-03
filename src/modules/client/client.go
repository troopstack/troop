package main

import (
	"fmt"
	"os"

	"github.com/troopstack/troop/src/modules/client/cmd"
	"github.com/troopstack/troop/src/modules/client/utils"

	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use: "troop",
	RunE: func(c *cobra.Command, args []string) error {
		return c.Usage()
	},
}

func init() {
	RootCmd.AddCommand(cmd.Scout)
	RootCmd.AddCommand(cmd.ScoutUpdate)
	RootCmd.AddCommand(cmd.Ping)
	RootCmd.AddCommand(cmd.File)
	RootCmd.AddCommand(cmd.Command)
	RootCmd.AddCommand(cmd.Result)
	RootCmd.AddCommand(cmd.PluginUpdate)
	RootCmd.AddCommand(cmd.PluginVersion)
	RootCmd.AddCommand(cmd.PluginPull)
	RootCmd.AddCommand(cmd.Plugin)
}

func main() {
	utils.ParseConfig("config.ini")

	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
