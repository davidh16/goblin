package edit

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"goblin/cli_config"
	"goblin/utils"
)

var EditConfigCmd = &cobra.Command{
	Use:   "edit-goblin-config",
	Short: "Edit goblin goblin config file content",
	Run: func(cmd *cobra.Command, args []string) {
		EditConfigCmdHandler()
	},
}

func EditConfigCmdHandler() {
	configMap, err := cli_config.LoadConfigAsMap()
	if err != nil {
		utils.HandleError(err, "Failed loading config file")
	}

	var selectedKey string
	prompt := &survey.Select{
		Message: "Which goblin config value do you want to edit?",
		Options: utils.MapToString(configMap),
	}
	if err = survey.AskOne(prompt, &selectedKey); err != nil {
		utils.HandleError(err)
	}

	selectedKey = utils.ExtractKeyFromLine(selectedKey)

	oldValue := configMap[selectedKey]
	currentValue := fmt.Sprintf("%v", oldValue)

	var newValueRaw string

	if err = survey.AskOne(&survey.Input{
		Message: fmt.Sprintf("Edit value for '%s':", selectedKey),
		Default: fmt.Sprintf("%v", currentValue), // 👈 pre-filled!
	}, &newValueRaw); err != nil {
		utils.HandleError(err)
	}

	preview := fmt.Sprintf("Preview:\n%s: %v -> %v", selectedKey, oldValue, newValueRaw)
	fmt.Println(preview)

	var confirm bool
	confirmPrompt := &survey.Confirm{
		Message: "Are you sure you want to save this change?",
		Default: true,
	}
	if err = survey.AskOne(confirmPrompt, &confirm); err != nil {
		utils.HandleError(err)
	}

	if !confirm {
		fmt.Println("Aborted. No changes saved.")
	}

	configMap[selectedKey] = newValueRaw

	err = cli_config.UpdateConfigFile(configMap)
	if err != nil {
		utils.HandleError(err, "Failed updating goblin config file")
	}

	fmt.Println("✅ Config updated successfully.")
	return
}
