package migration

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/davidh16/goblin/cli_config"
	"github.com/davidh16/goblin/utils"
	"github.com/davidh16/goblin/utils/migration_utils"
	"github.com/spf13/cobra"
	"path"
	"time"
)

var MigrationCmd = &cobra.Command{
	Use:   "migration",
	Short: "Generate custom migration",
	Run: func(cmd *cobra.Command, args []string) {
		migrationCmdHandler()
	},
}

func migrationCmdHandler() {

	migrationData := migration_utils.NewMigrationData()
	defaultMigrationName := "my_custom_migration"
	for {
		if err := survey.AskOne(&survey.Input{
			Message: "Please type the migration file name (snake_case), keep in mind it will get a timestamp prefix :",
			Default: defaultMigrationName,
		}, &migrationData.MigrationNameSnakeCase); err != nil {
			utils.HandleError(err)
		}

		if !utils.IsSnakeCase(migrationData.MigrationNameSnakeCase) {
			fmt.Printf("🛑 %s is not in snake case\n", migrationData.MigrationNameSnakeCase)
			continue
		}

		example := time.Now().Format("20060102150405") + "_" + migrationData.MigrationNameSnakeCase + "_(up/down).sql"
		migrationData.MigrationUpFileName = time.Now().Format("20060102150405") + "_" + migrationData.MigrationNameSnakeCase + "_up.sql"
		migrationData.MigrationDownFileName = time.Now().Format("20060102150405") + "_" + migrationData.MigrationNameSnakeCase + "_down.sql"

		var confirm bool
		confirmPrompt := &survey.Confirm{
			Message: fmt.Sprintf("You are about to create a migration up and down files named %s, do you want to continue ?", example),
		}
		if err := survey.AskOne(confirmPrompt, &confirm); err != nil {
			utils.HandleError(err)
		}

		if !confirm {
			continue
		}

		break
	}

	migrationData.MigrationUpFileFullPath = path.Join(cli_config.CliConfig.MigrationsFolderPath, migrationData.MigrationUpFileName)
	migrationData.MigrationDownFileFullPath = path.Join(cli_config.CliConfig.MigrationsFolderPath, migrationData.MigrationDownFileName)

	err := migration_utils.GenerateMigrationFiles(migrationData)
	if err != nil {
		utils.HandleError(err)
	}

	fmt.Println(fmt.Sprintf("✅ %s migration generated successfully.", migrationData.MigrationUpFileName))
	fmt.Println(fmt.Sprintf("✅ %s migration generated successfully.", migrationData.MigrationDownFileName))
	return
}
