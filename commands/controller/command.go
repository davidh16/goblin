package controller

import (
	"github.com/spf13/cobra"
	central_controller "goblin/commands/controller/flags/central-controller"
)

var CentralControllerFlag bool

var ControllerCmd = &cobra.Command{
	Use:   "controller",
	Short: "Generate a controller",
	Run: func(cmd *cobra.Command, args []string) {
		if CentralControllerFlag {
			central_controller.CentralControllerFlagHandler()
		}
	},
}
