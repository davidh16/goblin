package logger

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/davidh16/goblin/cli_config"
	"github.com/davidh16/goblin/utils"
	"github.com/davidh16/goblin/utils/logger_utils"
	"github.com/spf13/cobra"
	"os"
	"path"
)

var LoggerCmd = &cobra.Command{
	Use:   "logger",
	Short: "Generate logger logic",
	Run: func(cmd *cobra.Command, args []string) {
		loggerCmdHandler()
	},
}

func loggerCmdHandler() {
	// if exists - logger already exists, by continuing you will overwrite it. do you wish to continue ?

	alreadyExists := utils.FileExists(path.Join(cli_config.CliConfig.LoggerFolderPath, logger_utils.LoggerFileName))

	if alreadyExists {
		var confirmContinue bool
		confirmPrompt := &survey.Confirm{
			Message: "Logger already exists. Do you want to overwrite?",
		}
		if err := survey.AskOne(confirmPrompt, &confirmContinue); err != nil {
			utils.HandleError(err)
		}

		if !confirmContinue {
			return
		}
	}

	loggerDirectoryExists := utils.FileExists(cli_config.CliConfig.LoggerFolderPath)
	if !loggerDirectoryExists {
		err := os.MkdirAll(cli_config.CliConfig.LoggerFolderPath, 0755) // 0755 = rwxr-xr-x
		if err != nil {
			fmt.Println("Error creating folder:", err)
		}
	}

	// generate
	err := logger_utils.GenerateLogger()
	if err != nil {
		utils.HandleError(err)
	}

	fmt.Println("✅ Logger generated successfully")
	return
}
