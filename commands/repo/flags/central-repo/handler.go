package central_repo

import (
	"errors"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/davidh16/goblin/cli_config"
	"github.com/davidh16/goblin/utils"
	"github.com/davidh16/goblin/utils/repo_utils"
	"github.com/spf13/cobra"
	"os"
	"path"
	"strings"
	"text/template"
)

var CentralRepoCmd = &cobra.Command{
	Use:   "central-repo",
	Short: "Generate a central repository",
	Run: func(cmd *cobra.Command, args []string) {
		GenerateCentralRepo()
	},
}

const (
	CentralRepoTemplatePath    = "commands/repo/flags/central-repo/central_repo.tmpl"
	UnitOfWorkRepoTemplatePath = "commands/repo/flags/central-repo/unit_of_work.tmpl"
)

func GenerateCentralRepo() {

	centralRepoPath := path.Join(cli_config.CliConfig.RepositoriesFolderPath, "central_repo.go")

	alreadyExists := utils.FileExists(centralRepoPath)
	if alreadyExists {
		var confirm bool
		confirmPrompt := &survey.Confirm{
			Message: "Central repository already exists. Do you want to overwrite it ?",
			Default: false,
		}
		if err := survey.AskOne(confirmPrompt, &confirm); err != nil {
			utils.HandleError(err)
		}
		if confirm {
			confirmPrompt = &survey.Confirm{
				Message: "Are you sure you want to overwrite central repository ?",
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

	unitOFWorkRepoPath := path.Join(cli_config.CliConfig.RepositoriesFolderPath, "unit_of_work.go")
	alreadyExists = utils.FileExists(unitOFWorkRepoPath)
	if alreadyExists {
		var confirm bool
		confirmPrompt := &survey.Confirm{
			Message: "Unit of work repository util already exists. Do you want to overwrite it ?",
			Default: false,
		}
		if err := survey.AskOne(confirmPrompt, &confirm); err != nil {
			utils.HandleError(err)
		}
		if confirm {
			confirmPrompt = &survey.Confirm{
				Message: "Are you sure you want to overwrite unit of work repository util ?",
				Default: false,
			}
			if err := survey.AskOne(confirmPrompt, &confirm); err != nil {
				utils.HandleError(err)
			}

			if !confirm {
				return
			}
			generateUnitOfWorkRepoUtil(unitOFWorkRepoPath)
		}
	} else {
		generateUnitOfWorkRepoUtil(unitOFWorkRepoPath)
	}

	tmpl, err := template.ParseFiles(CentralRepoTemplatePath)
	if err != nil {
		utils.HandleError(err)
	}

	f, err := os.Create(centralRepoPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err = os.Mkdir(cli_config.CliConfig.RepositoriesFolderPath, 0755) // 0755 = rwxr-xr-x
			if err != nil {
				fmt.Println("Error creating folder:", err)
			}
			f, err = os.Create(centralRepoPath)
			if err != nil {
				fmt.Println("File creation error:", err)
				return
			}
		}
	}
	defer f.Close()

	templateData := struct {
		RepoPackage string
	}{
		RepoPackage: strings.Split(cli_config.CliConfig.RepositoriesFolderPath, "/")[len(strings.Split(cli_config.CliConfig.RepositoriesFolderPath, "/"))-1],
	}

	err = tmpl.Execute(f, templateData)
	if err != nil {
		fmt.Println("Template exec error:", err)
		return
	}

	if !alreadyExists {
		if centralServiceExists := utils.FileExists(path.Join(cli_config.CliConfig.ServicesFolderPath, "central_service.go")); centralServiceExists {
			err = repo_utils.AddCentralRepoToCentralServiceConstructor()
			if err != nil {
				utils.HandleError(err)
			}
		}
	}

	fmt.Println("✅ Central repository generated successfully.")
}

func generateUnitOfWorkRepoUtil(unitOFWorkRepoPath string) {
	tmpl, err := template.ParseFiles(UnitOfWorkRepoTemplatePath)
	if err != nil {
		utils.HandleError(err)
	}

	f, err := os.Create(unitOFWorkRepoPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err = os.Mkdir(cli_config.CliConfig.RepositoriesFolderPath, 0755) // 0755 = rwxr-xr-x
			if err != nil {
				fmt.Println("Error creating folder:", err)
			}
			f, err = os.Create(unitOFWorkRepoPath)
			if err != nil {
				fmt.Println("File creation error:", err)
				return
			}
		}
	}
	defer f.Close()

	err = tmpl.Execute(f, struct {
		RepoPackage string
	}{strings.Split(cli_config.CliConfig.RepositoriesFolderPath, "/")[len(strings.Split(cli_config.CliConfig.RepositoriesFolderPath, "/"))-1]})
	if err != nil {
		fmt.Println("Template exec error:", err)
		return
	}
}
