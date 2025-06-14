package logger_utils

import (
	"github.com/davidh16/goblin/cli_config"
	"github.com/davidh16/goblin/templates"
	"github.com/davidh16/goblin/utils"
	"os"
	"path"
	"strings"
	"text/template"
)

const (
	LoggerTemplatePath = "logger.tmpl"
	LoggerFileName     = "logger.go"
)

func GenerateLogger() error {

	if exists := utils.FileExists(cli_config.CliConfig.LoggerFolderPath); !exists {
		err := os.MkdirAll(cli_config.CliConfig.LoggerFolderPath, os.ModePerm)
		if err != nil {
			return err
		}
	}

	tmpl, err := template.ParseFS(templates.Files, LoggerTemplatePath)
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
