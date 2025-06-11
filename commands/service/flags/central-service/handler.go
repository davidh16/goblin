package central_service

import (
	"errors"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/davidh16/goblin/cli_config"
	central_repo "github.com/davidh16/goblin/commands/repo/flags/central-repo"
	"github.com/davidh16/goblin/templates"
	"github.com/davidh16/goblin/utils"
	"github.com/davidh16/goblin/utils/service_utils"
	"github.com/spf13/cobra"
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
	}

	centralRepoExists := utils.FileExists(path.Join(cli_config.CliConfig.RepositoriesFolderPath, "central_repo.go"))
	if !centralRepoExists {
		var confirm bool
		confirmPrompt := &survey.Confirm{
			Message: "Do you wish to inject a central repo in your service ?",
			Default: true,
		}
		if err := survey.AskOne(confirmPrompt, &confirm); err != nil {
			utils.HandleError(err)
		}

		if confirm {
			central_repo.GenerateCentralRepo()
			centralRepoExists = true
		}
	}

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
		CentralRepoExists bool
		RepoPackage       string
		RepoPackageImport string
		ServicePackage    string
	}{
		RepoPackage:       strings.Split(cli_config.CliConfig.RepositoriesFolderPath, "/")[len(strings.Split(cli_config.CliConfig.RepositoriesFolderPath, "/"))-1],
		RepoPackageImport: cli_config.CliConfig.RepositoriesFolderPath,
		ServicePackage:    strings.Split(cli_config.CliConfig.ServicesFolderPath, "/")[len(strings.Split(cli_config.CliConfig.ServicesFolderPath, "/"))-1],
		CentralRepoExists: centralRepoExists,
	}

	err = tmpl.Execute(f, templateData)
	if err != nil {
		fmt.Println("Template exec error:", err)
		return
	}

	if !alreadyExists {
		if centralControllerExists := utils.FileExists(path.Join(cli_config.CliConfig.ControllersFolderPath, "central_controller.go")); centralControllerExists {
			err = service_utils.AddCentralServiceToCentralControllerConstructor()
			if err != nil {
				utils.HandleError(err)
			}
		}
	}

	fmt.Println("✅ Central service generated successfully.")
}
