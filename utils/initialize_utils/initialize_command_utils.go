package initialize_utils

import (
	"errors"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/davidh16/goblin/utils"
	"github.com/davidh16/goblin/utils/database_utils"
	"github.com/davidh16/goblin/utils/middleware_utils"
	"github.com/davidh16/goblin/utils/router_utils"
	"github.com/samber/lo"
	"os"
	"path"
)

type InitData struct {
	ImplementDatabase          bool
	ImplementCentralRepository bool
	ImplementCentralService    bool
}

func NewInitData() *InitData {
	return &InitData{}
}

func ExecuteDatabases(selectedDatabaseNames []string) error {
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

func ExecuteRouter(routerData *router_utils.RouterData, selectedMiddlewares []string) error {
	if len(selectedMiddlewares) > 0 {
		routerData.ImplementMiddlewares = true
		err := middleware_utils.GenerateMiddlewares(selectedMiddlewares)
		if err != nil {
			utils.HandleError(err)
		}
	}
	for _, m := range selectedMiddlewares {
		switch m {
		case "RecoverMiddleware":
			routerData.RecoverMiddleware = true
		case "LoggingMiddleware":
			routerData.LoggingMiddleware = true
		case "RateLimiterMiddleware":
			routerData.RateLimiterMiddleware = true
		case "AllowOriginMiddleware":
			routerData.AllowOriginMiddleware = true
		}
	}

	err := router_utils.GenerateRouter(routerData)
	if err != nil {
		utils.HandleError(err)
	}

	fmt.Println("✅ Router generated successfully.")
	return nil
}
