/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"goblin/cli_config"
	"goblin/root_cmd"
	"goblin/utils"
)

func init() {
	err := cli_config.LoadConfig()
	if err != nil {
		utils.HandleError(err)
	}
}

func main() {
	root_cmd.Execute()
}
