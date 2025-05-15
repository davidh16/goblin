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

	var jobName string
	for {
		if err := survey.AskOne(&survey.Input{
			Message: "Please type the job file name (snake_case), keep in mind that it will get a suffix _job.go automatically:",
			Default: "my_custom_job",
		}, &customJobData.JobNameSnakeCase); err != nil {
			utils.HandleError(err)
		}

		if !utils.IsSnakeCase(jobName) {
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
			Message: fmt.Sprintf("You are about to create a repo file named %s, do you want to continue ?", customJobData.JobFileName),
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
			}

			if !overwriteConfirmed {
				continue
			}
		}

		break
	}

	err := workerize_utils.GenerateCustomJobMetadataFile(customJobData)
	if err != nil {
		utils.HandleError(err, "Error generating custom job metadata file")
	}

	err = workerize_utils.AddCustomJobToBaseJob(customJobData)
	if err != nil {
		utils.HandleError(err, "Error adding custom job to base job")
	}

}
