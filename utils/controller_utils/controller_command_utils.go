package controller_utils

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/davidh16/goblin/cli_config"
	central_service "github.com/davidh16/goblin/commands/service/flags/central-service"
	"github.com/davidh16/goblin/templates"
	"github.com/davidh16/goblin/utils"
	"github.com/davidh16/goblin/utils/repo_utils"
	"github.com/davidh16/goblin/utils/service_utils"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"
)

type ControllerData struct {
	ControllerNameSnakeCase                  string // i.e user
	ControllerEntity                         string // i.e User
	ControllerFullName                       string // i.e UserController
	ControllerFileName                       string // i.e user_service.go
	ControllerFilePath                       string // i.e services/user_service.go
	SelectedControllerProxyMethodToImplement []string
	ModelEntity                              string // i.e User
	CentralControllerExists                  bool
	ServiceStrategy                          ServiceStrategy
	ServiceData                              []service_utils.ServiceData
}

// NewControllerData creates and returns a new ControllerData struct pointer.
// It initializes its ServiceData field with a new ServiceData instance.
func NewControllerData() *ControllerData {
	return &ControllerData{
		ServiceData: []service_utils.ServiceData{},
	}
}

var ServiceStrategyOptionsMap = map[ServiceStrategy]string{
	ServiceStrategyUnspecified:      "Unspecified",
	ServiceStrategyNewService:       "Create new service",
	ServiceStrategyExistingService:  "Use existing service",
	ServiceStrategyNoImplementation: "No implementation of service",
}

var ServiceOptionsStrategyMap = map[string]ServiceStrategy{
	"Create new service":           ServiceStrategyNewService,
	"Use existing service":         ServiceStrategyExistingService,
	"No implementation of service": ServiceStrategyNoImplementation,
}

type ServiceStrategy int

const (
	ServiceStrategyUnspecified ServiceStrategy = iota
	ServiceStrategyNewService
	ServiceStrategyExistingService
	ServiceStrategyNoImplementation
)

// ListExistingServices scans the services folder and identifies existing services
// service struct is detected by checking if it has a suffix "Service" in its name
func ListExistingServices() ([]service_utils.ServiceData, error) {
	var services []service_utils.ServiceData

	err := filepath.WalkDir(cli_config.CliConfig.ServicesFolderPath, func(servicePath string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !strings.HasSuffix(servicePath, ".go") {
			return nil // skip non-Go files
		}

		fileSet := token.NewFileSet()
		node, err := parser.ParseFile(fileSet, servicePath, nil, parser.ParseComments)
		if err != nil {
			return err
		}

		for _, decl := range node.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.TYPE {
				continue
			}
			for _, spec := range genDecl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				if !strings.HasSuffix(typeSpec.Name.Name, "Service") {
					continue
				}

				if typeSpec.Name.Name == "CentralService" {
					continue
				}

				// Check that it's actually a struct type
				_, ok = typeSpec.Type.(*ast.StructType)
				if !ok {
					continue
				}

				structName := typeSpec.Name.Name
				services = append(services, service_utils.ServiceData{
					ServiceEntity:        strings.TrimSuffix(structName, "Service"),
					ServiceFullName:      structName,
					ServiceFilePath:      servicePath,
					ServiceFileName:      filepath.Base(servicePath),
					ServiceNameSnakeCase: utils.PascalToSnake(strings.TrimSuffix(structName, "Service")),
				})
			}
		}

		return nil
	})
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
	}

	return services, nil
}

func PrepareService() (*service_utils.ServiceData, error) {
	serviceData := service_utils.NewServiceData()
	defaultServiceName := "my_service_file"
	for {
		if err := survey.AskOne(&survey.Input{
			Message: "Please type the services file name (snake_case) :",
			Default: defaultServiceName,
		}, &serviceData.ServiceNameSnakeCase); err != nil {
			return nil, err
		}

		if !utils.IsSnakeCase(serviceData.ServiceNameSnakeCase) {
			fmt.Printf("🛑 %s is not in snake case\n", serviceData.ServiceNameSnakeCase)
			continue
		}

		defaultServiceName = serviceData.ServiceNameSnakeCase

		var confirm bool
		confirmPrompt := &survey.Confirm{
			Message: fmt.Sprintf("You are about to create a service file named %s_service.go, do you want to continue ?", serviceData.ServiceNameSnakeCase),
		}
		if err := survey.AskOne(confirmPrompt, &confirm); err != nil {
			return nil, err
		}

		if !confirm {
			continue
		}

		serviceData.ServiceEntity = utils.SnakeToPascal(serviceData.ServiceNameSnakeCase)
		serviceData.ServiceFullName = serviceData.ServiceEntity + "Service"
		serviceData.ServiceFileName = serviceData.ServiceNameSnakeCase + "_service.go"
		serviceData.ServiceFilePath = path.Join(cli_config.CliConfig.ServicesFolderPath, serviceData.ServiceFileName)

		if utils.FileExists(serviceData.ServiceFilePath) {
			var confirmOverwrite bool
			confirmPrompt = &survey.Confirm{
				Message: fmt.Sprintf("%s service already exists. Do you want to overwrite it ?", serviceData.ServiceFileName),
				Default: false,
			}
			if err := survey.AskOne(confirmPrompt, &confirmOverwrite); err != nil {
				return nil, err
			}

			if !confirmOverwrite {
				continue
			}

			if confirmOverwrite {
				confirmPrompt = &survey.Confirm{
					Message: fmt.Sprintf("Are you sure you want to overwrite %s service ?", serviceData.ServiceFileName),
					Default: false,
				}
				if err := survey.AskOne(confirmPrompt, &confirmOverwrite); err != nil {
					return nil, err
				}
			}

			if !confirmOverwrite {
				continue
			}
		}
		break
	}

	serviceData.CentralServiceExists = utils.FileExists(path.Join(cli_config.CliConfig.ServicesFolderPath, "central_service.go"))

	existingRepos, err := service_utils.ListExistingRepos()
	if err != nil {
		return nil, err
	}
	existingReposMap := make(map[string]*repo_utils.RepoData)
	for _, repo := range existingRepos {
		existingReposMap[repo.RepoFullName] = &repo
	}

	// service_utils.RepoOptionsStrategyMap keys are used to list options for choosing repo strategy, if there are no existing repos, that key (option) has to be removed from the map
	if len(existingRepos) == 0 {
		delete(service_utils.RepoOptionsStrategyMap, service_utils.RepoStrategyOptionsMap[service_utils.RepoStrategyExistingRepo])
	}

	var repoStrategyChosenOption string
	err = survey.AskOne(&survey.Select{
		Message: "Choose repo strategy:",
		Options: utils.Keys(service_utils.RepoOptionsStrategyMap),
	}, &repoStrategyChosenOption)
	if err != nil {
		return nil, err
	}

	serviceData.RepoStrategy = service_utils.RepoOptionsStrategyMap[repoStrategyChosenOption]

	if serviceData.RepoStrategy == service_utils.RepoStrategyExistingRepo {
		var chosenRepos []string
		err = survey.AskOne(&survey.MultiSelect{
			Message: "Select a repo to use:",
			Options: utils.Keys(existingReposMap),
		}, &chosenRepos)
		if err != nil {
			utils.HandleError(err)
		}

		for _, repo := range chosenRepos {
			serviceData.RepoData = append(serviceData.RepoData, repo_utils.RepoData{
				RepoEntity: repo,
			})
		}
	}

	if serviceData.RepoStrategy == service_utils.RepoStrategyNewRepo {
		serviceData.RepoData = append(serviceData.RepoData, *service_utils.PrepareRepo())
	}

	var toImplement bool
	switch serviceData.RepoStrategy {
	case service_utils.RepoStrategyNewRepo:
		var decision string
		prompt := &survey.Select{
			Message: service_utils.GenerateImplementProxyMethodsNowQuestionWithExistingRepoMethodsPreview(&serviceData.RepoData[0], serviceData.RepoData[0].SelectedRepoMethodsToImplement),
			Options: []string{
				"Yes, choose methods to implement",
				"No, skip this step",
			},
		}
		err = survey.AskOne(prompt, &decision)
		if err != nil {
			return nil, err
		}

		toImplement = decision == "Yes, choose methods to implement"

		if toImplement {
			selectedServiceProxyMethodsPrompt := &survey.MultiSelect{
				Message: "Which service proxy methods do you want to implement?\n  [Press enter without selecting any of the options to skip]\n",
				Options: serviceData.RepoData[0].SelectedRepoMethodsToImplement,
			}
			err = survey.AskOne(selectedServiceProxyMethodsPrompt, &serviceData.SelectedServiceProxyMethodToImplement)
			if err != nil {
				return nil, err
			}
		}

	case service_utils.RepoStrategyExistingRepo:
		for _, repo := range serviceData.RepoData {
			existingRepoMethods, err := service_utils.ListExistingRepoMethods(&repo)
			if err != nil {
				return nil, err
			}

			var decision string
			prompt := &survey.Select{
				Message: service_utils.GenerateImplementProxyMethodsNowQuestionWithExistingRepoMethodsPreview(&repo, existingRepoMethods),
				Options: []string{
					"Yes, choose methods to implement",
					"No, skip this step",
				},
			}
			err = survey.AskOne(prompt, &decision)
			if err != nil {
				return nil, err
			}
			toImplement = decision == "Yes, choose methods to implement"
			if toImplement {
				selectMethodsToImplementPrompt := &survey.MultiSelect{
					Message: "Which methods do you want to implement?\n  [Press enter without selecting any of the options to skip]\n",
					Options: existingRepoMethods,
				}
				err = survey.AskOne(selectMethodsToImplementPrompt, &serviceData.SelectedServiceProxyMethodToImplement)
				if err != nil {
					return nil, err
				}
			}
		}

	case service_utils.RepoStrategyNoImplementation:
		serviceData.RepoData = nil
	default:
		return nil, fmt.Errorf("invalid repo strategy: %v", serviceData.RepoStrategy)
	}

	return serviceData, nil
}

func ExecuteCreateService(serviceData []service_utils.ServiceData) error {
	for _, service := range serviceData {
		if service.RepoStrategy == service_utils.RepoStrategyNewRepo {
			err := service_utils.ExecuteCreateRepo(service.RepoData)
			if err != nil {
				return err
			}
		}

		if !service.CentralServiceExists {
			central_service.GenerateCentralService()
		}

		if !utils.FileExists(service.ServiceFilePath) {
			err := service_utils.AddNewServiceToCentralService(&service)
			if err != nil {
				return err
			}
		}

		err := service_utils.CreateService(&service)
		if err != nil {
			return err
		}

		if service.RepoStrategy != service_utils.RepoStrategyNoImplementation {
			err = service_utils.AddRepoToService(&service)
			if err != nil {
				return err
			}
		}

		if len(service.SelectedServiceProxyMethodToImplement) > 0 {
			err = service_utils.CopyRepoMethodsToService(&service, service.SelectedServiceProxyMethodToImplement)
			if err != nil {
				return err
			}
		}

		fmt.Println(fmt.Sprintf("✅ %s service generated successfully.", service.ServiceEntity))
	}

	return nil
}

// AddNewControllerToCentralController updates the central_service.go file by injecting a new
// controller interface as a field into the CentralController struct, and wiring it
// through the constructor NewCentralController.
//
// It automatically converts the given snake_case controller name into PascalCase
// to follow Go naming conventions for attributes and types.
func AddNewControllerToCentralController(controllerData *ControllerData) error {
	centralControllerFilePath := path.Join(cli_config.CliConfig.ControllersFolderPath, "central_controller.go")
	const structName = "CentralController"
	const constructorName = "NewCentralController"

	attributeDataType := "*" + controllerData.ControllerFullName
	controllerConstructor := "New" + controllerData.ControllerFullName

	fileSet := token.NewFileSet()
	node, err := parser.ParseFile(fileSet, centralControllerFilePath, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	servicePackageImport := path.Join(cli_config.CliConfig.ProjectName, cli_config.CliConfig.ServicesFolderPath)
	servicePackage := path.Base(servicePackageImport)
	validatorImport := "github.com/davidh16/goblin/validator"
	validatorAlias := "validator"

	// Add imports for services and validator if missing
	importsNeeded := map[string]string{}
	if controllerData.ServiceData != nil {
		importsNeeded[servicePackageImport] = servicePackage
	}
	importsNeeded[validatorImport] = validatorAlias

	for impPath, asName := range importsNeeded {
		found := false
		for _, imp := range node.Imports {
			if strings.Trim(imp.Path.Value, `"`) == impPath {
				found = true
				break
			}
		}
		if !found {
			newImp := &ast.ImportSpec{
				Path: &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", impPath)},
			}
			if asName != "" {
				newImp.Name = ast.NewIdent(asName)
			}
			insertPos := -1
			for _, decl := range node.Decls {
				if gd, ok := decl.(*ast.GenDecl); ok && gd.Tok == token.IMPORT {
					gd.Specs = append(gd.Specs, newImp)
					insertPos = -2
					break
				}
			}
			if insertPos == -1 {
				imDecl := &ast.GenDecl{Tok: token.IMPORT, Specs: []ast.Spec{newImp}}
				node.Decls = append([]ast.Decl{imDecl}, node.Decls...)
			}
		}
	}

	ast.Inspect(node, func(n ast.Node) bool {
		if ts, ok := n.(*ast.TypeSpec); ok && ts.Name.Name == structName {
			if st, ok := ts.Type.(*ast.StructType); ok {
				st.Fields.List = append(st.Fields.List, &ast.Field{
					Names: []*ast.Ident{ast.NewIdent(controllerData.ControllerFullName)},
					Type:  ast.NewIdent(attributeDataType),
				})
				st.Fields.Opening = token.Pos(1)
			}
		}

		if fn, ok := n.(*ast.FuncDecl); ok && fn.Name.Name == constructorName {
			var consArgs []ast.Expr

			// Inject validator instantiation at the top
			declared := false
			for _, stmt := range fn.Body.List {
				if asn, ok := stmt.(*ast.AssignStmt); ok {
					if ident, ok := asn.Lhs[0].(*ast.Ident); ok && ident.Name == "validator" {
						declared = true
						break
					}
				}
			}
			if !declared {
				declStmt := &ast.AssignStmt{
					Lhs: []ast.Expr{ast.NewIdent("validator")},
					Tok: token.DEFINE,
					Rhs: []ast.Expr{&ast.CallExpr{
						Fun: &ast.SelectorExpr{
							X:   ast.NewIdent(validatorAlias),
							Sel: ast.NewIdent("NewValidator"),
						},
					}},
				}
				fn.Body.List = append([]ast.Stmt{declStmt}, fn.Body.List...)
			}

			// Collect service args
			if controllerData.ServiceData != nil {
				for _, svc := range controllerData.ServiceData {
					consArgs = append(consArgs, &ast.SelectorExpr{
						X:   ast.NewIdent("centralService"),
						Sel: ast.NewIdent(svc.ServiceFullName),
					})
				}
			}
			// Always add validator
			consArgs = append(consArgs, ast.NewIdent("validator"))

			// Update return composite-literal
			if ret, ok := fn.Body.List[len(fn.Body.List)-1].(*ast.ReturnStmt); ok {
				if ul, ok := ret.Results[0].(*ast.UnaryExpr); ok {
					if cl, ok := ul.X.(*ast.CompositeLit); ok {
						cl.Elts = append(cl.Elts, &ast.KeyValueExpr{
							Key: ast.NewIdent(controllerData.ControllerFullName),
							Value: &ast.CallExpr{
								Fun:  ast.NewIdent(controllerConstructor),
								Args: consArgs,
							},
						})
					}
				}
			}
		}
		return true
	})

	out, err := os.Create(centralControllerFilePath)
	if err != nil {
		return err
	}
	defer out.Close()

	cfg := &printer.Config{Mode: printer.UseSpaces | printer.TabIndent, Tabwidth: 4}
	return cfg.Fprint(out, fileSet, node)
}

func CreateValidator() error {

	err := os.MkdirAll(cli_config.CliConfig.ValidatorFolderPath, 0755)
	if err != nil {
		return err
	}

	tmpl, err := template.ParseFS(templates.Files, ValidatorTemplatePath)
	if err != nil {
		return err
	}

	validatorFilePath := path.Join(cli_config.CliConfig.ValidatorFolderPath, "validator.go")
	f, err := os.Create(validatorFilePath)
	if err != nil {
		return err
	}
	defer f.Close()

	templateData := struct {
		ValidatorPackage string
	}{
		ValidatorPackage: strings.Split(cli_config.CliConfig.ValidatorFolderPath, "/")[len(strings.Split(cli_config.CliConfig.ValidatorFolderPath, "/"))-1],
	}

	err = tmpl.Execute(f, templateData)
	if err != nil {
		return err
	}
	return nil
}

// CreateController generates a new controller Go file from a predefined template.
// It fills in the controller entity and package data into the template and writes it to the destination path.
func CreateController(controllerData *ControllerData) error {
	tmpl, err := template.ParseFS(templates.Files, ControllerTemplatePath)
	if err != nil {
		return err
	}

	f, err := os.Create(controllerData.ControllerFilePath)
	if err != nil {
		return err
	}
	defer f.Close()

	templateData := struct {
		ControllerPackage      string
		ControllerEntity       string
		ValidatorPackageImport string
	}{
		ControllerPackage:      strings.Split(cli_config.CliConfig.ControllersFolderPath, "/")[len(strings.Split(cli_config.CliConfig.ControllersFolderPath, "/"))-1],
		ControllerEntity:       controllerData.ControllerEntity,
		ValidatorPackageImport: path.Join(cli_config.CliConfig.ProjectName, cli_config.CliConfig.ValidatorFolderPath),
	}

	err = tmpl.Execute(f, templateData)
	if err != nil {
		return err
	}
	return nil
}

// AddServiceToController adds a service dependency to a controller struct and its constructor.
// It updates the controller file's AST to add the service interface to the struct and constructor parameters.
func AddServiceToController(controllerData *ControllerData) error {
	servicePackageImport := path.Join(cli_config.CliConfig.ProjectName, cli_config.CliConfig.ServicesFolderPath)
	servicePackage := strings.Split(servicePackageImport, "/")[len(strings.Split(servicePackageImport, "/"))-1]

	fileSet := token.NewFileSet()
	node, err := parser.ParseFile(fileSet, controllerData.ControllerFilePath, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("failed to parse file: %w", err)
	}

	// Add import if needed
	importFound := false
	for _, imp := range node.Imports {
		if strings.Trim(imp.Path.Value, `"`) == servicePackageImport {
			importFound = true
			break
		}
	}
	if !importFound {
		newImport := &ast.ImportSpec{
			Path: &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", servicePackageImport)},
		}
		inserted := false
		for _, decl := range node.Decls {
			if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.IMPORT {
				genDecl.Specs = append(genDecl.Specs, newImport)
				inserted = true
				break
			}
		}
		if !inserted {
			node.Decls = append([]ast.Decl{&ast.GenDecl{
				Tok: token.IMPORT,
				Specs: []ast.Spec{
					newImport,
				},
			}}, node.Decls...)
		}
	}

	structUpdated := false
	constructorUpdated := false

	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.TypeSpec:
			if x.Name.Name == controllerData.ControllerFullName {
				if structType, ok := x.Type.(*ast.StructType); ok {
					for _, service := range controllerData.ServiceData {
						structType.Fields.List = append(structType.Fields.List, &ast.Field{
							Names: []*ast.Ident{ast.NewIdent(service.ServiceFullName)},
							Type:  ast.NewIdent(servicePackage + "." + service.ServiceFullName + "Interface"),
						})
					}
					structUpdated = true
				}
			}
		case *ast.FuncDecl:
			if x.Name.Name == "New"+controllerData.ControllerFullName {
				for _, service := range controllerData.ServiceData {
					paramExists := false
					for _, param := range x.Type.Params.List {
						for _, name := range param.Names {
							if name.Name == utils.PascalToCamel(service.ServiceFullName) {
								paramExists = true
								break
							}
						}
						if paramExists {
							break
						}
					}
					if !paramExists {
						x.Type.Params.List = append(x.Type.Params.List, &ast.Field{
							Names: []*ast.Ident{ast.NewIdent(utils.PascalToCamel(service.ServiceFullName))},
							Type: &ast.SelectorExpr{
								X:   ast.NewIdent(servicePackage),
								Sel: ast.NewIdent(service.ServiceFullName + "Interface"),
							},
						})
					}
					if len(x.Body.List) > 0 {
						if retStmt, ok := x.Body.List[0].(*ast.ReturnStmt); ok {
							if compositeLit, ok := retStmt.Results[0].(*ast.UnaryExpr).X.(*ast.CompositeLit); ok {
								compositeLit.Elts = append(compositeLit.Elts, &ast.KeyValueExpr{
									Key:   ast.NewIdent(service.ServiceFullName),
									Value: ast.NewIdent(utils.PascalToCamel(service.ServiceFullName)),
								})
								constructorUpdated = true
							}
						}
					}
				}
			}
		}
		return true
	})

	if !structUpdated {
		return fmt.Errorf("struct %sController not found", controllerData.ControllerEntity)
	}
	if !constructorUpdated {
		return fmt.Errorf("constructor New%sController not found", controllerData.ControllerEntity)
	}

	outFile, err := os.Create(controllerData.ControllerFilePath)
	if err != nil {
		return fmt.Errorf("failed to open controller file for writing: %w", err)
	}
	defer outFile.Close()
	if err := printer.Fprint(outFile, fileSet, node); err != nil {
		return fmt.Errorf("failed to write updated controller file: %w", err)
	}

	// Now update the CentralController constructor
	centralControllerFilePath := path.Join(cli_config.CliConfig.ControllersFolderPath, "central_controller.go")
	centralFileSet := token.NewFileSet()
	centralNode, err := parser.ParseFile(centralFileSet, centralControllerFilePath, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("failed to parse central controller file: %w", err)
	}

	centralUpdated := false

	ast.Inspect(centralNode, func(n ast.Node) bool {
		if fn, ok := n.(*ast.FuncDecl); ok && fn.Name.Name == "NewCentralController" {
			for _, stmt := range fn.Body.List {
				if retStmt, ok := stmt.(*ast.ReturnStmt); ok {
					if unaryExpr, ok := retStmt.Results[0].(*ast.UnaryExpr); ok {
						if compositeLit, ok := unaryExpr.X.(*ast.CompositeLit); ok {
							for _, elt := range compositeLit.Elts {
								if kv, ok := elt.(*ast.KeyValueExpr); ok {
									if callExpr, ok := kv.Value.(*ast.CallExpr); ok {
										if ident, ok := callExpr.Fun.(*ast.Ident); ok {
											for _, service := range controllerData.ServiceData {
												expected := "New" + controllerData.ControllerFullName
												if ident.Name == expected {
													callExpr.Args = append(callExpr.Args, &ast.SelectorExpr{
														X:   ast.NewIdent("centralService"),
														Sel: ast.NewIdent(service.ServiceFullName),
													})
													centralUpdated = true
												}
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
		return true
	})

	if !centralUpdated {
		return fmt.Errorf("failed to update NewCentralController with New%sController arguments", controllerData.ControllerEntity)
	}

	outCentralFile, err := os.Create(centralControllerFilePath)
	if err != nil {
		return fmt.Errorf("failed to open central controller file for writing: %w", err)
	}
	defer outCentralFile.Close()
	if err := printer.Fprint(outCentralFile, centralFileSet, centralNode); err != nil {
		return fmt.Errorf("failed to write updated central controller file: %w", err)
	}

	return nil
}

// ListExistingControllers scans the services folder and identifies existing services
// service struct is detected by checking if it has a suffix "Controller" in its name
func ListExistingControllers() ([]ControllerData, error) {
	var controllers []ControllerData

	err := filepath.WalkDir(cli_config.CliConfig.ControllersFolderPath, func(controllerPath string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !strings.HasSuffix(controllerPath, ".go") {
			return nil // skip non-Go files
		}

		fileSet := token.NewFileSet()
		node, err := parser.ParseFile(fileSet, controllerPath, nil, parser.ParseComments)
		if err != nil {
			return err
		}

		for _, decl := range node.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.TYPE {
				continue
			}
			for _, spec := range genDecl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				if !strings.HasSuffix(typeSpec.Name.Name, "Controller") {
					continue
				}

				if typeSpec.Name.Name == "CentralController" {
					continue
				}

				// Check that it's actually a struct type
				_, ok = typeSpec.Type.(*ast.StructType)
				if !ok {
					continue
				}

				structName := typeSpec.Name.Name
				controllers = append(controllers, ControllerData{
					ControllerEntity:        strings.TrimSuffix(structName, "Controller"),
					ControllerFullName:      structName,
					ControllerFilePath:      controllerPath,
					ControllerFileName:      filepath.Base(controllerPath),
					ControllerNameSnakeCase: utils.PascalToSnake(strings.TrimSuffix(structName, "Controller")),
				})
			}
		}

		return nil
	})
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
	}

	return controllers, nil
}
