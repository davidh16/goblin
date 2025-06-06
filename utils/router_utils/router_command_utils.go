package router_utils

import (
	"goblin/cli_config"
	"goblin/utils"
	"os"
	"path"
	"strings"
	"text/template"
)

type RouterData struct {
	LoggingMiddleware     bool
	RecoverMiddleware     bool
	RateLimiterMiddleware bool
	AllowOriginMiddleware bool
	ImplementMiddlewares  bool
}

func NewRouterData() *RouterData {
	return &RouterData{}
}

func GenerateRouter(routerData *RouterData) error {

	if !utils.FileExists(cli_config.CliConfig.RouterFolderPath) {
		err := os.MkdirAll(cli_config.CliConfig.RouterFolderPath, os.ModePerm)
		if err != nil {
			return err
		}
	}

	tmpl, err := template.ParseFiles(RouterTemplatePath)
	if err != nil {
		return err
	}

	f, err := os.Create(path.Join(cli_config.CliConfig.RouterFolderPath, "router.go"))
	if err != nil {
		return err
	}
	defer f.Close()

	templateData := struct {
		RouterPackage            string
		ImplementMiddlewares     bool
		MiddlewaresPackage       string
		MiddlewaresPackageImport string
		RecoverMiddleware        bool
		AllowOriginMiddleware    bool
		RateLimiterMiddleware    bool
		LoggingMiddleware        bool
	}{
		RouterPackage:            strings.Split(cli_config.CliConfig.RouterFolderPath, "/")[len(strings.Split(cli_config.CliConfig.RouterFolderPath, "/"))-1],
		ImplementMiddlewares:     routerData.ImplementMiddlewares,
		MiddlewaresPackage:       strings.Split(cli_config.CliConfig.MiddlewaresFolderPath, "/")[len(strings.Split(cli_config.CliConfig.MiddlewaresFolderPath, "/"))-1],
		MiddlewaresPackageImport: path.Join(cli_config.CliConfig.ProjectName, cli_config.CliConfig.MiddlewaresFolderPath),
		RecoverMiddleware:        routerData.RecoverMiddleware,
		AllowOriginMiddleware:    routerData.AllowOriginMiddleware,
		RateLimiterMiddleware:    routerData.RateLimiterMiddleware,
		LoggingMiddleware:        routerData.LoggingMiddleware,
	}

	err = tmpl.Execute(f, templateData)
	if err != nil {
		return err
	}

	tmpl, err = template.ParseFiles(CustomBinderTemplatePath)
	if err != nil {
		return err
	}

	f, err = os.Create(path.Join(cli_config.CliConfig.RouterFolderPath, "custom_request_binder.go"))
	if err != nil {
		return err
	}

	err = tmpl.Execute(f, templateData)
	if err != nil {
		return err
	}

	return nil
}
