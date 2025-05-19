package logger

import (
	"github.com/spf13/cobra"
)

var LoggerCmd = &cobra.Command{
	Use:   "logger",
	Short: "Generate logger logic",
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func loggerCmdHandler() {

}
