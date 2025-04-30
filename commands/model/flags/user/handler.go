package user

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

var UserCmd = &cobra.Command{
	Use:   "user",
	Short: "Generate a user model",
	Run: func(cmd *cobra.Command, args []string) {
		GenerateUserModel()
	},
}

func GenerateUserModel() {

	userModelPath := path.Join(cli_config.CliConfig.ModelsFolderPath, "user.go")

	existingModelAttributes := detectExistingModelAttributes(userModelPath, utils.Keys(optionalUserModelAttributes))
	prompt := &survey.MultiSelect{
		Message: "Select fields to include in the User model:",
		Options: utils.Keys(optionalUserModelAttributes),
		Default: existingModelAttributes,
	}

	selectedOptionalAttributes := []string{}
	err := survey.AskOne(prompt, &selectedOptionalAttributes)
	if err != nil {
		fmt.Println("Prompt failed:", err)
		return
	}

	var optionalAttributes []UserModelAttribute
	for _, attribute := range selectedOptionalAttributes {
		optionalAttributes = append(optionalAttributes, NewUserModelAttribute(attribute))
	}

	// Write the model to file
	tmpl, err := template.ParseFiles(UserModelTemplatePath)
	if err != nil {
		panic(err)
	}

	f, err := os.Create(userModelPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err = os.Mkdir(cli_config.CliConfig.ModelsFolderPath, 0755) // 0755 = rwxr-xr-x
			if err != nil {
				fmt.Println("Error creating folder:", err)
			}
			f, err = os.Create(userModelPath)
			if err != nil {
				fmt.Println("File creation error:", err)
				return
			}
		}
	}
	defer f.Close()

	err = tmpl.Execute(f, optionalAttributes)
	if err != nil {
		fmt.Println("Template exec error:", err)
		return
	}

	fmt.Println("✅ User model generated with selected fields.")
}

func detectExistingModelAttributes(filepath string, attributes []string) []string {
	existing := []string{}
	content, err := os.ReadFile(filepath)
	if err != nil {
		return existing // file doesn't exist, skip
	}

	for _, key := range attributes {
		// very simple matching based on presence in file
		if strings.Contains(string(content), key) {
			existing = append(existing, key)
		}
	}
	return existing
}
