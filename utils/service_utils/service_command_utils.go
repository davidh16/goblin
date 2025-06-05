package service_utils

import (
	"errors"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/samber/lo"
	"go/ast"
	"go/format"
	"go/parser"
	"go/printer"
	"go/token"
	"goblin/cli_config"
	"goblin/commands/model"
	centralRepo "goblin/commands/repo/flags/central-repo"
	"goblin/utils"
	"goblin/utils/model_utils"
	"goblin/utils/repo_utils"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"
)

type ServiceData struct {
	ServiceNameSnakeCase                  string // i.e user
	ServiceEntity                         string // i.e User
	ServiceFullName                       string // i.e UserService
	ServiceFileName                       string // i.e user_repo.go
	ServiceFilePath                       string // i.e services/user_repo.go
	SelectedServiceProxyMethodToImplement []string
	ModelEntity                           string // i.e User
	CentralServiceExists                  bool
	RepoStrategy                          RepoStrategy
	RepoData                              []repo_utils.RepoData
}

// NewServiceData creates and returns a new ServiceData struct pointer.
// It initializes its RepoData field with a new RepoData instance.
func NewServiceData() *ServiceData {
	return &ServiceData{
		RepoData: []repo_utils.RepoData{},
	}
}

var RepoStrategyOptionsMap = map[RepoStrategy]string{
	RepoStrategyUnspecified:      "Unspecified",
	RepoStrategyNewRepo:          "Create new repo",
	RepoStrategyExistingRepo:     "Use existing repo",
	RepoStrategyNoImplementation: "No implementation of repo",
}

var RepoOptionsStrategyMap = map[string]RepoStrategy{
	"Create new repo":           RepoStrategyNewRepo,
	"Use existing repo":         RepoStrategyExistingRepo,
	"No implementation of repo": RepoStrategyNoImplementation,
}

type RepoStrategy int

const (
	RepoStrategyUnspecified RepoStrategy = iota
	RepoStrategyNewRepo
	RepoStrategyExistingRepo
	RepoStrategyNoImplementation
)

// PrepareRepo interacts with the user through CLI prompts to configure a new repository setup.
// It asks for repo name, model strategy, and repository methods to implement, then returns the filled repo_utils.RepoData.
func PrepareRepo() *repo_utils.RepoData {

	repoData := repo_utils.NewRepoData()

	for {
		if err := survey.AskOne(&survey.Input{
			Message: "Please type the repository file name (snake_case) :",
			Default: "my_repo_file",
		}, &repoData.RepoNameSnakeCase); err != nil {
			utils.HandleError(err)
		}

		if !utils.IsSnakeCase(repoData.RepoNameSnakeCase) {
			fmt.Printf("🛑 %s is not in snake case\n", repoData.RepoNameSnakeCase)
			continue
		}

		var confirmContinue bool
		confirmPrompt := &survey.Confirm{
			Message: fmt.Sprintf("You are about to create a repo file named %s_repo.go, do you want to continue ?", repoData.RepoNameSnakeCase),
		}
		if err := survey.AskOne(confirmPrompt, &confirmContinue); err != nil {
			utils.HandleError(err)
		}

		if !confirmContinue {
			continue
		}

		repoData.RepoEntity = utils.SnakeToPascal(repoData.RepoNameSnakeCase)
		repoData.RepoFullName = repoData.RepoEntity + "Repo"
		repoData.RepoFileName = repoData.RepoNameSnakeCase + "_repo.go"
		repoData.RepoFilePath = path.Join(cli_config.CliConfig.RepositoriesFolderPath, repoData.RepoFileName)

		if utils.FileExists(repoData.RepoFilePath) {
			var overwriteConfirmed bool
			confirmPrompt = &survey.Confirm{
				Message: fmt.Sprintf("%s repository already exists. Do you want to overwrite it ?", repoData.RepoFileName),
				Default: false,
			}
			if err := survey.AskOne(confirmPrompt, &overwriteConfirmed); err != nil {
				utils.HandleError(err)
			}

			if overwriteConfirmed {
				confirmPrompt = &survey.Confirm{
					Message: fmt.Sprintf("Are you sure you want to overwrite %s repository ?", repoData.RepoFileName),
					Default: false,
				}
				if err := survey.AskOne(confirmPrompt, &overwriteConfirmed); err != nil {
					utils.HandleError(err)
				}
			}

			if !overwriteConfirmed {
				continue
			}
		}
		break
	}

	repoData.CentralRepoExists = utils.FileExists(path.Join(cli_config.CliConfig.RepositoriesFolderPath, "central_repo.go"))

	existingModels, err := repo_utils.ListExistingModels()
	if err != nil {
		utils.HandleError(err)
	}

	options := []string{repo_utils.ModelStrategyOptionsMap[repo_utils.ModelStrategyNewModel]}
	if len(existingModels) > 0 {
		options = append(options, repo_utils.ModelStrategyOptionsMap[repo_utils.ModelStrategyExistingModel])
	}

	var optionChoice string
	err = survey.AskOne(&survey.Select{
		Message: "Choose model strategy:",
		Options: options,
	}, &optionChoice)
	if err != nil {
		utils.HandleError(err)
	}

	repoData.ModelStrategy = repo_utils.ModelOptionsStrategyMap[optionChoice]

	switch repoData.ModelStrategy {
	case repo_utils.ModelStrategyNewModel:
		modelData, err := model_utils.TriggerGetModelNameFlow()
		if err != nil {
			utils.HandleError(err)
		}

		repoData.ModelData = modelData

	case repo_utils.ModelStrategyExistingModel:
		existingModelOptionsModelDataMap := map[string]*model_utils.ModelData{}
		existingModelOptions := lo.Map(existingModels, func(item model_utils.ModelData, index int) string {
			existingModelOptionsModelDataMap[item.ModelEntity] = &item
			return item.ModelEntity
		})

		var selectedModelOption string
		err = survey.AskOne(&survey.Select{
			Message: "Select a model to use:",
			Options: existingModelOptions,
		}, &selectedModelOption)
		if err != nil {
			utils.HandleError(err)
		}

		repoData.ModelData = existingModelOptionsModelDataMap[selectedModelOption]
	default:
		utils.HandleError(fmt.Errorf("invalid model strategy: %d", repoData.ModelStrategy))
	}

	var decision string
	prompt := &survey.Select{
		Message: repo_utils.GenerateImplementRepoMethodsNowQuestion(repoData.ModelData.ModelEntity),
		Options: []string{
			"Yes, choose methods to implement",
			"No, skip this step",
		},
	}
	err = survey.AskOne(prompt, &decision)
	if err != nil {
		utils.HandleError(err)
	}

	toImplementRepoMethods := decision == "Yes, choose methods to implement"

	if toImplementRepoMethods {
		selectMethodsPrompt := &survey.MultiSelect{
			Message: "Which methods do you want to implement?\n  [Press enter without selecting any of the options to skip]\n",
			Options: repo_utils.GenerateSortedRepoMethodNames(repoData.ModelData.ModelEntity),
		}
		err = survey.AskOne(selectMethodsPrompt, &repoData.SelectedRepoMethodsToImplement)
		if err != nil {
			utils.HandleError(err)
		}
	}

	return repoData
}

// ExecuteCreateRepo executes the full creation process for a repository.
// It generates the model (if needed), central repo (if needed), repository file, and optionally implements selected methods.
func ExecuteCreateRepo(repoData []repo_utils.RepoData) error {
	for _, repo := range repoData {
		if repo.ModelStrategy == repo_utils.ModelStrategyNewModel {
			err := model.CreateModel(repo.ModelData)
			if err != nil {
				return err
			}
		}

		// create central repo
		if !repo.CentralRepoExists {
			centralRepo.GenerateCentralRepo()
		}

		// add repo to central repo
		if !utils.FileExists(repo.RepoFilePath) {
			err := repo_utils.AddNewRepoToCentralRepo(&repo)
			if err != nil {
				utils.HandleError(err)
			}
		}

		// create repo
		err := repo_utils.CreateRepo(&repo)
		if err != nil {
			return err
		}

		if len(repo.SelectedRepoMethodsToImplement) > 0 {
			rawMethodsMap := repo_utils.GenerateRepoMethodNamesMap(repo.ModelData.ModelEntity)
			selectedRawMethods := lo.Map(repo.SelectedRepoMethodsToImplement, func(item string, index int) repo_utils.Method {
				return rawMethodsMap[item]
			})

			err = repo_utils.AddMethodsToRepo(&repo, selectedRawMethods)
			if err != nil {
				return err
			}
		}

		fmt.Println(fmt.Sprintf("✅ %s repository generated successfully.", repo.RepoEntity))
	}
	return nil
}

// ListExistingRepos scans the repositories folder and identifies existing repositories
// by locating types with a WithTx method signature. It returns metadata about each found repository.
func ListExistingRepos() ([]repo_utils.RepoData, error) {
	var repos []repo_utils.RepoData

	err := filepath.WalkDir(cli_config.CliConfig.RepositoriesFolderPath, func(repoPath string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !strings.HasSuffix(repoPath, ".go") {
			return nil // skip non-Go files
		}

		fileSet := token.NewFileSet()
		node, err := parser.ParseFile(fileSet, repoPath, nil, parser.ParseComments)
		if err != nil {
			return err
		}

		for _, decl := range node.Decls {
			funcDecl, ok := decl.(*ast.FuncDecl)
			if !ok || funcDecl.Recv == nil || funcDecl.Name.Name != "WithTx" {
				continue
			}

			if len(funcDecl.Recv.List) == 0 {
				continue
			}

			recvType, ok := funcDecl.Recv.List[0].Type.(*ast.StarExpr)
			if !ok {
				continue
			}
			ident, ok := recvType.X.(*ast.Ident)
			if !ok {
				continue
			}
			structName := ident.Name

			if funcDecl.Type.Params == nil || len(funcDecl.Type.Params.List) != 1 {
				continue
			}

			param := funcDecl.Type.Params.List[0]
			paramType, ok := param.Type.(*ast.StarExpr)
			if !ok {
				continue
			}
			paramIdent, ok := paramType.X.(*ast.SelectorExpr)
			if !ok {
				continue
			}
			pkgIdent, ok := paramIdent.X.(*ast.Ident)
			if !ok {
				continue
			}
			if pkgIdent.Name != "gorm" || paramIdent.Sel.Name != "DB" {
				continue
			}

			// Found a matching repository!
			repos = append(repos, repo_utils.RepoData{
				RepoEntity:        strings.Trim(structName, "Repo"),
				RepoFullName:      structName,
				RepoFilePath:      repoPath,
				RepoFileName:      strings.Split(repoPath, "/")[len(strings.Split(repoPath, "/"))-1],
				RepoNameSnakeCase: utils.PascalToSnake(structName),
			})
		}

		return nil
	})
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
	}

	return repos, nil
}

// ListExistingRepoMethods returns a list of method names defined for the given repository entity.
// It walks through the repository Go files and collects methods based on receiver types.
func ListExistingRepoMethods(repoData *repo_utils.RepoData) ([]string, error) {
	var methods []string

	fileSet := token.NewFileSet()
	node, err := parser.ParseFile(fileSet, repoData.RepoFilePath, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	for _, decl := range node.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok || funcDecl.Recv == nil {
			continue
		}

		if funcDecl.Name.Name == "WithTx" {
			continue
		}

		// Check receiver type
		recv := funcDecl.Recv.List[0].Type

		// If pointer receiver, unwrap
		if starExpr, ok := recv.(*ast.StarExpr); ok {
			recv = starExpr.X
		}

		recvIdent, ok := recv.(*ast.Ident)
		if !ok {
			continue
		}

		if recvIdent.Name == repoData.RepoFullName {
			methods = append(methods, funcDecl.Name.Name)
		}
	}

	return methods, nil
}

// GenerateImplementProxyMethodsNowQuestionWithExistingRepoMethodsPreview builds a formatted string that
// previews the available proxy methods for implementation.
//
// It returns a message string like:
//
//	Do you want to implement service proxy methods now?
//	--------------------------------------------
//	Available methods:
//	CreateCar
//	DeleteCar
//	ListCarsWithPagination
//	...
//	--------------------------------------------
//
// This is used as a message for survey.Select or other CLI confirmations.
func GenerateImplementProxyMethodsNowQuestionWithExistingRepoMethodsPreview(repoData *repo_utils.RepoData, existingRepoMethods []string) string {

	message := fmt.Sprintf("Do you want to implement service proxy methods for %s now?\n", repoData.RepoFullName)
	message += "--------------------------------------------\n"
	message += "Available methods:\n"

	for i := 0; i < len(existingRepoMethods); i++ {
		message += existingRepoMethods[i] + "\n"
	}

	message += "--------------------------------------------\n"

	return message

}

// AddNewServiceToCentralService updates the central_service.go file by injecting a new
// service interface as a field into the CentralService struct, and wiring it
// through the constructor NewCentralService.
//
// It automatically converts the given snake_case service name into PascalCase
// to follow Go naming conventions for attributes and types.
func AddNewServiceToCentralService(serviceData *ServiceData) error {
	centralServiceFilePath := path.Join(cli_config.CliConfig.ServicesFolderPath, "central_service.go")
	const structName = "CentralService"
	const constructorName = "NewCentralService"

	attributeDataType := serviceData.ServiceFullName + "Interface"
	serviceConstructor := "New" + serviceData.ServiceFullName

	fileSet := token.NewFileSet()
	node, err := parser.ParseFile(fileSet, centralServiceFilePath, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	repoPackageImport := path.Join(cli_config.CliConfig.ProjectName, cli_config.CliConfig.RepositoriesFolderPath)
	repoPackage := strings.Split(repoPackageImport, "/")[len(strings.Split(repoPackageImport, "/"))-1]

	// if serviceData.RepoData != nil that means we are adding a new repo, therefore we should also add imports if needed, if we have no implemented repo, we can skip this step
	if serviceData.RepoData != nil {
		// Track if import already exists
		importFound := false
		for _, imp := range node.Imports {
			impPath := strings.Trim(imp.Path.Value, `"`)
			if impPath == repoPackageImport {
				importFound = true
				break
			}
		}

		// Add import if needed
		if !importFound {
			newImport := &ast.ImportSpec{
				Path: &ast.BasicLit{
					Kind:  token.STRING,
					Value: fmt.Sprintf("%q", repoPackageImport),
				},
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
				importDecl := &ast.GenDecl{
					Tok: token.IMPORT,
					Specs: []ast.Spec{
						newImport,
					},
				}
				node.Decls = append([]ast.Decl{importDecl}, node.Decls...)
			}
		}
	}

	ast.Inspect(node, func(n ast.Node) bool {
		if gen, ok := n.(*ast.TypeSpec); ok && gen.Name.Name == structName {
			if structType, ok := gen.Type.(*ast.StructType); ok {
				structType.Fields.List = append(structType.Fields.List, &ast.Field{
					Names: []*ast.Ident{ast.NewIdent(serviceData.ServiceFullName)},
					Type:  ast.NewIdent(attributeDataType),
				})
				structType.Fields.Opening = token.Pos(1)
			}
		}

		if fn, ok := n.(*ast.FuncDecl); ok && fn.Name.Name == constructorName {

			var constructorArgs []ast.Expr
			for _, repoData := range serviceData.RepoData {

				constructorArgs = append(constructorArgs, &ast.SelectorExpr{
					X:   ast.NewIdent("centralRepo"),
					Sel: ast.NewIdent(repoData.RepoFullName),
				})

				paramValue := "centralRepo"
				receiverType := "CentralRepo"

				paramExists := false
				for _, param := range fn.Type.Params.List {
					for _, name := range param.Names {
						if name.Name == utils.PascalToCamel(paramValue) {
							paramExists = true
							break
						}
					}
					if paramExists {
						break
					}
				}

				if !paramExists && serviceData.RepoData != nil {
					fn.Type.Params.List = append(fn.Type.Params.List, &ast.Field{
						Names: []*ast.Ident{ast.NewIdent(paramValue)},
						Type: &ast.StarExpr{
							X: &ast.SelectorExpr{
								X:   ast.NewIdent(repoPackage),
								Sel: ast.NewIdent(receiverType),
							},
						},
					})
				}
			}

			// if service has no repo implementation, constructorArgs will be empty
			if retStmt, ok := fn.Body.List[len(fn.Body.List)-1].(*ast.ReturnStmt); ok {
				if compLit, ok := retStmt.Results[0].(*ast.UnaryExpr).X.(*ast.CompositeLit); ok {
					compLit.Elts = append(compLit.Elts, &ast.KeyValueExpr{
						Key: ast.NewIdent(serviceData.ServiceFullName),
						Value: &ast.CallExpr{
							Fun:  ast.NewIdent(serviceConstructor),
							Args: constructorArgs, // 🔥 dynamic argument list
						},
					})
				}
			}
		}

		return true
	})

	outFile, err := os.Create(centralServiceFilePath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	cfg := &printer.Config{Mode: printer.UseSpaces | printer.TabIndent, Tabwidth: 4}
	if err = cfg.Fprint(outFile, fileSet, node); err != nil {
		return err
	}
	return nil
}

// CreateService generates a new service Go file from a predefined template.
// It fills in the service entity and package data into the template and writes it to the destination path.
func CreateService(serviceData *ServiceData) error {
	tmpl, err := template.ParseFiles(ServiceTemplatePath)
	if err != nil {
		return err
	}

	f, err := os.Create(serviceData.ServiceFilePath)
	if err != nil {
		return err
	}
	defer f.Close()

	templateData := struct {
		ServicePackage string
		ServiceEntity  string
	}{
		ServicePackage: strings.Split(cli_config.CliConfig.ServicesFolderPath, "/")[len(strings.Split(cli_config.CliConfig.ServicesFolderPath, "/"))-1],
		ServiceEntity:  serviceData.ServiceEntity,
	}

	err = tmpl.Execute(f, templateData)
	if err != nil {
		return err
	}
	return nil
}

// AddRepoToService adds a repository dependency to a service struct and its constructor.
// It updates the service file's AST to add the repo interface to the struct and constructor parameters.
func AddRepoToService(serviceData *ServiceData) error {

	repoPackageImport := path.Join(cli_config.CliConfig.ProjectName, cli_config.CliConfig.RepositoriesFolderPath)
	repoPackage := strings.Split(repoPackageImport, "/")[len(strings.Split(repoPackageImport, "/"))-1]

	fileSet := token.NewFileSet()
	// Parse the file
	node, err := parser.ParseFile(fileSet, serviceData.ServiceFilePath, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("failed to parse file: %w", err)
	}

	// Track if import already exists
	importFound := false
	for _, imp := range node.Imports {
		impPath := strings.Trim(imp.Path.Value, `"`)
		if impPath == repoPackageImport {
			importFound = true
			break
		}
	}

	// Add import if needed
	if !importFound {
		newImport := &ast.ImportSpec{
			Path: &ast.BasicLit{
				Kind:  token.STRING,
				Value: fmt.Sprintf("%q", repoPackageImport),
			},
		}

		inserted := false
		// Try to find existing import block
		for _, decl := range node.Decls {
			if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.IMPORT {
				genDecl.Specs = append(genDecl.Specs, newImport)
				inserted = true
				break
			}
		}

		if !inserted {
			// No import block found, create one
			importDecl := &ast.GenDecl{
				Tok: token.IMPORT,
				Specs: []ast.Spec{
					newImport,
				},
			}
			// Insert it at the beginning of declarations
			node.Decls = append([]ast.Decl{importDecl}, node.Decls...)
		}
	}

	// Track if we updated struct and constructor
	structUpdated := false
	constructorUpdated := false

	// Walk the AST
	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {

		// Update the struct
		case *ast.TypeSpec:
			if x.Name.Name == serviceData.ServiceFullName {
				if structType, ok := x.Type.(*ast.StructType); ok {
					for _, repo := range serviceData.RepoData {
						structType.Fields.List = append(structType.Fields.List, &ast.Field{
							Names: []*ast.Ident{ast.NewIdent(repo.RepoFullName)},
							Type:  ast.NewIdent(repoPackage + "." + repo.RepoFullName + "Interface"),
						})
						structUpdated = true
					}
				}
			}

		// Update the constructor
		case *ast.FuncDecl:
			if x.Name.Name == "New"+serviceData.ServiceFullName {
				for _, repo := range serviceData.RepoData {
					// First: ensure the parameter is added if missing
					paramExists := false
					for _, param := range x.Type.Params.List {
						for _, name := range param.Names {
							if name.Name == utils.PascalToCamel(repo.RepoFullName) {
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
							Names: []*ast.Ident{ast.NewIdent(utils.PascalToCamel(repo.RepoFullName))},
							Type: &ast.SelectorExpr{
								X:   ast.NewIdent(repoPackage),
								Sel: ast.NewIdent(repo.RepoFullName + "Interface"),
							},
						})
					}
				}

				// Then: update the constructor body as you already did
				if len(x.Body.List) > 0 {
					if retStmt, ok := x.Body.List[0].(*ast.ReturnStmt); ok {
						if compositeLit, ok := retStmt.Results[0].(*ast.UnaryExpr).X.(*ast.CompositeLit); ok {
							for _, repo := range serviceData.RepoData {
								compositeLit.Elts = append(compositeLit.Elts, &ast.KeyValueExpr{
									Key:   ast.NewIdent(repo.RepoFullName),
									Value: ast.NewIdent(utils.PascalToCamel(repo.RepoFullName)),
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
		return fmt.Errorf("struct %s not found", serviceData.ServiceFullName)
	}
	if !constructorUpdated {
		return fmt.Errorf("constructor New%s not found", serviceData.ServiceFullName)
	}

	// Create the output file
	outFile, err := os.Create(serviceData.ServiceFilePath)
	if err != nil {
		return fmt.Errorf("failed to open file for writing: %w", err)
	}
	defer outFile.Close()

	// Write the modified AST back to the file
	err = printer.Fprint(outFile, fileSet, node)
	if err != nil {
		return fmt.Errorf("failed to write updated file: %w", err)
	}

	return nil
}

// CopyRepoMethodsToService copies selected methods from a repository interface to a service interface.
// It also generates proxy methods that call the underlying repository methods directly from the service.
func CopyRepoMethodsToService(serviceData *ServiceData, methodNames []string) error {
	// 1. Parse repository file
	fileSet := token.NewFileSet()
	var f *os.File
	defer f.Close()

	// 3. Parse service file
	serviceAst, err := parser.ParseFile(fileSet, serviceData.ServiceFilePath, nil, parser.AllErrors)
	if err != nil {
		return err
	}

	for _, repo := range serviceData.RepoData {
		repoAst, err := parser.ParseFile(fileSet, repo.RepoFilePath, nil, parser.AllErrors)
		if err != nil {
			return err
		}

		// 2. Find repo interface and extract methods
		methodMap := make(map[string]*ast.Field)

		for _, decl := range repoAst.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.TYPE {
				continue
			}
			for _, spec := range genDecl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok || typeSpec.Name.Name != repo.RepoFullName+"Interface" {
					continue
				}
				ifaceType, ok := typeSpec.Type.(*ast.InterfaceType)
				if !ok {
					continue
				}

				for _, method := range ifaceType.Methods.List {
					if method == nil {
						continue
					}

					if len(method.Names) == 0 {
						continue // Skip embedded interfaces
					}
					if _, ok := method.Type.(*ast.FuncType); !ok {
						continue // Skip non-methods
					}
					for _, name := range method.Names {
						if contains(methodNames, name.Name) {
							methodMap[name.Name] = method
						}
					}
				}
			}
		}

		if len(methodMap) == 0 {
			return fmt.Errorf("no methods found for %s", repo.RepoFullName+"Interface")
		}

		modelImport := path.Join(cli_config.CliConfig.ProjectName, cli_config.CliConfig.ModelsFolderPath)
		//modelPackage := strings.Split(modelImport, "/")[strings.LastIndex(modelImport, "/")]

		// Track if import already exists
		importFound := false
		for _, imp := range serviceAst.Imports {
			impPath := strings.Trim(imp.Path.Value, `"`)
			if impPath == modelImport {
				importFound = true
				break
			}
		}

		// Add import if needed
		if !importFound {
			newImport := &ast.ImportSpec{
				Path: &ast.BasicLit{
					Kind:  token.STRING,
					Value: fmt.Sprintf("%q", modelImport),
				},
			}

			inserted := false
			// Try to find existing import block
			for _, decl := range serviceAst.Decls {
				if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.IMPORT {
					genDecl.Specs = append(genDecl.Specs, newImport)
					inserted = true
					break
				}
			}

			if !inserted {
				// No import block found, create one
				importDecl := &ast.GenDecl{
					Tok: token.IMPORT,
					Specs: []ast.Spec{
						newImport,
					},
				}
				// Insert it at the beginning of declarations
				serviceAst.Decls = append([]ast.Decl{importDecl}, serviceAst.Decls...)
			}
		}

		// 4. Update service interface
		for _, decl := range serviceAst.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.TYPE {
				continue
			}
			for _, spec := range genDecl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok || typeSpec.Name.Name != serviceData.ServiceFullName+"Interface" {
					continue
				}
				ifaceType, ok := typeSpec.Type.(*ast.InterfaceType)
				if !ok {
					continue
				}

				// Append methods
				for _, methodName := range methodNames {
					if method, ok := methodMap[methodName]; ok {
						ifaceType.Methods.List = append(ifaceType.Methods.List, method)
					}
				}
			}
		}

		// 5. Update service struct methods
		for _, methodName := range methodNames {
			method, found := methodMap[methodName]
			if !found {
				continue
			}
			funcDecl, err := buildProxyFuncDecl(method, serviceData.ServiceFullName, &repo)
			if err != nil {
				return err
			}
			serviceAst.Decls = append(serviceAst.Decls, funcDecl)
		}

		// 6. Write back to service file
		f, err = os.Create(serviceData.ServiceFilePath)
		if err != nil {
			return err
		}
	}

	return format.Node(f, fileSet, serviceAst)
}

func contains(arr []string, str string) bool {
	for _, v := range arr {
		if v == str {
			return true
		}
	}
	return false
}

func buildProxyFuncDecl(method *ast.Field, serviceFullName string, repo *repo_utils.RepoData) (*ast.FuncDecl, error) {

	funcType, ok := method.Type.(*ast.FuncType)
	if !ok || funcType == nil {
		return nil, errors.New(fmt.Sprintf("method %v does not have a valid FuncType", method.Names[0].Name))
	}

	var paramNames []ast.Expr
	if funcType.Params != nil {
		for _, param := range funcType.Params.List {
			for _, name := range param.Names {
				paramNames = append(paramNames, ast.NewIdent(name.Name))
			}
		}
	}

	return &ast.FuncDecl{
		Recv: &ast.FieldList{
			List: []*ast.Field{
				{
					Names: []*ast.Ident{ast.NewIdent("s")},
					Type: &ast.StarExpr{
						X: ast.NewIdent(serviceFullName), // now passed dynamically
					},
				},
			},
		},
		Name: method.Names[0],
		Type: funcType,
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.ReturnStmt{
					Results: []ast.Expr{
						&ast.CallExpr{
							Fun: &ast.SelectorExpr{
								X: &ast.SelectorExpr{
									X:   ast.NewIdent("s"),
									Sel: ast.NewIdent(repo.RepoFullName),
								},
								Sel: ast.NewIdent(method.Names[0].Name),
							},
							Args: paramNames,
						},
					},
				},
			},
		},
	}, nil
}

// ListExistingRepos scans the repositories folder and identifies existing repositories
// by locating types with a WithTx method signature. It returns metadata about each found repository.
func ListExistingServices() ([]ServiceData, error) {
	var services []ServiceData

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
			funcDecl, ok := decl.(*ast.FuncDecl)
			if !ok || funcDecl.Name == nil {
				continue
			}

			// Looking for function name starting with "New" and ending with "Service"
			if !strings.HasPrefix(funcDecl.Name.Name, "New") || !strings.HasSuffix(funcDecl.Name.Name, "Service") {
				continue
			}

			if funcDecl.Name.Name == "NewCentralService" {
				continue
			}

			// Check if return type is a pointer to struct ending with "Service"
			if funcDecl.Type.Results == nil || len(funcDecl.Type.Results.List) != 1 {
				continue
			}

			starExpr, ok := funcDecl.Type.Results.List[0].Type.(*ast.StarExpr)
			if !ok {
				continue
			}

			ident, ok := starExpr.X.(*ast.Ident)
			if !ok || !strings.HasSuffix(ident.Name, "Service") {
				continue
			}

			serviceEntity := strings.TrimSuffix(ident.Name, "Service")

			// Found a matching servicesitory!
			services = append(services, ServiceData{
				ServiceEntity:        serviceEntity,
				ServiceFullName:      ident.Name,
				ServiceFilePath:      servicePath,
				ServiceFileName:      strings.Split(servicePath, "/")[len(strings.Split(servicePath, "/"))-1],
				ServiceNameSnakeCase: utils.PascalToSnake(serviceEntity),
			})
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

func AddCentralServiceToCentralControllerConstructor() error {
	fset := token.NewFileSet()

	centralControllerFilePath := path.Join(cli_config.CliConfig.ControllersFolderPath, "central_controller.go")

	node, err := parser.ParseFile(fset, centralControllerFilePath, nil, parser.AllErrors)
	if err != nil {
		return err
	}

	// Add import if not already present
	importPath := path.Join(cli_config.CliConfig.ProjectName, cli_config.CliConfig.ServicesFolderPath)
	hasImport := false
	for _, imp := range node.Imports {
		if imp.Path.Value == fmt.Sprintf("\"%s\"", importPath) {
			hasImport = true
			break
		}
	}
	if !hasImport {
		newImport := &ast.ImportSpec{
			Path: &ast.BasicLit{
				Kind:  token.STRING,
				Value: fmt.Sprintf("\"%s\"", importPath),
			},
		}
		// Add to the import declarations
		found := false
		for _, decl := range node.Decls {
			if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.IMPORT {
				genDecl.Specs = append(genDecl.Specs, newImport)
				found = true
				break
			}
		}
		if !found {
			// No existing import block, create one
			node.Decls = append([]ast.Decl{
				&ast.GenDecl{
					Tok: token.IMPORT,
					Specs: []ast.Spec{
						newImport,
					},
				},
			}, node.Decls...)
		}
	}

	// Modify constructor parameter
	ast.Inspect(node, func(n ast.Node) bool {
		fn, ok := n.(*ast.FuncDecl)
		if !ok || fn.Name.Name != "NewCentralController" {
			return true
		}

		// Check if centralService already exists
		for _, param := range fn.Type.Params.List {
			if len(param.Names) > 0 && param.Names[0].Name == "centralService" {
				return false
			}
		}

		// Add parameter: centralService *services.CentralService
		param := &ast.Field{
			Names: []*ast.Ident{ast.NewIdent("centralService")},
			Type: &ast.StarExpr{
				X: &ast.SelectorExpr{
					X:   ast.NewIdent("services"),
					Sel: ast.NewIdent("CentralService"),
				},
			},
		}
		fn.Type.Params.List = append(fn.Type.Params.List, param)

		return false
	})

	// Write back to file
	file, err := os.Create(centralControllerFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	return printer.Fprint(file, fset, node)
}
