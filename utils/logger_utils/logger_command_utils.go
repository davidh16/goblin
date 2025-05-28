package logger_utils

import (
	"goblin/cli_config"
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
