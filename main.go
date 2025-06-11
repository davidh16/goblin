package main

import (
	"embed"
	"github.com/davidh16/goblin/cli_config"
	"github.com/davidh16/goblin/root_cmd"
	"github.com/davidh16/goblin/utils"
)

//go:embed templates/*.tmpl
var Files embed.FS

func init() {
	err := cli_config.LoadConfig()
	if err != nil {
		utils.HandleError(err)
	}
}

func main() {
	root_cmd.Execute()
}
