package workerize

import (
	"github.com/AlecAivazis/survey/v2"
	"github.com/davidh16/goblin/commands/database"
	central_service "github.com/davidh16/goblin/commands/service/flags/central-service"
	"github.com/davidh16/goblin/commands/workerize/flags/job"
	"github.com/davidh16/goblin/utils"
	"github.com/davidh16/goblin/utils/database_utils"
	"github.com/davidh16/goblin/utils/logger_utils"
	"github.com/davidh16/goblin/utils/workerize_utils"
	"github.com/spf13/cobra"
)

var CustomJobFlag bool

var WorkerizeCmd = &cobra.Command{
	Use:   "workerize",
	Short: "Generate workers and jobs logic",
	Run: func(cmd *cobra.Command, args []string) {
		if CustomJobFlag {
			job.GenerateCustomJob()
		} else {
			WorkerizeCmdHandler()
		}
	},
}

func WorkerizeCmdHandler() {
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

	data := workerize_utils.InitBoilerplateWorkerizeData()

	if data.JobsExists {
		confirmOverwritePrompt := &survey.Confirm{
			Message: "job.go already exists, do you wish to overwrite?",
			Default: false,
		}
		err = survey.AskOne(confirmOverwritePrompt, &data.JobsOverwrite)
		if err != nil {
			utils.HandleError(err)
		}
	}

	if data.JobsManagerExists {
		confirmOverwritePrompt := &survey.Confirm{
			Message: "jobs_manager.go already exists, do you wish to overwrite?",
			Default: false,
		}
		err = survey.AskOne(confirmOverwritePrompt, &data.JobsManagerOverwrite)
		if err != nil {
			utils.HandleError(err)
		}
	}

	if !data.CentralServiceExists {
		central_service.GenerateCentralService()
	}

	if data.WorkerPoolExists {
		confirmOverwritePrompt := &survey.Confirm{
			Message: "worker_pool.go already exists, do you wish to overwrite?",
			Default: false,
		}
		err = survey.AskOne(confirmOverwritePrompt, &data.WorkerPoolOverwrite)
		if err != nil {
			utils.HandleError(err)
		}
	}

	if data.OrchestratorExists {
		confirmOverwritePrompt := &survey.Confirm{
			Message: "orchestrator.go already exists, do you wish to overwrite?",
			Default: false,
		}
		err = survey.AskOne(confirmOverwritePrompt, &data.OrchestratorOverwrite)
		if err != nil {
			utils.HandleError(err)
		}
	}

	if !data.LoggerExists {
		implementLoggerPrompt := &survey.Confirm{
			Message: "logger is not implemented, do you wish to implement it to enrich workers and jobs logic with useful logs ?",
			Default: false,
		}
		err = survey.AskOne(implementLoggerPrompt, &data.LoggerImplemented)
		if err != nil {
			utils.HandleError(err)
		}
	} else {
		data.LoggerImplemented = true
	}

	if data.LoggerImplemented {
		err = logger_utils.GenerateLogger()
		if err != nil {
			utils.HandleError(err)
		}
	}

	err = workerize_utils.ImplementJobsLogic(data)
	if err != nil {
		utils.HandleError(err)
	}

	err = workerize_utils.ImplementWorkersLogic(data)
	if err != nil {
		utils.HandleError(err)
	}

	return
}
