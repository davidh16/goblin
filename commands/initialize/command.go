package initialize

import (
	"github.com/AlecAivazis/survey/v2"
	"github.com/davidh16/goblin/cli_config"
	central_controller "github.com/davidh16/goblin/commands/controller/flags/central-controller"
	central_repo "github.com/davidh16/goblin/commands/repo/flags/central-repo"
	central_service "github.com/davidh16/goblin/commands/service/flags/central-service"
	"github.com/davidh16/goblin/utils"
	"github.com/davidh16/goblin/utils/database_utils"
	"github.com/davidh16/goblin/utils/initialize_utils"
	"github.com/davidh16/goblin/utils/middleware_utils"
	"github.com/davidh16/goblin/utils/router_utils"
	"github.com/spf13/cobra"
	"os"
	"path"
	"strings"
	"text/template"
)

var InitializeCmd = &cobra.Command{
	Use:   "initialize",
	Short: "Generate boilerplate code.",
	Run: func(cmd *cobra.Command, args []string) {
		initCmdHandler()
	},
}

func initCmdHandler() {

	initData := initialize_utils.NewInitData()

	implementPrompt := &survey.Confirm{
		Message: "Do you want to implement central service?",
		Default: true,
	}
	err := survey.AskOne(implementPrompt, &initData.ImplementCentralService)
	if err != nil {
		utils.HandleError(err)
	}

	implementPrompt = &survey.Confirm{
		Message: "Do you want to implement central repository?",
		Default: true,
	}
	err = survey.AskOne(implementPrompt, &initData.ImplementCentralRepository)
	if err != nil {
		utils.HandleError(err)
	}

	var selectedDatabaseNames []string
	if initData.ImplementCentralRepository {
		// baze
		for {
			selectDatabasesPrompt := &survey.MultiSelect{
				Message: "Which databases do you want to use?",
				Options: database_utils.GetSortedDatabaseOptions(),
			}
			err := survey.AskOne(selectDatabasesPrompt, &selectedDatabaseNames)
			if err != nil {
				utils.HandleError(err)
			}

			if len(selectedDatabaseNames) != 0 {
				initData.ImplementDatabase = true
				break
			}
		}
	}

	// router
	routerData := router_utils.NewRouterData()

	var selectedMiddlewares []string
	selectMiddlewaresPrompt := &survey.MultiSelect{
		Message: "Which middlewares do you want to inject into your router?\n  [Press enter without selecting any of the options to skip]\n",
		Options: middleware_utils.MiddlewareOptions,
	}
	err = survey.AskOne(selectMiddlewaresPrompt, &selectedMiddlewares)
	if err != nil {
		utils.HandleError(err)
	}

	// ask for server port
	var serverPort string
	if err = survey.AskOne(&survey.Input{
		Message: "Please type in server port you want to use :",
		Default: "8080",
	}, &serverPort); err != nil {
		utils.HandleError(err)
	}

	workingDirectory, err := os.Getwd()
	if err != nil {
		utils.HandleError(err)
	}
	envFilePath := path.Join(workingDirectory, ".env")
	envFile, err := os.OpenFile(envFilePath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		utils.HandleError(err)
	}
	err = utils.WriteToEnvFile(envFile, map[string]string{
		"SERVER_BIND_PORT":    serverPort,
		"SERVER_BIND_ADDRESS": "0.0.0.0",
	})
	if err != nil {
		utils.HandleError(err)
	}

	if initData.ImplementDatabase {
		err = initialize_utils.ExecuteDatabases(selectedDatabaseNames)
		if err != nil {
			utils.HandleError(err)
		}
	}

	if initData.ImplementCentralRepository {
		central_repo.GenerateCentralRepo()
	}

	if initData.ImplementCentralService {
		central_service.GenerateCentralService()
	}

	err = central_controller.GenerateCentralController()
	if err != nil {
		utils.HandleError(err)
	}

	// execute router
	err = initialize_utils.ExecuteRouter(routerData, selectedMiddlewares)
	if err != nil {
		utils.HandleError(err)
	}

	// execute main
	tmpl, err := template.ParseFiles(initialize_utils.MainTemplatePath)
	if err != nil {
		utils.HandleError(err)
	}

	f, err := os.Create(path.Join(workingDirectory, "main1.go"))
	if err != nil {
		utils.HandleError(err)
	}
	defer f.Close()

	templateData := struct {
		RouterPackage       string
		RouterPackageImport string

		ControllersPackageImport string
		ControllersPackage       string

		ServicesPackageImport string
		ServicesPackage       string

		RepositoriesPackageImport string
		RepositoriesPackage       string

		DatabasesPackageImport string
		DatabasesPackage       string

		ImplementCentralRepository bool
		ImplementCentralService    bool

		LoggerImplemented   bool
		LoggerPackage       string
		LoggerPackageImport string
	}{
		RouterPackage:       strings.Split(cli_config.CliConfig.RouterFolderPath, "/")[len(strings.Split(cli_config.CliConfig.RouterFolderPath, "/"))-1],
		RouterPackageImport: path.Join(cli_config.CliConfig.ProjectName, cli_config.CliConfig.RouterFolderPath),

		ControllersPackageImport: path.Join(cli_config.CliConfig.ProjectName, cli_config.CliConfig.ControllersFolderPath),
		ControllersPackage:       strings.Split(cli_config.CliConfig.ControllersFolderPath, "/")[len(strings.Split(cli_config.CliConfig.ControllersFolderPath, "/"))-1],

		ServicesPackageImport: path.Join(cli_config.CliConfig.ProjectName, cli_config.CliConfig.ServicesFolderPath),
		ServicesPackage:       strings.Split(cli_config.CliConfig.ServicesFolderPath, "/")[len(strings.Split(cli_config.CliConfig.ServicesFolderPath, "/"))-1],

		RepositoriesPackageImport: path.Join(cli_config.CliConfig.ProjectName, cli_config.CliConfig.RepositoriesFolderPath),
		RepositoriesPackage:       strings.Split(cli_config.CliConfig.RepositoriesFolderPath, "/")[len(strings.Split(cli_config.CliConfig.RepositoriesFolderPath, "/"))-1],

		DatabasesPackage:       strings.Split(cli_config.CliConfig.DatabaseInstancesFolderPath, "/")[len(strings.Split(cli_config.CliConfig.DatabaseInstancesFolderPath, "/"))-1],
		DatabasesPackageImport: path.Join(cli_config.CliConfig.ProjectName, cli_config.CliConfig.DatabaseInstancesFolderPath),

		ImplementCentralRepository: initData.ImplementCentralRepository,
		ImplementCentralService:    initData.ImplementCentralService,

		LoggerImplemented:   routerData.LoggingMiddleware,
		LoggerPackage:       strings.Split(cli_config.CliConfig.LoggerFolderPath, "/")[len(strings.Split(cli_config.CliConfig.LoggerFolderPath, "/"))-1],
		LoggerPackageImport: path.Join(cli_config.CliConfig.ProjectName, cli_config.CliConfig.LoggerFolderPath),
	}

	err = tmpl.Execute(f, templateData)
	if err != nil {
		utils.HandleError(err)
	}

	return
}
