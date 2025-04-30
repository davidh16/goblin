package central_service

import (
	"errors"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"goblin/cli_config"
	"goblin/utils"
	"goblin/utils/service_utils"
	"os"
	"path"
	"strings"
	"text/template"
)

var CentralServiceCmd = &cobra.Command{
	Use:   "central-service",
	Short: "Generate a central service",
	Run: func(cmd *cobra.Command, args []string) {
		GenerateCentralService()
	},
}

func GenerateCentralService() {

	centralServicePath := path.Join(cli_config.CliConfig.ServicesFolderPath, "central_service.go")

	alreadyExists := utils.FileExists(centralServicePath)
	if alreadyExists {
		var confirm bool
		confirmPrompt := &survey.Confirm{
			Message: "Central service already exists. Do you want to overwrite it ?",
			Default: false,
		}
		if err := survey.AskOne(confirmPrompt, &confirm); err != nil {
			utils.HandleError(err)
		}
		if confirm {
			confirmPrompt = &survey.Confirm{
				Message: "Are you sure you want to overwrite central service ?",
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

	tmpl, err := template.New(service_utils.CentralServiceTemplateName).Funcs(funcMap).ParseFiles(service_utils.CentralServiceTemplatePath)
	if err != nil {
		utils.HandleError(err)
	}

	f, err := os.Create(centralServicePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err = os.Mkdir(cli_config.CliConfig.ServicesFolderPath, 0755) // 0755 = rwxr-xr-x
			if err != nil {
				fmt.Println("Error creating folder:", err)
			}
			f, err = os.Create(centralServicePath)
			if err != nil {
				fmt.Println("File creation error:", err)
				return
			}
		}
	}
	defer f.Close()

	templateData := struct {
		RepoPackage    string
		ServicePackage string
	}{
		RepoPackage:    strings.Split(cli_config.CliConfig.RepositoriesFolderPath, "/")[len(strings.Split(cli_config.CliConfig.RepositoriesFolderPath, "/"))-1],
		ServicePackage: strings.Split(cli_config.CliConfig.ServicesFolderPath, "/")[len(strings.Split(cli_config.CliConfig.ServicesFolderPath, "/"))-1],
	}

	err = tmpl.Execute(f, templateData)
	if err != nil {
		fmt.Println("Template exec error:", err)
		return
	}

	fmt.Println("✅ Central service generated successfully.")
}
