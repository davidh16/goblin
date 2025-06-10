package database

import (
	"errors"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"goblin/cli_config"
	"goblin/utils"
	"goblin/utils/database_utils"
	"os"
	"path"
	"strings"
)

var DatabaseCmd = &cobra.Command{
	Use:   "database",
	Short: "Select a database to use",
	Run: func(cmd *cobra.Command, args []string) {
		databaseCmdHandler()
	},
}

func databaseCmdHandler() {
	var selectedDatabaseNames []string
	var preselectedOptions []string

	for {
		selectDatabasesPrompt := &survey.MultiSelect{
			Message: "Which databases do you want to use?",
			Options: database_utils.GetSortedDatabaseOptions(),
			Default: preselectedOptions,
		}
		err := survey.AskOne(selectDatabasesPrompt, &selectedDatabaseNames)
		if err != nil {
			utils.HandleError(err)
		}
		preselectedOptions = selectedDatabaseNames

		var existingDatabaseInstances []string
		for _, databaseName := range selectedDatabaseNames {
			if utils.FileExists(path.Join(cli_config.CliConfig.DatabaseInstancesFolderPath, database_utils.DatabaseOptionInstanceDefaultFileNamesMap[database_utils.DatabaseNameOptionsMap[databaseName]])) {
				existingDatabaseInstances = append(existingDatabaseInstances, databaseName)
			}
		}

		if len(existingDatabaseInstances) != 0 {

			var propmtText string
			if len(existingDatabaseInstances) == 1 {
				propmtText = fmt.Sprintf("%s instance already exists, do you want to overwrite it?", existingDatabaseInstances[0])
			} else {
				propmtText = fmt.Sprintf("%s instances already exist, do you want to overwrite them?", strings.Join(existingDatabaseInstances, ", "))
			}

			var confirmOverwrite bool
			confirmOverwritePrompt := &survey.Confirm{
				Message: propmtText,
				Default: false,
			}
			err = survey.AskOne(confirmOverwritePrompt, &confirmOverwrite)
			if err != nil {
				utils.HandleError(err)
			}

			if !confirmOverwrite {
				selectedDatabaseNames = []string{}
				continue
			}

			confirmOverwritePrompt = &survey.Confirm{
				Message: "Are you sure you want to overwrite?",
				Default: false,
			}
			err = survey.AskOne(confirmOverwritePrompt, &confirmOverwrite)
			if err != nil {
				utils.HandleError(err)
			}

			if !confirmOverwrite {
				selectedDatabaseNames = []string{}
				continue
			}
		}

		break
	}

	selectedDatabaseOptions := lo.Map(selectedDatabaseNames, func(item string, index int) database_utils.DatabaseOption {
		return database_utils.DatabaseNameOptionsMap[item]
	})

	var databases []database_utils.DatabaseData
	for _, databaseOption := range selectedDatabaseOptions {
		var databasePort string
		if err := survey.AskOne(&survey.Input{
			Message: fmt.Sprintf("Please type in %s port you want to use :", database_utils.DatabaseOptionNamesMap[databaseOption]),
			Default: database_utils.DatabaseOptionDefaultPortsMap[databaseOption],
		}, &databasePort); err != nil {
			utils.HandleError(err)
		}

		databases = append(databases, database_utils.DatabaseData{
			DatabaseType: databaseOption,
			Port:         databasePort,
		})
	}

	workingDirectory, err := os.Getwd()
	if err != nil {
		utils.HandleError(err)
	}

	envFilePath := path.Join(workingDirectory, ".env")

	var envFile *os.File
	defer envFile.Close()
	envDataMap := map[string]string{}
	for _, database := range databases {

		envFile, err = os.OpenFile(envFilePath, os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			utils.HandleError(err, "Error opening environment file")
		}

		database_utils.DatabaseOptionDefaultPortsMap[database.DatabaseType] = database.Port

		envData, err := database_utils.GetDatabaseOptionDefaultEnvDataMap(database.DatabaseType)
		if err != nil {
			utils.HandleError(err, fmt.Sprintf("Error getting default env data for %s", database_utils.DatabaseOptionNamesMap[database.DatabaseType]))
		}

		envDataMap = utils.MergeMaps(envDataMap, envData)

		err = database_utils.InitializeDatabaseInstance(database)
		if err != nil {
			utils.HandleError(err, fmt.Sprintf("Error initializing %s database instance", database_utils.DatabaseOptionNamesMap[database.DatabaseType]))
		}
	}

	err = utils.WriteToEnvFile(envFile, envDataMap)
	if err != nil {
		utils.HandleError(err, fmt.Sprintf("Error writing environment file %s", envFilePath))
	}

	return
}

func ImplementRedisAndOtherGormDb() error {
	var selectedDatabaseNames []string
	var preselectedOptions []string
	//var redisImplemented bool
	var gormDatabaseImplemented bool

	for {
		preselectedOptions = []string{database_utils.DatabaseOptionNamesMap[database_utils.Redis]}
		preselectedOptions = utils.Keys(database_utils.DatabaseNameOptionsMap)
		selectDatabasesPrompt := &survey.MultiSelect{
			Message: "Which databases do you want to use?",
			Options: database_utils.GetSortedDatabaseOptions(),
			Default: preselectedOptions,
		}
		err := survey.AskOne(selectDatabasesPrompt, &selectedDatabaseNames)
		if err != nil {
			return err
		}
		preselectedOptions = selectedDatabaseNames

		var existingDatabaseInstances []string
		for _, databaseName := range selectedDatabaseNames {

			//if databaseName == database_utils.DatabaseOptionNamesMap[database_utils.Redis] {
			//	redisImplemented = true
			//}

			if databaseName == database_utils.DatabaseOptionNamesMap[database_utils.PostgresSQL] || databaseName == database_utils.DatabaseOptionNamesMap[database_utils.MariaDB] {
				gormDatabaseImplemented = true
			}

			if utils.FileExists(path.Join(cli_config.CliConfig.DatabaseInstancesFolderPath, database_utils.DatabaseOptionInstanceDefaultFileNamesMap[database_utils.DatabaseNameOptionsMap[databaseName]])) {
				existingDatabaseInstances = append(existingDatabaseInstances, databaseName)
			}
		}

		//if !redisImplemented {
		//	fmt.Println("🔴 Redis must be implemented")
		//	continue
		//}

		if !gormDatabaseImplemented {
			fmt.Println("🔴 At least one persistent database must be implemented")
			continue
		}

		if len(existingDatabaseInstances) != 0 {

			var promptText string
			if len(existingDatabaseInstances) == 1 {
				promptText = fmt.Sprintf("%s instance already exists, do you want to overwrite it?", existingDatabaseInstances[0])
			} else {
				promptText = fmt.Sprintf("%s instances already exist, do you want to overwrite them?", strings.Join(existingDatabaseInstances, ", "))
			}

			var confirmOverwrite bool
			confirmOverwritePrompt := &survey.Confirm{
				Message: promptText,
				Default: false,
			}
			err = survey.AskOne(confirmOverwritePrompt, &confirmOverwrite)
			if err != nil {
				return err
			}

			if !confirmOverwrite {
				selectedDatabaseNames = []string{}
				continue
			}

			confirmOverwritePrompt = &survey.Confirm{
				Message: "Are you sure you want to overwrite?",
				Default: false,
			}
			err = survey.AskOne(confirmOverwritePrompt, &confirmOverwrite)
			if err != nil {
				return err
			}

			if !confirmOverwrite {
				selectedDatabaseNames = []string{}
				continue
			}
		}

		break
	}

	selectedDatabaseOptions := lo.Map(selectedDatabaseNames, func(item string, index int) database_utils.DatabaseOption {
		return database_utils.DatabaseNameOptionsMap[item]
	})

	var databases []database_utils.DatabaseData
	for _, databaseOption := range selectedDatabaseOptions {
		var databasePort string
		if err := survey.AskOne(&survey.Input{
			Message: fmt.Sprintf("Please type in %s port you want to use :", database_utils.DatabaseOptionNamesMap[databaseOption]),
			Default: database_utils.DatabaseOptionDefaultPortsMap[databaseOption],
		}, &databasePort); err != nil {
			return err
		}

		databases = append(databases, database_utils.DatabaseData{
			DatabaseType: databaseOption,
			Port:         databasePort,
		})
	}

	workingDirectory, err := os.Getwd()
	if err != nil {
		return err
	}

	envFilePath := path.Join(workingDirectory, ".env")

	var envFile *os.File
	defer envFile.Close()
	envDataMap := map[string]string{}
	for _, database := range databases {

		envFile, err = os.OpenFile(envFilePath, os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			return errors.New("error opening environment file")
		}

		database_utils.DatabaseOptionDefaultPortsMap[database.DatabaseType] = database.Port

		envData, err := database_utils.GetDatabaseOptionDefaultEnvDataMap(database.DatabaseType)
		if err != nil {
			return errors.New(fmt.Sprintf("Error getting default env data for %s", database_utils.DatabaseOptionNamesMap[database.DatabaseType]))
		}

		envDataMap = utils.MergeMaps(envDataMap, envData)

		err = database_utils.InitializeDatabaseInstance(database)
		if err != nil {
			return errors.New(fmt.Sprintf("Error initializing %s database instance", database_utils.DatabaseOptionNamesMap[database.DatabaseType]))
		}
	}

	err = utils.WriteToEnvFile(envFile, envDataMap)
	if err != nil {
		return errors.New(fmt.Sprintf("Error writing environment file %s", envFilePath))
	}

	return nil
}

func ImplementRedis() error {
	var redisPort string
	if err := survey.AskOne(&survey.Input{
		Message: fmt.Sprintf("Please type in %s port you want to use :", database_utils.DatabaseOptionNamesMap[database_utils.Redis]),
		Default: database_utils.DatabaseOptionDefaultPortsMap[database_utils.Redis],
	}, &redisPort); err != nil {
		return err
	}

	workingDirectory, err := os.Getwd()
	if err != nil {
		return err
	}

	envFilePath := path.Join(workingDirectory, ".env")

	var envFile *os.File
	defer envFile.Close()
	envDataMap := map[string]string{}

	envFile, err = os.OpenFile(envFilePath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return errors.New("Error opening environment file")
	}

	database_utils.DatabaseOptionDefaultPortsMap[database_utils.Redis] = redisPort

	envData, err := database_utils.GetDatabaseOptionDefaultEnvDataMap(database_utils.Redis)
	if err != nil {
		return errors.New("Error getting default env data for redis")
	}

	envDataMap = utils.MergeMaps(envDataMap, envData)

	err = database_utils.InitializeDatabaseInstance(database_utils.DatabaseData{
		DatabaseType: database_utils.Redis,
		Port:         redisPort,
	})
	if err != nil {
		return errors.New("Error initializing redis database instance")
	}

	err = utils.WriteToEnvFile(envFile, envDataMap)
	if err != nil {
		return errors.New("Error writing environment file for redis")
	}

	return nil

}
