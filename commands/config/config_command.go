package config

import (
	"github.com/spf13/cobra"
	"goblin/cli_config"
	"goblin/commands/config/flags/edit"
	"goblin/utils"
)

var EditCliConfigFlag bool

var ConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Show goblin cli_config file content",
	Run: func(cmd *cobra.Command, args []string) {
		if EditCliConfigFlag {
			edit.EditConfigCmdHandler()
		} else {
			configCmdHandler()
		}
	},
}

func configCmdHandler() {
	configMap, err := cli_config.LoadConfigAsMap()
	if err != nil {
		utils.HandleError(err, "❌ Failed to load cli_config")
	}
	cli_config.PrintConfigMap(configMap)
}
