package router

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/davidh16/goblin/cli_config"
	"github.com/davidh16/goblin/utils"
	"github.com/davidh16/goblin/utils/middleware_utils"
	"github.com/davidh16/goblin/utils/router_utils"
	"github.com/spf13/cobra"
	"path"
)

var RouterCmd = &cobra.Command{
	Use:   "router",
	Short: "Generate a router",
	Run: func(cmd *cobra.Command, args []string) {
		routerCmdHandler()
	},
}

func routerCmdHandler() {

	if exists := utils.FileExists(path.Join(cli_config.CliConfig.RouterFolderPath, "router.go")); exists {
		var overwrite bool
		injectPrompt := &survey.Confirm{
			Message: "Router already exists, do you wish to overwrite it ?",
			Default: false,
		}
		err := survey.AskOne(injectPrompt, &overwrite)
		if err != nil {
			utils.HandleError(err)
		}
		if !overwrite {
			return
		}
	}

	routerData := router_utils.NewRouterData()

	var selectedMiddlewares []string
	selectMiddlewaresPrompt := &survey.MultiSelect{
		Message: "Which middlewares do you want to inject into your router?\n  [Press enter without selecting any of the options to skip]\n",
		Options: middleware_utils.MiddlewareOptions,
	}
	err := survey.AskOne(selectMiddlewaresPrompt, &selectedMiddlewares)
	if err != nil {
		utils.HandleError(err)
	}

	if len(selectedMiddlewares) > 0 {
		routerData.ImplementMiddlewares = true
		err = middleware_utils.GenerateMiddlewares(selectedMiddlewares)
		if err != nil {
			utils.HandleError(err)
		}
	}
	for _, m := range selectedMiddlewares {
		switch m {
		case "RecoverMiddleware":
			routerData.RecoverMiddleware = true
		case "LoggingMiddleware":
			routerData.LoggingMiddleware = true
		case "RateLimiterMiddleware":
			routerData.RateLimiterMiddleware = true
		case "AllowOriginMiddleware":
			routerData.AllowOriginMiddleware = true
		}
	}

	err = router_utils.GenerateRouter(routerData)
	if err != nil {
		utils.HandleError(err)
	}

	fmt.Println("✅ Router generated successfully.")
}
