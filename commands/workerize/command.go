package workerize

import (
	"github.com/spf13/cobra"
)

var WorkerizeCmd = &cobra.Command{
	Use:   "workerize",
	Short: "Generate workers and jobs logic",
	Run: func(cmd *cobra.Command, args []string) {
		workerizeCmdHandler()
	},
}

func workerizeCmdHandler() {
	//todo provjeriti ima li implementirane redis i neku od gorm bazi

}
