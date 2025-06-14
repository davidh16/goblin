package service

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/davidh16/goblin/cli_config"
	central_service "github.com/davidh16/goblin/commands/service/flags/central-service"
	"github.com/davidh16/goblin/utils"
	"github.com/davidh16/goblin/utils/controller_utils"
	"github.com/davidh16/goblin/utils/repo_utils"
	"github.com/davidh16/goblin/utils/service_utils"
	"github.com/spf13/cobra"
	"path"
)

var (
	CentralServiceFlag bool
)

var ServiceCmd = &cobra.Command{
	Use:   "service",
	Short: "Generate custom service",
	Run: func(cmd *cobra.Command, args []string) {
		if CentralServiceFlag {
			central_service.GenerateCentralService()
		} else {
			serviceCmdHandler()
		}
	},
}

func serviceCmdHandler() {

	serviceData := service_utils.NewServiceData()
	defaultServiceName := "my_service_file"
	for {
		if err := survey.AskOne(&survey.Input{
			Message: "Please type the services file name (snake_case) :",
			Default: defaultServiceName,
		}, &serviceData.ServiceNameSnakeCase); err != nil {
			utils.HandleError(err)
		}

		if !utils.IsSnakeCase(serviceData.ServiceNameSnakeCase) {
			fmt.Printf("🛑 %s is not in snake case\n", serviceData.ServiceNameSnakeCase)
			continue
		}

		defaultServiceName = serviceData.ServiceNameSnakeCase

		var confirm bool
		confirmPrompt := &survey.Confirm{
			Message: fmt.Sprintf("You are about to create a service file named %s_service.go, do you want to continue ?", serviceData.ServiceNameSnakeCase),
		}
		if err := survey.AskOne(confirmPrompt, &confirm); err != nil {
			utils.HandleError(err)
		}

		if !confirm {
			continue
		}

		serviceData.ServiceEntity = utils.SnakeToPascal(serviceData.ServiceNameSnakeCase)
		serviceData.ServiceFullName = serviceData.ServiceEntity + "Service"
		serviceData.ServiceFileName = serviceData.ServiceNameSnakeCase + "_service.go"
		serviceData.ServiceFilePath = path.Join(cli_config.CliConfig.ServicesFolderPath, serviceData.ServiceFileName)

		if utils.FileExists(serviceData.ServiceFilePath) {
			var confirmOverwrite bool
			confirmPrompt = &survey.Confirm{
				Message: fmt.Sprintf("%s service already exists. Do you want to overwrite it ?", serviceData.ServiceFileName),
				Default: false,
			}
			if err := survey.AskOne(confirmPrompt, &confirmOverwrite); err != nil {
				utils.HandleError(err)
			}

			if !confirmOverwrite {
				continue
			}

			if confirmOverwrite {
				confirmPrompt = &survey.Confirm{
					Message: fmt.Sprintf("Are you sure you want to overwrite %s service ?", serviceData.ServiceFileName),
					Default: false,
				}
				if err := survey.AskOne(confirmPrompt, &confirmOverwrite); err != nil {
					utils.HandleError(err)
				}
			}

			if !confirmOverwrite {
				continue
			}
		}
		break
	}

	serviceData.CentralServiceExists = utils.FileExists(path.Join(cli_config.CliConfig.ServicesFolderPath, "central_service.go"))

	existingRepos, err := service_utils.ListExistingRepos()
	if err != nil {
		utils.HandleError(err, "Unable to list existing repositories")
	}
	existingReposMap := make(map[string]*repo_utils.RepoData)
	for _, repo := range existingRepos {
		existingReposMap[repo.RepoFullName] = &repo
	}

	// service_utils.RepoOptionsStrategyMap keys are used to list options for choosing repo strategy, if there are no existing repos, that key (option) has to be removed from the map
	if len(existingRepos) == 0 {
		delete(service_utils.RepoOptionsStrategyMap, service_utils.RepoStrategyOptionsMap[service_utils.RepoStrategyExistingRepo])
	}

	var repoStrategyChosenOption string
	err = survey.AskOne(&survey.Select{
		Message: "Choose repo strategy:",
		Options: utils.Keys(service_utils.RepoOptionsStrategyMap),
	}, &repoStrategyChosenOption)
	if err != nil {
		utils.HandleError(err)
	}

	serviceData.RepoStrategy = service_utils.RepoOptionsStrategyMap[repoStrategyChosenOption]

	if serviceData.RepoStrategy == service_utils.RepoStrategyExistingRepo {
		var chosenRepos []string
		err = survey.AskOne(&survey.MultiSelect{
			Message: "Select a repo to use:",
			Options: utils.Keys(existingReposMap),
		}, &chosenRepos)
		if err != nil {
			utils.HandleError(err)
		}

		for _, repo := range chosenRepos {
			serviceData.RepoData = append(serviceData.RepoData, *existingReposMap[repo])
		}
	}

	if serviceData.RepoStrategy == service_utils.RepoStrategyNewRepo {
		serviceData.RepoData = append(serviceData.RepoData, *service_utils.PrepareRepo())
	}

	var toImplement bool
	switch serviceData.RepoStrategy {
	case service_utils.RepoStrategyNewRepo:
		var decision string
		prompt := &survey.Select{
			Message: service_utils.GenerateImplementProxyMethodsNowQuestionWithExistingRepoMethodsPreview(&serviceData.RepoData[0], serviceData.RepoData[0].SelectedRepoMethodsToImplement),
			Options: []string{
				"Yes, choose methods to implement",
				"No, skip this step",
			},
		}
		err = survey.AskOne(prompt, &decision)
		if err != nil {
			utils.HandleError(err)
		}

		toImplement = decision == "Yes, choose methods to implement"

		if toImplement {
			selectedServiceProxyMethodsPrompt := &survey.MultiSelect{
				Message: "Which service proxy methods do you want to implement?\n  [Press enter without selecting any of the options to skip]\n",
				Options: serviceData.RepoData[0].SelectedRepoMethodsToImplement,
			}
			err = survey.AskOne(selectedServiceProxyMethodsPrompt, &serviceData.SelectedServiceProxyMethodToImplement)
			if err != nil {
				utils.HandleError(err)
			}
		}

	case service_utils.RepoStrategyExistingRepo:

		for _, repo := range serviceData.RepoData {

			existingRepoMethods, err := service_utils.ListExistingRepoMethods(&repo)
			if err != nil {
				utils.HandleError(err)
			}

			var decision string
			prompt := &survey.Select{
				Message: service_utils.GenerateImplementProxyMethodsNowQuestionWithExistingRepoMethodsPreview(&repo, existingRepoMethods),
				Options: []string{
					"Yes, choose methods to implement",
					"No, skip this step",
				},
			}
			err = survey.AskOne(prompt, &decision)
			if err != nil {
				utils.HandleError(err)
			}
			toImplement = decision == "Yes, choose methods to implement"
			if toImplement {
				selectMethodsToImplementPrompt := &survey.MultiSelect{
					Message: "Which methods do you want to implement?\n  [Press enter without selecting any of the options to skip]\n",
					Options: existingRepoMethods,
				}
				err = survey.AskOne(selectMethodsToImplementPrompt, &serviceData.SelectedServiceProxyMethodToImplement)
				if err != nil {
					utils.HandleError(err)
				}
			}
		}
	case service_utils.RepoStrategyNoImplementation:
		serviceData.RepoData = nil
	default:
		utils.HandleError(fmt.Errorf("invalid repo strategy: %s", serviceData.RepoStrategy))
	}

	if serviceData.RepoStrategy == service_utils.RepoStrategyNewRepo {
		err = service_utils.ExecuteCreateRepo(serviceData.RepoData)
		if err != nil {
			utils.HandleError(err)
		}
	}

	if !serviceData.CentralServiceExists {
		central_service.GenerateCentralService()
	}

	if !utils.FileExists(serviceData.ServiceFilePath) {
		err = service_utils.AddNewServiceToCentralService(serviceData)
		if err != nil {
			utils.HandleError(err)
		}
	}

	err = service_utils.CreateService(serviceData)
	if err != nil {
		utils.HandleError(err)
	}

	if serviceData.RepoStrategy != service_utils.RepoStrategyNoImplementation {
		err = service_utils.AddRepoToService(serviceData)
		if err != nil {
			utils.HandleError(err)
		}
	}

	if len(serviceData.SelectedServiceProxyMethodToImplement) > 0 {
		err = service_utils.CopyRepoMethodsToService(serviceData, serviceData.SelectedServiceProxyMethodToImplement)
		if err != nil {
			utils.HandleError(err)
		}
	}

	var toInjectServiceToController bool
	injectPrompt := &survey.Confirm{
		Message: "Do you wish to inject this service to a controller ?",
		Default: true,
	}
	err = survey.AskOne(injectPrompt, &toInjectServiceToController)
	if err != nil {
		utils.HandleError(err)
	}

	if toInjectServiceToController {

		existingControllers, err := controller_utils.ListExistingControllers()
		if err != nil {
			utils.HandleError(err, "Unable to list existing services")
		}
		existingControllersMap := make(map[string]*controller_utils.ControllerData)
		for _, existingController := range existingControllers {
			existingController.ServiceData = []service_utils.ServiceData{*serviceData}
			existingControllersMap[existingController.ControllerFullName] = &existingController
		}

		var selectedController string
		err = survey.AskOne(&survey.Select{
			Message: "Select a controller to inject repo to:",
			Options: utils.Keys(existingControllersMap),
		}, &selectedController)
		if err != nil {
			utils.HandleError(err)
		}

		fmt.Println(existingControllersMap[selectedController].ControllerFullName)

		err = controller_utils.AddServiceToController(existingControllersMap[selectedController])
		if err != nil {
			utils.HandleError(err)
		}
	}

	fmt.Println(fmt.Sprintf("✅ %s service generated successfully.", serviceData.ServiceEntity))
	return
}
