package initialize_utils

import (
	"errors"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/davidh16/goblin/cli_config"
	central_repo "github.com/davidh16/goblin/commands/repo/flags/central-repo"
	"github.com/davidh16/goblin/templates"
	"github.com/davidh16/goblin/utils"
	"github.com/davidh16/goblin/utils/controller_utils"
	"github.com/davidh16/goblin/utils/database_utils"
	"github.com/davidh16/goblin/utils/middleware_utils"
	"github.com/davidh16/goblin/utils/repo_utils"
	"github.com/davidh16/goblin/utils/router_utils"
	"github.com/davidh16/goblin/utils/service_utils"
	"github.com/samber/lo"
	"os"
	"path"
	"strings"
	"text/template"
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
			return err
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
		return err
	}

	fmt.Println("✅ Router generated successfully.")
	return nil
}

func ExecuteCentralService(implementRepo bool) error {
	centralServicePath := path.Join(cli_config.CliConfig.ServicesFolderPath, "central_service.go")

	funcMap := template.FuncMap{
		"GetProjectName": utils.GetProjectName,
	}

	tmpl, err := template.New(service_utils.CentralServiceTemplateName).Funcs(funcMap).ParseFS(templates.Files, service_utils.CentralServiceTemplatePath)
	if err != nil {
		utils.HandleError(err)
	}

	f, err := os.Create(centralServicePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err = os.MkdirAll(cli_config.CliConfig.ServicesFolderPath, 0755) // 0755 = rwxr-xr-x
			if err != nil {
				return err
			}
			f, err = os.Create(centralServicePath)
			if err != nil {
				return err
			}
		}
	}
	defer f.Close()

	templateData := struct {
		CentralRepoExists bool
		RepoPackage       string
		RepoPackageImport string
		ServicePackage    string
	}{
		RepoPackage:       strings.Split(cli_config.CliConfig.RepositoriesFolderPath, "/")[len(strings.Split(cli_config.CliConfig.RepositoriesFolderPath, "/"))-1],
		RepoPackageImport: cli_config.CliConfig.RepositoriesFolderPath,
		ServicePackage:    strings.Split(cli_config.CliConfig.ServicesFolderPath, "/")[len(strings.Split(cli_config.CliConfig.ServicesFolderPath, "/"))-1],
		CentralRepoExists: implementRepo,
	}

	err = tmpl.Execute(f, templateData)
	if err != nil {
		return err
	}

	err = service_utils.AddCentralServiceToCentralControllerConstructor()
	if err != nil {
		utils.HandleError(err)
	}

	fmt.Println("✅ Central service generated successfully.")
	return nil
}

func ExecuteCentralRepo() error {
	centralRepoPath := path.Join(cli_config.CliConfig.RepositoriesFolderPath, "central_repo.go")

	unitOFWorkRepoPath := path.Join(cli_config.CliConfig.RepositoriesFolderPath, "unit_of_work.go")
	central_repo.GenerateUnitOfWorkRepoUtil(unitOFWorkRepoPath)

	tmpl, err := template.ParseFS(templates.Files, central_repo.CentralRepoTemplatePath)
	if err != nil {
		utils.HandleError(err)
	}

	err = os.MkdirAll(cli_config.CliConfig.RepositoriesFolderPath, 0755) // 0755 = rwxr-xr-x
	if err != nil {
		return err
	}

	f, err := os.Create(centralRepoPath)
	if err != nil {
		return err
	}
	defer f.Close()

	templateData := struct {
		RepoPackage string
	}{
		RepoPackage: strings.Split(cli_config.CliConfig.RepositoriesFolderPath, "/")[len(strings.Split(cli_config.CliConfig.RepositoriesFolderPath, "/"))-1],
	}

	err = tmpl.Execute(f, templateData)
	if err != nil {
		return err
	}

	err = repo_utils.AddCentralRepoToCentralServiceConstructor()
	if err != nil {
		return err
	}

	fmt.Println("✅ Central repository generated successfully.")
	return nil
}

func ExecuteCentralController(implementCentralService bool) error {

	centralControllerPath := path.Join(cli_config.CliConfig.ControllersFolderPath, "central_controller.go")
	err := os.MkdirAll(cli_config.CliConfig.ControllersFolderPath, 0755) // 0755 = rwxr-xr-x
	if err != nil {
		return err
	}

	f, err := os.Create(centralControllerPath)
	if err != nil {
		return err
	}
	defer f.Close()

	funcMap := template.FuncMap{
		"GetProjectName": utils.GetProjectName,
	}

	tmpl, err := template.New(controller_utils.CentralControllerTemplateName).Funcs(funcMap).ParseFS(templates.Files, controller_utils.CentralControllerTemplatePath)
	if err != nil {
		return err
	}

	templateData := struct {
		Package              string
		CentralServiceExists bool
		ServicePackage       string
		ServicePackageImport string
	}{
		Package:              strings.Split(cli_config.CliConfig.ControllersFolderPath, "/")[len(strings.Split(cli_config.CliConfig.ControllersFolderPath, "/"))-1],
		ServicePackageImport: cli_config.CliConfig.ServicesFolderPath,
		ServicePackage:       strings.Split(cli_config.CliConfig.ServicesFolderPath, "/")[len(strings.Split(cli_config.CliConfig.ServicesFolderPath, "/"))-1],
		CentralServiceExists: implementCentralService,
	}

	err = tmpl.Execute(f, templateData)
	if err != nil {
		return err
	}

	return nil

}
