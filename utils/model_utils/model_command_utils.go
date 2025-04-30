package model_utils

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"goblin/cli_config"
	"goblin/utils"
	"path"
)

type ModelData struct {
	NameSnakeCase string
	ModelFileName string
	ModelFilePath string
	ModelEntity   string
}

func NewModelData() *ModelData {
	return &ModelData{}
}

func TriggerGetModelNameFlow() (*ModelData, error) {
	modelData := &ModelData{}

	for {
		for {
			if err := survey.AskOne(&survey.Input{
				Message: "Please type the model file name (snake_case) :",
				Default: "my_model_file",
			}, &modelData.NameSnakeCase); err != nil {
				return nil, err
			}

			if !utils.IsSnakeCase(modelData.NameSnakeCase) {
				fmt.Printf("🛑 %s is not in snake case\n", modelData.NameSnakeCase)
			} else {
				break
			}
		}

		modelData.ModelEntity = utils.SnakeToPascal(modelData.NameSnakeCase)
		modelData.ModelFileName = modelData.NameSnakeCase + ".go"
		modelData.ModelFilePath = path.Join(cli_config.CliConfig.ModelsFolderPath, modelData.ModelFileName)

		var confirm bool
		confirmPrompt := &survey.Confirm{
			Message: fmt.Sprintf("You are about to create a model file named %s, do you want to continue ?", modelData.ModelFileName),
		}
		if err := survey.AskOne(confirmPrompt, &confirm); err != nil {
			return nil, err
		}

		if !confirm {
			continue
		}

		for {
			if utils.FileExists(modelData.ModelFilePath) {
				var overwriteConfirmed bool
				confirmPrompt = &survey.Confirm{
					Message: fmt.Sprintf("%s model already exists. Do you want to overwrite it ?", modelData.ModelFileName),
					Default: false,
				}
				if err := survey.AskOne(confirmPrompt, &overwriteConfirmed); err != nil {
					return nil, err
				}

				if overwriteConfirmed {
					confirmPrompt = &survey.Confirm{
						Message: fmt.Sprintf("Are you sure you want to overwrite %s model ?", modelData.ModelFileName),
						Default: false,
					}
					if err := survey.AskOne(confirmPrompt, &overwriteConfirmed); err != nil {
						return nil, err
					}
				}

				if !overwriteConfirmed {
					continue
				}
			}
			break
		}
		break
	}
	return modelData, nil
}
