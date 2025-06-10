package controller

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"goblin/cli_config"
	central_controller "goblin/commands/controller/flags/central-controller"
	"goblin/utils"
	"goblin/utils/controller_utils"
	"goblin/utils/service_utils"
	"path"
)

var CentralControllerFlag bool

var ControllerCmd = &cobra.Command{
	Use:   "controller",
	Short: "Generate a controller",
	Run: func(cmd *cobra.Command, args []string) {
		if CentralControllerFlag {
			central_controller.CentralControllerFlagHandler()
		} else {
			controllerCmdHandler()
		}
	},
}

func controllerCmdHandler() {
	controllerData := controller_utils.NewControllerData()
	defaultControllerName := "my_controller_file"
	for {
		if err := survey.AskOne(&survey.Input{
			Message: "Please type the controllers file name (snake_case) :",
			Default: defaultControllerName,
		}, &controllerData.ControllerNameSnakeCase); err != nil {
			utils.HandleError(err)
		}

		if !utils.IsSnakeCase(controllerData.ControllerNameSnakeCase) {
			fmt.Printf("🛑 %s is not in snake case\n", controllerData.ControllerNameSnakeCase)
			continue
		}

		defaultControllerName = controllerData.ControllerNameSnakeCase

		var confirm bool
		confirmPrompt := &survey.Confirm{
			Message: fmt.Sprintf("You are about to create a controller file named %s_controller.go, do you want to continue ?", controllerData.ControllerNameSnakeCase),
		}
		if err := survey.AskOne(confirmPrompt, &confirm); err != nil {
			utils.HandleError(err)
		}

		if !confirm {
			continue
		}

		controllerData.ControllerEntity = utils.SnakeToPascal(controllerData.ControllerNameSnakeCase)
		controllerData.ControllerFullName = controllerData.ControllerEntity + "Controller"
		controllerData.ControllerFileName = controllerData.ControllerNameSnakeCase + "_controller.go"
		controllerData.ControllerFilePath = path.Join(cli_config.CliConfig.ControllersFolderPath, controllerData.ControllerFileName)

		if utils.FileExists(controllerData.ControllerFilePath) {
			var confirmOverwrite bool
			confirmPrompt = &survey.Confirm{
				Message: fmt.Sprintf("%s controller already exists. Do you want to overwrite it ?", controllerData.ControllerFileName),
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
					Message: fmt.Sprintf("Are you sure you want to overwrite %s controller ?", controllerData.ControllerFileName),
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

	controllerData.CentralControllerExists = utils.FileExists(path.Join(cli_config.CliConfig.ControllersFolderPath, "central_controller.go"))

	existingServices, err := controller_utils.ListExistingServices()
	if err != nil {
		utils.HandleError(err, "Unable to list existing services")
	}
	existingServicesMap := make(map[string]*service_utils.ServiceData)
	for _, existingService := range existingServices {
		existingServicesMap[existingService.ServiceFullName] = &existingService
	}

	// controller_utils.ServiceOptionsStrategyMap keys are used to list options for choosing service strategy, if there are no existing services, that key (option) has to be removed from the map
	if len(existingServices) == 0 {
		delete(controller_utils.ServiceOptionsStrategyMap, controller_utils.ServiceStrategyOptionsMap[controller_utils.ServiceStrategyExistingService])
	}

	var serviceStrategyChosenOption string
	err = survey.AskOne(&survey.Select{
		Message: "Choose service strategy:",
		Options: utils.Keys(controller_utils.ServiceOptionsStrategyMap),
	}, &serviceStrategyChosenOption)
	if err != nil {
		utils.HandleError(err)
	}

	controllerData.ServiceStrategy = controller_utils.ServiceOptionsStrategyMap[serviceStrategyChosenOption]

	if controllerData.ServiceStrategy == controller_utils.ServiceStrategyExistingService {
		var chosenServices []string
		err = survey.AskOne(&survey.MultiSelect{
			Message: "Select a service to use:",
			Options: utils.Keys(existingServicesMap),
		}, &chosenServices)
		if err != nil {
			utils.HandleError(err)
		}

		for _, service := range chosenServices {
			controllerData.ServiceData = append(controllerData.ServiceData, *existingServicesMap[service])
		}
	}

	if controllerData.ServiceStrategy == controller_utils.ServiceStrategyNewService {
		service, err := controller_utils.PrepareService()
		if err != nil {
			utils.HandleError(err)
		}
		controllerData.ServiceData = append(controllerData.ServiceData, *service)
	}

	if controllerData.ServiceStrategy == controller_utils.ServiceStrategyNewService {
		err = controller_utils.ExecuteCreateService(controllerData.ServiceData)
		if err != nil {
			utils.HandleError(err)
		}
	}

	if !controllerData.CentralControllerExists {
		err = central_controller.GenerateCentralController()
		if err != nil {
			utils.HandleError(err)
		}
	}

	if !utils.FileExists(controllerData.ControllerFilePath) {
		err = controller_utils.AddNewControllerToCentralController(controllerData)
		if err != nil {
			utils.HandleError(err)
		}
	}

	err = controller_utils.CreateController(controllerData)
	if err != nil {
		utils.HandleError(err)
	}

	if controllerData.ServiceStrategy != controller_utils.ServiceStrategyNoImplementation {
		err = controller_utils.AddServiceToController(controllerData)
		if err != nil {
			utils.HandleError(err)
		}
	}

	fmt.Println(fmt.Sprintf("✅ %s controller generated successfully.", controllerData.ControllerEntity))
	return
}
