package config

import (
	"github.com/davidh16/goblin/cli_config"
	"github.com/davidh16/goblin/commands/config/flags/edit"
	"github.com/davidh16/goblin/utils"
	"github.com/spf13/cobra"
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
