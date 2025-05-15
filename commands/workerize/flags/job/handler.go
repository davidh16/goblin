package job

import (
	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"goblin/utils"
	"goblin/utils/workerize_utils"
)

var JobCmd = &cobra.Command{
	Use:   "job",
	Short: "Generate a custom job",
	Run: func(cmd *cobra.Command, args []string) {
		GenerateCustomJob()
	},
}

func GenerateCustomJob() {
	workerizeInitialized := workerize_utils.IfWorkerizeIsInitialized()
	if !workerizeInitialized {

		var confirmContinue bool
		confirmContinuePrompt := &survey.Confirm{
			Message: "There are missing workerize files, to implement a custom job, workerize command needs to be initialized first, do you wish to continue?",
			Default: false,
		}
		err := survey.AskOne(confirmContinuePrompt, &confirmContinue)
		if err != nil {
			utils.HandleError(err)
		}
		if !confirmContinue {
			return
		}

		workerize_utils.WorkerizeCmdHandlerCopy()

	}

}
