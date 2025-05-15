package workerize

import (
	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"goblin/commands/database"
	"goblin/utils"
	"goblin/utils/database_utils"
	"goblin/utils/workerize_utils"
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
	implementedDatabases, err := workerize_utils.ListImplementedDatabases()
	if err != nil {
		utils.HandleError(err, "Unable to list implemented databases")
	}

	if len(implementedDatabases) > 1 {
		var redisImplemented bool
		for _, impl := range implementedDatabases {
			if impl == database_utils.DatabaseOptionNamesMap[database_utils.Redis] {
				redisImplemented = true
				break
			}
		}

		if !redisImplemented {
			var confirmContinue bool
			confirmContinuePrompt := &survey.Confirm{
				Message: "For implementing background jobs and workers, Redis needs to be implemented, do you wish to continue with Redis implementation?",
				Default: false,
			}
			err = survey.AskOne(confirmContinuePrompt, &confirmContinue)
			if err != nil {
				utils.HandleError(err)
			}
			if !confirmContinue {
				return
			}

			err = database.ImplementRedis()
			if err != nil {
				utils.HandleError(err)
			}
		}

	} else {

		var confirmContinue bool
		confirmContinuePrompt := &survey.Confirm{
			Message: "For implementing background jobs and workers, one persistent database and Redis need to be implemented, do you wish to continue with database implementations?",
			Default: false,
		}
		err = survey.AskOne(confirmContinuePrompt, &confirmContinue)
		if err != nil {
			utils.HandleError(err)
		}
		if !confirmContinue {
			return
		}
		err = database.ImplementRedisAndOtherGormDb()
		if err != nil {
			utils.HandleError(err)
		}
	}

}
