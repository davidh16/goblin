package router_utils

import (
	"goblin/cli_config"
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
		ImplementMiddlewares     bool
		MiddlewaresPackage       string
		MiddlewaresPackageImport string
	}{
		ImplementMiddlewares:     routerData.ImplementMiddlewares,
		MiddlewaresPackage:       strings.Split(cli_config.CliConfig.MiddlewaresFolderPath, "/")[len(strings.Split(cli_config.CliConfig.MiddlewaresFolderPath, "/"))-1],
		MiddlewaresPackageImport: path.Join(cli_config.CliConfig.ProjectName, cli_config.CliConfig.MiddlewaresFolderPath),
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
