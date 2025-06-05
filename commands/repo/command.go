package repo

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"goblin/cli_config"
	"goblin/commands/model"
	central_repo "goblin/commands/repo/flags/central-repo"
	"goblin/utils"
	"goblin/utils/controller_utils"
	"goblin/utils/model_utils"
	"goblin/utils/repo_utils"
	"goblin/utils/service_utils"
	"path"
	"strings"
)

var (
	CentralRepoFlag bool
)

var RepoCmd = &cobra.Command{
	Use:   "repo",
	Short: "Generate custom repository repository",
	Run: func(cmd *cobra.Command, args []string) {
		if CentralRepoFlag {
			central_repo.GenerateCentralRepo()
		} else {
			repoCmdHandler()
		}
	},
}

func repoCmdHandler() {

	repoData := repo_utils.NewRepoData()
	defaultRepoName := "my_repo_file"
	for {
		if err := survey.AskOne(&survey.Input{
			Message: "Please type the repository file name (snake_case) :",
			Default: defaultRepoName,
		}, &repoData.RepoNameSnakeCase); err != nil {
			utils.HandleError(err)
		}

		if !utils.IsSnakeCase(repoData.RepoNameSnakeCase) {
			fmt.Printf("🛑 %s is not in snake case\n", repoData.RepoNameSnakeCase)
			continue
		}

		defaultRepoName = repoData.RepoNameSnakeCase

		var confirmContinue bool
		confirmPrompt := &survey.Confirm{
			Message: fmt.Sprintf("You are about to create a repo file named %s_repo.go, do you want to continue ?", repoData.RepoNameSnakeCase),
		}
		if err := survey.AskOne(confirmPrompt, &confirmContinue); err != nil {
			utils.HandleError(err)
		}

		if !confirmContinue {
			continue
		}

		repoData.RepoEntity = utils.SnakeToPascal(repoData.RepoNameSnakeCase)
		repoData.RepoFullName = repoData.RepoEntity + "Repo"
		repoData.RepoFileName = repoData.RepoNameSnakeCase + "_repo.go"
		repoData.RepoFilePath = path.Join(cli_config.CliConfig.RepositoriesFolderPath, repoData.RepoFileName)

		if utils.FileExists(repoData.RepoFilePath) {
			var overwriteConfirmed bool
			confirmPrompt = &survey.Confirm{
				Message: fmt.Sprintf("%s repository already exists. Do you want to overwrite it ?", repoData.RepoFileName),
				Default: false,
			}
			if err := survey.AskOne(confirmPrompt, &overwriteConfirmed); err != nil {
				utils.HandleError(err)
			}

			if overwriteConfirmed {
				confirmPrompt = &survey.Confirm{
					Message: fmt.Sprintf("Are you sure you want to overwrite %s repository ?", repoData.RepoFileName),
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

	repoData.CentralRepoExists = utils.FileExists(path.Join(cli_config.CliConfig.RepositoriesFolderPath, "central_repo.go"))

	existingModels, err := repo_utils.ListExistingModels()
	if err != nil {
		utils.HandleError(err, "Unable to list existing models")
	}

	options := []string{repo_utils.ModelStrategyOptionsMap[repo_utils.ModelStrategyNewModel]}
	if len(existingModels) > 0 {
		options = append(options, repo_utils.ModelStrategyOptionsMap[repo_utils.ModelStrategyExistingModel])
	}

	var optionChoice string
	err = survey.AskOne(&survey.Select{
		Message: "Choose model strategy:",
		Options: options,
	}, &optionChoice)
	if err != nil {
		utils.HandleError(err)
	}

	repoData.ModelStrategy = repo_utils.ModelOptionsStrategyMap[optionChoice]

	switch repoData.ModelStrategy {
	case repo_utils.ModelStrategyNewModel:
		modelData, err := model_utils.TriggerGetModelNameFlow()
		if err != nil {
			utils.HandleError(err, "Unable to get model name")
		}

		repoData.ModelData = modelData

	case repo_utils.ModelStrategyExistingModel:

		existingModelOptionsModelDataMap := map[string]*model_utils.ModelData{}
		existingModelOptions := lo.Map(existingModels, func(item model_utils.ModelData, index int) string {
			existingModelOptionsModelDataMap[item.ModelEntity] = &item
			return item.ModelEntity
		})

		var selectedModelOption string
		err = survey.AskOne(&survey.Select{
			Message: "Select a model to use:",
			Options: existingModelOptions,
		}, &selectedModelOption)
		if err != nil {
			utils.HandleError(err)
		}

		repoData.ModelData = existingModelOptionsModelDataMap[selectedModelOption]
	default:
		utils.HandleError(fmt.Errorf("invalid model strategy: %d", repoData.ModelStrategy))
	}

	var decision string
	prompt := &survey.Select{
		Message: repo_utils.GenerateImplementRepoMethodsNowQuestion(repoData.ModelData.ModelEntity),
		Options: []string{
			"Yes, choose methods to implement",
			"No, skip this step",
		},
	}
	err = survey.AskOne(prompt, &decision)
	if err != nil {
		utils.HandleError(err)
	}

	toImplementRepoMethods := decision == "Yes, choose methods to implement"

	if toImplementRepoMethods {
		selectMethodsPrompt := &survey.MultiSelect{
			Message: "Which methods do you want to implement?\n  [Press enter without selecting any of the options to skip]\n",
			Options: repo_utils.GenerateSortedRepoMethodNames(repoData.ModelData.ModelEntity),
		}
		err = survey.AskOne(selectMethodsPrompt, &repoData.SelectedRepoMethodsToImplement)
		if err != nil {
			utils.HandleError(err)
		}
	}

	// create model
	if repoData.ModelStrategy == repo_utils.ModelStrategyNewModel {
		err = model.CreateModel(repoData.ModelData)
		if err != nil {
			utils.HandleError(err, "Unable to create model")
		}
	}

	// create central repo
	if !repoData.CentralRepoExists {
		central_repo.GenerateCentralRepo()
	}

	// add repo to central repo
	if !utils.FileExists(repoData.RepoFilePath) {
		err = repo_utils.AddNewRepoToCentralRepo(repoData)
		if err != nil {
			utils.HandleError(err, "Unable to add new repo to central repo")
		}
	}

	// create repo
	err = repo_utils.CreateRepo(repoData)
	if err != nil {
		utils.HandleError(err, "Unable to create repo")
	}

	if len(repoData.SelectedRepoMethodsToImplement) > 0 {
		rawMethodsMap := repo_utils.GenerateRepoMethodNamesMap(repoData.ModelData.ModelEntity)
		selectedRawMethods := lo.Map(repoData.SelectedRepoMethodsToImplement, func(item string, index int) repo_utils.Method {
			return rawMethodsMap[item]
		})

		err = repo_utils.AddMethodsToRepo(repoData, selectedRawMethods)
		if err != nil {
			utils.HandleError(err, "Unable to add methods to repo")
		}
	}

	var toInjectRepoToService bool
	injectPrompt := &survey.Confirm{
		Message: "Do you wish to inject this repo to a service ?",
		Default: true,
	}
	err = survey.AskOne(injectPrompt, &toInjectRepoToService)
	if err != nil {
		utils.HandleError(err)
	}

	if toInjectRepoToService {

		//todo list existing services
		existingServices, err := controller_utils.ListExistingServices()
		if err != nil {
			utils.HandleError(err, "Unable to list existing services")
		}
		existingServicesMap := make(map[string]*service_utils.ServiceData)
		for _, existingService := range existingServices {
			serviceFileName := utils.PascalToSnake(existingService.ServiceFullName) + ".go"
			serviceFilePath := path.Join(cli_config.CliConfig.ServicesFolderPath, serviceFileName)
			serviceEntity := strings.TrimSuffix(existingService.ServiceFullName, "Service")

			service := service_utils.ServiceData{
				ServiceEntity:   serviceEntity,
				ServiceFileName: serviceFileName,
				ServiceFilePath: serviceFilePath,
				ServiceFullName: existingService.ServiceFullName,
				RepoData:        []repo_utils.RepoData{*repoData},
			}

			existingServicesMap[service.ServiceFullName] = &service
		}

		var selectedService string
		err = survey.AskOne(&survey.Select{
			Message: "Select a service to inject repo to:",
			Options: utils.Keys(existingServicesMap),
		}, &selectedService)
		if err != nil {
			utils.HandleError(err)
		}

		fmt.Println(existingServicesMap[selectedService].ServiceFullName)

		err = service_utils.AddRepoToService(existingServicesMap[selectedService])
		if err != nil {
			utils.HandleError(err)
		}

		if len(repoData.SelectedRepoMethodsToImplement) > 0 {
			selectMethodsToImplementPrompt := &survey.MultiSelect{
				Message: "Which methods do you want to implement?\n  [Press enter without selecting any of the options to skip]\n",
				Options: repoData.SelectedRepoMethodsToImplement,
			}
			err = survey.AskOne(selectMethodsToImplementPrompt, &existingServicesMap[selectedService].SelectedServiceProxyMethodToImplement)
			if err != nil {
				utils.HandleError(err)
			}

			err = service_utils.CopyRepoMethodsToService(existingServicesMap[selectedService], existingServicesMap[selectedService].SelectedServiceProxyMethodToImplement)
			if err != nil {
				utils.HandleError(err)
			}
		}
	}

	fmt.Println(fmt.Sprintf("✅ %s repository generated successfully.", repoData.RepoEntity))
	return
}
