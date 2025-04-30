package service

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"goblin/cli_config"
	central_service "goblin/commands/service/flags/central-service"
	"goblin/utils"
	"goblin/utils/repo_utils"
	"goblin/utils/service_utils"
	"path"
)

var (
	CentralService bool
)

var ServiceCmd = &cobra.Command{
	Use:   "service",
	Short: "Generate custom service",
	Run: func(cmd *cobra.Command, args []string) {
		if CentralService {
			central_service.GenerateCentralService()
		} else {
			serviceCmdHandler()
		}
	},
}

func serviceCmdHandler() {

	serviceData := service_utils.NewServiceData()

	for {
		if err := survey.AskOne(&survey.Input{
			Message: "Please type the services file name (snake_case) :",
			Default: "my_service_file",
		}, &serviceData.ServiceNameSnakeCase); err != nil {
			utils.HandleError(err)
		}

		if !utils.IsSnakeCase(serviceData.ServiceNameSnakeCase) {
			fmt.Printf("🛑 %s is not in snake case\n", serviceData.ServiceNameSnakeCase)
			continue
		}

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
		err = survey.AskOne(&survey.Select{
			Message: "Select a repo to use:",
			Options: utils.Keys(existingReposMap),
		}, &serviceData.RepoData.RepoEntity)
		if err != nil {
			utils.HandleError(err)
		}
	}

	if serviceData.RepoStrategy == service_utils.RepoStrategyNewRepo {
		serviceData.RepoData = service_utils.PrepareRepo()
	}

	var toImplement bool
	var selectedMethods []string
	switch serviceData.RepoStrategy {
	case service_utils.RepoStrategyNewRepo:
		var decision string
		prompt := &survey.Select{
			Message: service_utils.GenerateImplementProxyMethodsNowQuestionWithExistingRepoMethodsPreview(serviceData.RepoData.SelectedRepoMethodsToImplement),
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
				Options: serviceData.RepoData.SelectedRepoMethodsToImplement,
			}
			err = survey.AskOne(selectedServiceProxyMethodsPrompt, &serviceData.SelectedServiceProxyMethodToImplement)
			if err != nil {
				utils.HandleError(err)
			}
		}

	case service_utils.RepoStrategyExistingRepo:

		existingRepoMethods, err := service_utils.ListExistingRepoMethods(serviceData.RepoData)
		if err != nil {
			utils.HandleError(err)
		}

		var decision string
		prompt := &survey.Select{
			Message: service_utils.GenerateImplementProxyMethodsNowQuestionWithExistingRepoMethodsPreview(existingRepoMethods),
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

	if len(selectedMethods) > 0 {
		err = service_utils.CopyRepoMethodsToService(serviceData, selectedMethods)
		if err != nil {
			utils.HandleError(err)
		}
	}

	fmt.Println(fmt.Sprintf("✅ %s service generated successfully.", serviceData.ServiceEntity))
	return
}
