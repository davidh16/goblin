package root_cmd

import (
	"goblin/commands/config"
	"goblin/commands/controller"
	"goblin/commands/database"
	"goblin/commands/logger"
	"goblin/commands/middleware"
	"goblin/commands/migration"
	"goblin/commands/model"
	"goblin/commands/repo"
	"goblin/commands/router"
	"goblin/commands/service"
	"goblin/commands/workerize"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "goblin",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	//rootCmd.PersistentFlags().StringVar(&cfgFile, "cli_config", "", "cli_config file (default is $HOME/.goblin.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	rootCmd.AddCommand(config.ConfigCmd)
	config.ConfigCmd.Flags().BoolVarP(&config.EditCliConfigFlag, "edit", "e", false, "Edit goblin config file")

	rootCmd.AddCommand(database.DatabaseCmd)

	rootCmd.AddCommand(model.ModelCmd)
	model.ModelCmd.Flags().BoolVarP(&model.UserModelFlag, "user", "u", false, "Generate user model")

	rootCmd.AddCommand(repo.RepoCmd)
	repo.RepoCmd.Flags().BoolVarP(&repo.CentralRepoFlag, "central-repo", "c", false, "Generate central repository")

	rootCmd.AddCommand(service.ServiceCmd)
	service.ServiceCmd.Flags().BoolVarP(&service.CentralServiceFlag, "central-service", "c", false, "Generate central service")

	rootCmd.AddCommand(controller.ControllerCmd)
	controller.ControllerCmd.Flags().BoolVarP(&controller.CentralControllerFlag, "central-controller", "c", false, "Generate central controller")

	rootCmd.AddCommand(workerize.WorkerizeCmd)
	workerize.WorkerizeCmd.Flags().BoolVarP(&workerize.CustomJobFlag, "job", "j", false, "Generate custom job")

	rootCmd.AddCommand(logger.LoggerCmd)

	rootCmd.AddCommand(migration.MigrationCmd)

	rootCmd.AddCommand(router.RouterCmd)

	rootCmd.AddCommand(middleware.MiddlewareCmd)
}
