package router

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"goblin/utils"
	"goblin/utils/middleware_utils"
	"goblin/utils/router_utils"
)

var RouterCmd = &cobra.Command{
	Use:   "router",
	Short: "Generate a router",
	Run: func(cmd *cobra.Command, args []string) {
		routerCmdHandler()
	},
}

func routerCmdHandler() {

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

	fmt.Println("✅ router generated successfully.")
}
