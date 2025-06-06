package middleware

import "github.com/spf13/cobra"

var MiddlewareCmd = &cobra.Command{
	Use:   "middleware",
	Short: "Generate a middleware",
	Run: func(cmd *cobra.Command, args []string) {
		middlewareCmdHandler()
	},
}

func middlewareCmdHandler() {
	// list existing middlewares

	// choose which ones to implement from ones that don't exist

	// generate
}
