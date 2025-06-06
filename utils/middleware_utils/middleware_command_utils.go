package middleware_utils

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"goblin/cli_config"
	"goblin/utils"
	"goblin/utils/logger_utils"
	"os"
	"path"
	"strings"
	"text/template"
)

var MiddlewareOptions = []string{"RecoverMiddleware", "LoggingMiddleware", "JwtMiddleware", "LoggingMiddleware", "RateLimiterMiddleware", "AllowOriginMiddleware"}

var MiddlewareOptionTemplatePathMap = map[string]string{
	"LoggingMiddleware":     LoggingMiddlewareTemplatePath,
	"JwtMiddleware":         JwtMiddlewareTemplatePath,
	"AllowOriginMiddleware": AllowOriginMiddlewareTemplatePath,
	"RateLimiterMiddleware": RateLimiterMiddlewareTemplatePath,
}

var MiddlewareOptionTemplateFileNameMap = map[string]string{
	"LoggingMiddleware":     "logging_middleware.go",
	"JwtMiddleware":         "jwt_middleware.go",
	"AllowOriginMiddleware": "allow_origin_middleware.go",
	"RateLimiterMiddleware": "rate_limiter_middleware.go",
}

func GenerateMiddlewares(middlewareOptions []string) error {
	if !utils.FileExists(cli_config.CliConfig.MiddlewaresFolderPath) {
		err := os.MkdirAll(cli_config.CliConfig.MiddlewaresFolderPath, os.ModePerm)
		if err != nil {
			return err
		}
	}

	templateData := struct {
		MiddlewaresPackage  string
		LoggerPackage       string
		LoggerPackageImport string
	}{
		MiddlewaresPackage:  strings.Split(cli_config.CliConfig.MiddlewaresFolderPath, "/")[len(strings.Split(cli_config.CliConfig.MiddlewaresFolderPath, "/"))-1],
		LoggerPackage:       strings.Split(cli_config.CliConfig.LoggerFolderPath, "/")[len(strings.Split(cli_config.CliConfig.LoggerFolderPath, "/"))-1],
		LoggerPackageImport: path.Join(cli_config.CliConfig.ProjectName, cli_config.CliConfig.LoggerFolderPath),
	}

	for _, option := range middlewareOptions {
		if _, exists := MiddlewareOptionTemplatePathMap[option]; !exists {
			continue
		}

		if exists := utils.FileExists(path.Join(cli_config.CliConfig.MiddlewaresFolderPath, MiddlewareOptionTemplateFileNameMap[option])); exists {
			var overwrite bool
			injectPrompt := &survey.Confirm{
				Message: fmt.Sprintf("%s already exists, do you wish to overwrite it ?", option),
				Default: false,
			}
			err := survey.AskOne(injectPrompt, &overwrite)
			if err != nil {
				return err
			}
			if !overwrite {
				continue
			}
		}

		if option == "LoggingMiddleware" {
			err := logger_utils.GenerateLogger()
			if err != nil {
				return err
			}
		}

		tmpl, err := template.ParseFiles(MiddlewareOptionTemplatePathMap[option])
		if err != nil {
			return err
		}

		f, err := os.Create(path.Join(cli_config.CliConfig.MiddlewaresFolderPath, MiddlewareOptionTemplateFileNameMap[option]))
		if err != nil {
			return err
		}

		err = tmpl.Execute(f, templateData)
		if err != nil {
			return err
		}
	}
	return nil
}
