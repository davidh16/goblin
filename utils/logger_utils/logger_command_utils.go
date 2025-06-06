package logger_utils

import (
	"goblin/cli_config"
	"goblin/utils"
	"os"
	"path"
	"strings"
	"text/template"
)

const (
	LoggerTemplatePath = "commands/logger/logger.tmpl"
	LoggerFileName     = "logger.go"
)

func GenerateLogger() error {

	if exists := utils.FileExists(cli_config.CliConfig.LoggerFolderPath); !exists {
		err := os.Mkdir(cli_config.CliConfig.LoggerFolderPath, os.ModePerm)
		if err != nil {
			return err
		}
	}

	tmpl, err := template.ParseFiles(LoggerTemplatePath)
	if err != nil {
		return err
	}

	f, err := os.Create(path.Join(cli_config.CliConfig.LoggerFolderPath, LoggerFileName))
	if err != nil {
		return err
	}
	defer f.Close()

	templateData := struct {
		LoggerPackage string
	}{
		LoggerPackage: strings.Split(cli_config.CliConfig.LoggerFolderPath, "/")[len(strings.Split(cli_config.CliConfig.LoggerFolderPath, "/"))-1],
	}

	err = tmpl.Execute(f, templateData)
	if err != nil {
		return err
	}
	return nil
}
