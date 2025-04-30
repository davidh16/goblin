package controller

import (
	"errors"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"goblin/cli_config"
	"goblin/utils"
	"os"
	"path"
	"strings"
	"text/template"
)

var CentralControllerCmd = &cobra.Command{
	Use:   "central-controller",
	Short: "Generate a central controller",
	Run: func(cmd *cobra.Command, args []string) {
		generateCentralController()
	},
}

func generateCentralController() {

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

	funcMap := template.FuncMap{
		"GetProjectName": utils.GetProjectName,
	}

	tmpl, err := template.New(CentralControllerTemplateName).Funcs(funcMap).ParseFiles(CentralControllerTemplatePath)
	if err != nil {
		utils.HandleError(err)
	}

	f, err := os.Create(centralControllerPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err = os.Mkdir(cli_config.CliConfig.ControllersFolderPath, 0755) // 0755 = rwxr-xr-x
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
		Package        string
		ServicePackage string
	}{
		Package:        strings.Split(cli_config.CliConfig.ControllersFolderPath, "/")[len(strings.Split(cli_config.CliConfig.ControllersFolderPath, "/"))-1],
		ServicePackage: cli_config.CliConfig.ServicesFolderPath,
	}

	err = tmpl.Execute(f, templateData)
	if err != nil {
		fmt.Println("Template exec error:", err)
		return
	}

	fmt.Println("✅ Central controller generated successfully.")
}
