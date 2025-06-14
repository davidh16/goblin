package middleware_utils

import (
	"errors"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/davidh16/goblin/cli_config"
	"github.com/davidh16/goblin/templates"
	"github.com/davidh16/goblin/utils"
	"github.com/davidh16/goblin/utils/logger_utils"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"
)

var MiddlewareOptions = []string{"RecoverMiddleware", "JwtMiddleware", "LoggingMiddleware", "RateLimiterMiddleware", "AllowOriginMiddleware"}

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

		if option == "JwtMiddleware" {
			err := generateAuthJwtFile()
			if err != nil {
				return err
			}

			workingDirectory, err := os.Getwd()
			if err != nil {
				return err
			}

			envFilePath := path.Join(workingDirectory, ".env")

			envFile, err := os.OpenFile(envFilePath, os.O_CREATE|os.O_RDWR, 0644)
			if err != nil {
				return errors.New("error opening environment file")
			}

			err = utils.WriteToEnvFile(envFile, map[string]string{
				"JWT_SECRET": "",
			})
			if err != nil {
				return err
			}
		}

		if option == "AllowOriginMiddleware" {
			workingDirectory, err := os.Getwd()
			if err != nil {
				return err
			}

			envFilePath := path.Join(workingDirectory, ".env")

			envFile, err := os.OpenFile(envFilePath, os.O_CREATE|os.O_RDWR, 0644)
			if err != nil {
				return errors.New("error opening environment file")
			}

			err = utils.WriteToEnvFile(envFile, map[string]string{
				"ALLOW_ORIGINS":           "",
				"ALLOW_ORIGINS_WILDCARDS": "",
			})
			if err != nil {
				return err
			}
		}

		tmpl, err := template.ParseFS(templates.Files, MiddlewareOptionTemplatePathMap[option])
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

		fmt.Println(fmt.Sprintf("✅ %s generated successfully.", option))
	}
	return nil
}

func ListExistingMiddlewares() ([]string, error) {

	var middlewares []string

	err := filepath.Walk(cli_config.CliConfig.MiddlewaresFolderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(info.Name(), ".go") {
			return nil
		}

		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return fmt.Errorf("failed to parse file %s: %w", path, err)
		}

		for _, decl := range node.Decls {
			if fn, ok := decl.(*ast.FuncDecl); ok && strings.HasSuffix(fn.Name.Name, "Middleware") {

				name := fn.Name.Name
				if strings.HasPrefix(name, "New") {
					name = strings.TrimPrefix(name, "New")
				}

				middlewares = append(middlewares, name)
			}
		}

		return nil
	})

	return middlewares, err

}

func generateAuthJwtFile() error {

	if exists := utils.FileExists(cli_config.CliConfig.AuthFolderPath); !exists {
		err := os.MkdirAll(cli_config.CliConfig.AuthFolderPath, os.ModePerm)
		if err != nil {
			return err
		}
	}

	tmpl, err := template.ParseFS(templates.Files, AuthJwtTemplatePath)
	if err != nil {
		return err
	}

	f, err := os.Create(path.Join(cli_config.CliConfig.AuthFolderPath, "jwt.go"))
	if err != nil {
		return err
	}

	err = tmpl.Execute(f, nil)
	if err != nil {
		return err
	}

	return nil
}
