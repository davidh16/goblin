package central_controller

import (
	"errors"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/davidh16/goblin/cli_config"
	central_service "github.com/davidh16/goblin/commands/service/flags/central-service"
	"github.com/davidh16/goblin/utils"
	"github.com/davidh16/goblin/utils/controller_utils"
	"os"
	"path"
	"strings"
	"text/template"
)

func CentralControllerFlagHandler() {

	centralControllerPath := path.Join(cli_config.CliConfig.ControllersFolderPath, "central_controller.go")

	alreadyExists := utils.FileExists(centralControllerPath)
	if alreadyExists {
		var confirm bool
		confirmPrompt := &survey.Confirm{
			Message: "Central controller already exists. Do you want to overwrite it ?",
			Default: false,
		}
		if err := survey.AskOne(confirmPrompt, &confirm); err != nil {
			utils.HandleError(err)
		}
		if confirm {
			confirmPrompt = &survey.Confirm{
				Message: "Are you sure you want to overwrite central controller ?",
				Default: false,
			}
			if err := survey.AskOne(confirmPrompt, &confirm); err != nil {
				utils.HandleError(err)
			}

			if !confirm {
				return
			}
		}

		return
	}

	centralServiceExists := utils.FileExists(path.Join(cli_config.CliConfig.ServicesFolderPath, "central_service.go"))
	if !centralServiceExists {
		var confirm bool
		confirmPrompt := &survey.Confirm{
			Message: "Do you wish to inject a central service in your controller ?",
			Default: true,
		}
		if err := survey.AskOne(confirmPrompt, &confirm); err != nil {
			utils.HandleError(err)
		}

		if confirm {
			central_service.GenerateCentralService()
			centralServiceExists = true
		}
	}

	funcMap := template.FuncMap{
		"GetProjectName": utils.GetProjectName,
	}

	tmpl, err := template.New(controller_utils.CentralControllerTemplateName).Funcs(funcMap).ParseFiles(controller_utils.CentralControllerTemplatePath)
	if err != nil {
		utils.HandleError(err)
	}

	f, err := os.Create(centralControllerPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err = os.MkdirAll(cli_config.CliConfig.ControllersFolderPath, 0755) // 0755 = rwxr-xr-x
			if err != nil {
				fmt.Println("Error creating folder:", err)
			}
			f, err = os.Create(centralControllerPath)
			if err != nil {
				fmt.Println("File creation error:", err)
				return
			}
		}
	}
	defer f.Close()

	templateData := struct {
		Package              string
		CentralServiceExists bool
		ServicePackage       string
		ServicePackageImport string
	}{
		Package:              strings.Split(cli_config.CliConfig.ControllersFolderPath, "/")[len(strings.Split(cli_config.CliConfig.ControllersFolderPath, "/"))-1],
		ServicePackageImport: cli_config.CliConfig.ServicesFolderPath,
		ServicePackage:       strings.Split(cli_config.CliConfig.ServicesFolderPath, "/")[len(strings.Split(cli_config.CliConfig.ServicesFolderPath, "/"))-1],
		CentralServiceExists: centralServiceExists,
	}

	err = tmpl.Execute(f, templateData)
	if err != nil {
		fmt.Println("Template exec error:", err)
		return
	}

	fmt.Println("✅ Central controller generated successfully.")
}

func GenerateCentralController() error {
	centralControllerPath := path.Join(cli_config.CliConfig.ControllersFolderPath, "central_controller.go")

	alreadyExists := utils.FileExists(centralControllerPath)
	if !alreadyExists {
		centralServiceExists := utils.FileExists(path.Join(cli_config.CliConfig.ServicesFolderPath, "central_service.go"))

		f, err := os.Create(centralControllerPath)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				err = os.MkdirAll(cli_config.CliConfig.ControllersFolderPath, 0755) // 0755 = rwxr-xr-x
				if err != nil {
					fmt.Println("Error creating folder:", err)
				}
				f, err = os.Create(centralControllerPath)
				if err != nil {
					return err
				}
			}
		}
		defer f.Close()

		funcMap := template.FuncMap{
			"GetProjectName": utils.GetProjectName,
		}

		tmpl, err := template.New(controller_utils.CentralControllerTemplateName).Funcs(funcMap).ParseFiles(controller_utils.CentralControllerTemplatePath)
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
			CentralServiceExists: centralServiceExists,
		}

		err = tmpl.Execute(f, templateData)
		if err != nil {
			return err
		}
	}

	return nil
}
