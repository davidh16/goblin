package model

import (
	"errors"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/davidh16/goblin/cli_config"
	"github.com/davidh16/goblin/commands/model/flags/user"
	"github.com/davidh16/goblin/utils"
	"github.com/davidh16/goblin/utils/migration_utils"
	"github.com/davidh16/goblin/utils/model_utils"
	"github.com/jinzhu/inflection"
	"github.com/spf13/cobra"
	"os"
	"path"
	"strings"
	"text/template"
)

var (
	UserModelFlag bool
)

var ModelCmd = &cobra.Command{
	Use:   "model",
	Short: "Generate model",
	Run: func(cmd *cobra.Command, args []string) {
		if UserModelFlag {
			user.GenerateUserModel()
		} else {
			ModelCmdHandler()
		}
	},
}

func ModelCmdHandler() {

	modelData := model_utils.ModelData{}
	defaultModelName := "my_model_file"
	for {
		if err := survey.AskOne(&survey.Input{
			Message: "Please type the model file name (snake_case) :",
			Default: defaultModelName,
		}, &modelData.NameSnakeCase); err != nil {
			utils.HandleError(err)
		}

		if !utils.IsSnakeCase(modelData.NameSnakeCase) {
			fmt.Printf("🛑 %s is not in snake case\n", modelData.NameSnakeCase)
			continue
		}

		defaultModelName = modelData.NameSnakeCase

		var confirmContinue bool
		confirmPrompt := &survey.Confirm{
			Message: fmt.Sprintf("You are about to create a model file named %s.go, do you want to continue ?", modelData.NameSnakeCase),
		}
		if err := survey.AskOne(confirmPrompt, &confirmContinue); err != nil {
			utils.HandleError(err)
		}

		if !confirmContinue {
			continue
		}

		modelData.ModelEntity = utils.SnakeToPascal(modelData.NameSnakeCase)
		modelData.ModelFileName = modelData.NameSnakeCase + ".go"
		modelData.ModelFilePath = path.Join(cli_config.CliConfig.ModelsFolderPath, modelData.ModelFileName)

		if utils.FileExists(modelData.ModelFilePath) {
			var overwriteConfirmed bool
			confirmPrompt = &survey.Confirm{
				Message: fmt.Sprintf("%s model already exists. Do you want to overwrite it ?", modelData.ModelFileName),
				Default: false,
			}
			if err := survey.AskOne(confirmPrompt, &overwriteConfirmed); err != nil {
				utils.HandleError(err)
			}

			if overwriteConfirmed {
				confirmPrompt = &survey.Confirm{
					Message: fmt.Sprintf("Are you sure you want to overwrite %s model ?", modelData.ModelFileName),
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

	funcMap := template.FuncMap{
		"Pluralize": inflection.Plural,
	}

	// Write the model to file
	tmpl, err := template.New(model_utils.ModelTemplateName).Funcs(funcMap).ParseFiles(model_utils.ModelTemplatePath)
	if err != nil {
		utils.HandleError(err, "Error parsing model template")
	}

	f, err := os.Create(modelData.ModelFilePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err = os.Mkdir(cli_config.CliConfig.ModelsFolderPath, 0755) // 0755 = rwxr-xr-x
			if err != nil {
				utils.HandleError(err, "Error creating model folder")
			}
			f, err = os.Create(modelData.ModelFilePath)
			if err != nil {
				utils.HandleError(err, "Error creating model file")
			}
		}
	}
	defer f.Close()

	templateData := struct {
		Package         string
		ModelPascalCase string
		ModelCamelCase  string
	}{
		Package:         strings.Split(cli_config.CliConfig.ModelsFolderPath, "/")[len(strings.Split(cli_config.CliConfig.ModelsFolderPath, "/"))-1],
		ModelPascalCase: modelData.ModelEntity,
		ModelCamelCase:  utils.PascalToCamel(modelData.ModelEntity),
	}

	err = tmpl.Execute(f, templateData)
	if err != nil {
		utils.HandleError(err, "Error executing model template")
	}

	var createMigration bool
	confirmPrompt := &survey.Confirm{
		Message: "Do you want to create a migration for your model ?",
		Default: true,
	}
	if err = survey.AskOne(confirmPrompt, &createMigration); err != nil {
		utils.HandleError(err)
	}

	if createMigration {

		migrationData := migration_utils.GenerateMigrationDataFromName(inflection.Plural(modelData.NameSnakeCase))

		err = migration_utils.GenerateMigrationFiles(migrationData)
		if err != nil {
			utils.HandleError(err)
		}
	}

	fmt.Println(fmt.Sprintf("✅ %s model generated.", modelData.ModelEntity))
}

func CreateModel(modelData *model_utils.ModelData) error {
	funcMap := template.FuncMap{
		"Pluralize": inflection.Plural,
	}

	// Write the model to file
	tmpl, err := template.New(model_utils.ModelTemplateName).Funcs(funcMap).ParseFiles(model_utils.ModelTemplatePath)
	if err != nil {
		return err
	}

	f, err := os.Create(modelData.ModelFilePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err = os.Mkdir(cli_config.CliConfig.ModelsFolderPath, 0755) // 0755 = rwxr-xr-x
			if err != nil {
				if !os.IsExist(err) {
					return err
				}
			}
			f, err = os.Create(modelData.ModelFilePath)
			if err != nil {
				return err
			}
		}
	}
	defer f.Close()

	templateData := struct {
		Package         string
		ModelPascalCase string
		ModelCamelCase  string
	}{
		Package:         strings.Split(cli_config.CliConfig.ModelsFolderPath, "/")[len(strings.Split(cli_config.CliConfig.ModelsFolderPath, "/"))-1],
		ModelPascalCase: modelData.ModelEntity,
		ModelCamelCase:  utils.PascalToCamel(modelData.ModelEntity),
	}

	err = tmpl.Execute(f, templateData)
	if err != nil {
		return err
	}

	if modelData.CreateMigration {
		migrationData := migration_utils.GenerateMigrationDataFromName(inflection.Plural(modelData.NameSnakeCase))

		err = migration_utils.GenerateMigrationFiles(migrationData)
		if err != nil {
			utils.HandleError(err)
		}
	}

	fmt.Println(fmt.Sprintf("✅ %s model generated.", modelData.ModelEntity))
	return nil
}
