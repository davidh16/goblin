package job

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"goblin/cli_config"
	"goblin/utils"
	"goblin/utils/workerize_utils"
	"path"
)

var JobCmd = &cobra.Command{
	Use:   "job",
	Short: "Generate a custom job",
	Run: func(cmd *cobra.Command, args []string) {
		GenerateCustomJob()
	},
}

func GenerateCustomJob() {

	customJobData := &workerize_utils.CustomJobData{}

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

	for {
		if err := survey.AskOne(&survey.Input{
			Message: "Please type the job file name (snake_case), keep in mind that it will get a suffix _job.go automatically:",
			Default: "my_custom_job",
		}, &customJobData.JobNameSnakeCase); err != nil {
			utils.HandleError(err)
		}

		if !utils.IsSnakeCase(customJobData.JobNameSnakeCase) {
			fmt.Printf("🛑 %s is not in snake case\n", customJobData.JobNameSnakeCase)
			continue
		}

		customJobData.JobFileName = customJobData.JobNameSnakeCase + "_job.go"
		customJobData.JobNameCamelCase = utils.SnakeToCamel(customJobData.JobNameSnakeCase)
		customJobData.JobNamePascalCase = utils.SnakeToPascal(customJobData.JobNameSnakeCase)
		customJobData.JobTypeName = "JobType" + customJobData.JobNamePascalCase
		customJobData.JobFilePath = path.Join(cli_config.CliConfig.JobsFolderPath, customJobData.JobFileName)
		customJobData.JobMetadataName = customJobData.JobNamePascalCase + "JobMetadata"
		customJobData.JobMetadataFileName = customJobData.JobNameSnakeCase + "_job_metadata.go"

		var confirmContinue bool
		confirmPrompt := &survey.Confirm{
			Message: fmt.Sprintf("You are about to create a custom job file named %s, do you want to continue ?", customJobData.JobFileName),
		}
		if err := survey.AskOne(confirmPrompt, &confirmContinue); err != nil {
			utils.HandleError(err)
		}

		if !confirmContinue {
			continue
		}

		if utils.FileExists(customJobData.JobFilePath) {
			var overwriteConfirmed bool
			confirmPrompt = &survey.Confirm{
				Message: fmt.Sprintf("%s already exists. Do you want to overwrite it ?", customJobData.JobFileName),
				Default: false,
			}
			if err := survey.AskOne(confirmPrompt, &overwriteConfirmed); err != nil {
				utils.HandleError(err)
			}

			if overwriteConfirmed {
				confirmPrompt = &survey.Confirm{
					Message: fmt.Sprintf("Are you sure you want to overwrite %s ?", customJobData.JobFileName),
					Default: false,
				}
				if err := survey.AskOne(confirmPrompt, &overwriteConfirmed); err != nil {
					utils.HandleError(err)
				}
				customJobData.AlreadyExists = true
			}

			if !overwriteConfirmed {
				continue
			}
		}

		customJobData.WorkerPoolFileName = customJobData.JobNameSnakeCase + "_worker_pool.go"

		implementWorkerPoolPrompt := &survey.Confirm{
			Message: fmt.Sprintf("Do you want to implement a worker pool (%s) for %s ?", customJobData.WorkerPoolFileName, customJobData.JobNamePascalCase),
			Default: false,
		}
		if err := survey.AskOne(implementWorkerPoolPrompt, &customJobData.CreateWorkerPool); err != nil {
			utils.HandleError(err)
		}

		if customJobData.CreateWorkerPool {

			customJobData.WorkerPoolNamePascalCase = customJobData.JobNamePascalCase + "WorkerPool"
			customJobData.WorkerPoolNameSnakeCase = customJobData.JobNameSnakeCase + "_worker_pool"

			for {
				workerPoolFileExists := utils.FileExists(path.Join(cli_config.CliConfig.WorkersFolderPath, customJobData.WorkerPoolFileName))
				if workerPoolFileExists {

					options := []string{"Overwrite", fmt.Sprintf("Rename %v", customJobData.WorkerPoolFileName)}
					var selectedOption string
					workerPoolOverwriteStrategyPrompt := &survey.MultiSelect{
						Message: fmt.Sprintf("Worker pool file %s already exists, please specify if you want to rename your custom worker pool or to overwrite existing one", customJobData.WorkerPoolFileName),
						Options: options,
					}
					if err := survey.AskOne(workerPoolOverwriteStrategyPrompt, &selectedOption); err != nil {
						utils.HandleError(err)
					}

					if selectedOption == "Overwrite" {
						customJobData.WorkerPoolOverwrite = true
					} else {
						for {
							if err := survey.AskOne(&survey.Input{
								Message: "Please type in custom worker pool name (snake_case), keep in mind that it will get a suffix _worker_pool.go automatically:",
								Default: customJobData.WorkerPoolNameSnakeCase,
							}, &customJobData.WorkerPoolNameSnakeCase); err != nil {
								utils.HandleError(err)
							}

							if !utils.IsSnakeCase(customJobData.WorkerPoolNameSnakeCase) {
								fmt.Printf("🛑 %s is not in snake case\n", customJobData.WorkerPoolNameSnakeCase)
								continue
							}

							customJobData.WorkerPoolNameSnakeCase = customJobData.WorkerPoolNameSnakeCase + "_worker_pool"
							customJobData.WorkerPoolNamePascalCase = utils.SnakeToPascal(customJobData.WorkerPoolNameSnakeCase)
							customJobData.WorkerPoolNameCamelCase = utils.SnakeToCamel(customJobData.WorkerPoolNameSnakeCase)
							customJobData.WorkerPoolFileName = customJobData.WorkerPoolNameSnakeCase + "_worker_pool.go"
							customJobData.WorkerName = customJobData.JobNamePascalCase + "Worker"
						}
					}
				} else {
					customJobData.WorkerPoolExists = workerPoolFileExists
					break
				}
			}

		}

		break
	}

	err := workerize_utils.GenerateCustomJobMetadataFile(customJobData)
	if err != nil {
		utils.HandleError(err, "Error generating custom job metadata file")
	}

	if !customJobData.AlreadyExists {
		err = workerize_utils.AddCustomJobToBaseJob(customJobData)
		if err != nil {
			utils.HandleError(err, "Error adding custom job to base job")
		}
	}

	if customJobData.CreateWorkerPool {

		if !customJobData.WorkerPoolExists || (customJobData.WorkerPoolExists && customJobData.WorkerPoolOverwrite) {
			// kreiraj
		}
	}

	return
}
