package middleware

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"goblin/utils"
	"goblin/utils/middleware_utils"
	"os"
)

var MiddlewareCmd = &cobra.Command{
	Use:   "middleware",
	Short: "Generate a middleware",
	Run: func(cmd *cobra.Command, args []string) {
		middlewareCmdHandler()
	},
}

func middlewareCmdHandler() {
	existingMiddlewares, err := middleware_utils.ListExistingMiddlewares()
	if err != nil {
		if !os.IsNotExist(err) {
			utils.HandleError(err)
		}
	}

	fmt.Println(existingMiddlewares)

	for _, middleware := range existingMiddlewares {
		delete(middleware_utils.MiddlewareOptionTemplatePathMap, middleware)
		delete(middleware_utils.MiddlewareOptionTemplateFileNameMap, middleware)
	}

	if len(middleware_utils.MiddlewareOptionTemplatePathMap) == 0 {
		fmt.Println("All predefined middlewares have already been implemented")
		return
	}

	availableMiddlewareOptions := utils.Keys(middleware_utils.MiddlewareOptionTemplatePathMap)
	var selectedMiddlewares []string
	selectMiddlewaresPrompt := &survey.MultiSelect{
		Message: "Which middlewares do you want to implement?\n  [Press enter without selecting any of the options to skip]\n",
		Options: availableMiddlewareOptions,
	}
	err = survey.AskOne(selectMiddlewaresPrompt, &selectedMiddlewares)
	if err != nil {
		utils.HandleError(err)
	}

	// generate
	err = middleware_utils.GenerateMiddlewares(selectedMiddlewares)
	if err != nil {
		utils.HandleError(err)
	}

	fmt.Println("✅ Keep in mind that newly implemented middlewares must be injected in router manually.")

	return
}
