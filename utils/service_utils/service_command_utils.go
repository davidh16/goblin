package service_utils

import (
	"errors"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"github.com/davidh16/goblin/cli_config"
	"github.com/davidh16/goblin/commands/model"
	centralRepo "github.com/davidh16/goblin/commands/repo/flags/central-repo"
	"github.com/davidh16/goblin/templates"
	"github.com/davidh16/goblin/utils"
	"github.com/davidh16/goblin/utils/model_utils"
	"github.com/davidh16/goblin/utils/repo_utils"
	"github.com/samber/lo"
	"go/ast"
	"go/format"
	"go/parser"
	"go/printer"
	"go/token"
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

// AddNewServiceToCentralService updates central_service.go: injects a new service field
// into CentralService and wires it through NewCentralService. It deduplicates repo args
// and is idempotent (re-runs won't duplicate fields or constructor entries).
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
	repoPackage := func() string {
		parts := strings.Split(repoPackageImport, "/")
		return parts[len(parts)-1]
	}()

	// Import repos package only if needed and missing
	if serviceData.RepoData != nil && len(serviceData.RepoData) > 0 {
		found := false
		for _, imp := range node.Imports {
			if strings.Trim(imp.Path.Value, `""`) == repoPackageImport {
				found = true
				break
			}
		}
		if !found {
			newImport := &ast.ImportSpec{
				Path: &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", repoPackageImport)},
			}
			inserted := false
			for _, decl := range node.Decls {
				if gen, ok := decl.(*ast.GenDecl); ok && gen.Tok == token.IMPORT {
					gen.Specs = append(gen.Specs, newImport)
					inserted = true
					break
				}
			}
			if !inserted {
				node.Decls = append([]ast.Decl{
					&ast.GenDecl{Tok: token.IMPORT, Specs: []ast.Spec{newImport}},
				}, node.Decls...)
			}
		}
	}

	ast.Inspect(node, func(n ast.Node) bool {
		// Add struct field if not already there
		if ts, ok := n.(*ast.TypeSpec); ok && ts.Name.Name == structName {
			if st, ok := ts.Type.(*ast.StructType); ok {
				already := false
				for _, f := range st.Fields.List {
					for _, name := range f.Names {
						if name.Name == serviceData.ServiceFullName {
							already = true
							break
						}
					}
					if already {
						break
					}
				}
				if !already {
					st.Fields.List = append(st.Fields.List, &ast.Field{
						Names: []*ast.Ident{ast.NewIdent(serviceData.ServiceFullName)},
						Type:  ast.NewIdent(attributeDataType),
					})
				}
				// st.Fields.Opening = token.Pos(1) // ← not needed
			}
		}

		// Mutate NewCentralService
		if fn, ok := n.(*ast.FuncDecl); ok && fn.Name.Name == constructorName {
			// Ensure parameter *repos.CentralRepo exists if service needs repos
			if serviceData.RepoData != nil && len(serviceData.RepoData) > 0 {
				paramExists := false
				for _, p := range fn.Type.Params.List {
					for _, name := range p.Names {
						if name.Name == "centralRepo" {
							paramExists = true
							break
						}
					}
					if paramExists {
						break
					}
				}
				if !paramExists {
					fn.Type.Params.List = append(fn.Type.Params.List, &ast.Field{
						Names: []*ast.Ident{ast.NewIdent("centralRepo")},
						Type: &ast.StarExpr{
							X: &ast.SelectorExpr{
								X:   ast.NewIdent(repoPackage),
								Sel: ast.NewIdent("CentralRepo"),
							},
						},
					})
				}
			}

			// Build constructor args: unique repos in order (no duplicates)
			constructorArgs := []ast.Expr{}
			if serviceData.RepoData != nil {
				seen := map[string]struct{}{}
				for _, rd := range serviceData.RepoData {
					if _, dup := seen[rd.RepoFullName]; dup {
						continue
					}
					seen[rd.RepoFullName] = struct{}{}
					constructorArgs = append(constructorArgs, &ast.SelectorExpr{
						X:   ast.NewIdent("centralRepo"),
						Sel: ast.NewIdent(rd.RepoFullName),
					})
				}
			}

			// Update or append the field in the return &CentralService{ ... }
			if len(fn.Body.List) > 0 {
				if ret, ok := fn.Body.List[len(fn.Body.List)-1].(*ast.ReturnStmt); ok && len(ret.Results) > 0 {
					if ul, ok := ret.Results[0].(*ast.UnaryExpr); ok {
						if cl, ok := ul.X.(*ast.CompositeLit); ok {
							replaced := false
							for i, elt := range cl.Elts {
								if kv, ok := elt.(*ast.KeyValueExpr); ok {
									if key, ok := kv.Key.(*ast.Ident); ok && key.Name == serviceData.ServiceFullName {
										kv.Value = &ast.CallExpr{
											Fun:  ast.NewIdent(serviceConstructor),
											Args: constructorArgs,
										}
										cl.Elts[i] = kv
										replaced = true
										break
									}
								}
							}
							if !replaced {
								cl.Elts = append(cl.Elts, &ast.KeyValueExpr{
									Key: ast.NewIdent(serviceData.ServiceFullName),
									Value: &ast.CallExpr{
										Fun:  ast.NewIdent(serviceConstructor),
										Args: constructorArgs,
									},
								})
							}
						}
					}
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
	return cfg.Fprint(outFile, fileSet, node)
}

// CreateService generates a new service Go file from a predefined template.
// It fills in the service entity and package data into the template and writes it to the destination path.
func CreateService(serviceData *ServiceData) error {
	tmpl, err := template.ParseFS(templates.Files, ServiceTemplatePath)
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

// AddRepoToService adds repository dependencies to a service (struct, constructor, return literal)
// and updates the central wiring. All updates are idempotent.
func AddRepoToService(serviceData *ServiceData) error {
	repoPackageImport := path.Join(cli_config.CliConfig.ProjectName, cli_config.CliConfig.RepositoriesFolderPath)
	repoPackage := func() string {
		parts := strings.Split(repoPackageImport, "/")
		return parts[len(parts)-1]
	}()

	fileSet := token.NewFileSet()
	node, err := parser.ParseFile(fileSet, serviceData.ServiceFilePath, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("failed to parse file: %w", err)
	}

	// ensure import once
	hasImport := false
	for _, imp := range node.Imports {
		if strings.Trim(imp.Path.Value, `"`) == repoPackageImport {
			hasImport = true
			break
		}
	}
	if !hasImport {
		newImp := &ast.ImportSpec{Path: &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", repoPackageImport)}}
		inserted := false
		for _, d := range node.Decls {
			if g, ok := d.(*ast.GenDecl); ok && g.Tok == token.IMPORT {
				g.Specs = append(g.Specs, newImp)
				inserted = true
				break
			}
		}
		if !inserted {
			node.Decls = append([]ast.Decl{
				&ast.GenDecl{Tok: token.IMPORT, Specs: []ast.Spec{newImp}},
			}, node.Decls...)
		}
	}

	structUpdated := false
	constructorUpdated := false

	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {

		// --- struct: add missing fields only
		case *ast.TypeSpec:
			if x.Name.Name != serviceData.ServiceFullName {
				return true
			}
			st, ok := x.Type.(*ast.StructType)
			if !ok {
				return true
			}
			for _, repo := range serviceData.RepoData {
				exists := false
				for _, f := range st.Fields.List {
					for _, nm := range f.Names {
						if nm.Name == repo.RepoFullName {
							exists = true
							break
						}
					}
					if exists {
						break
					}
				}
				if !exists {
					st.Fields.List = append(st.Fields.List, &ast.Field{
						Names: []*ast.Ident{ast.NewIdent(repo.RepoFullName)},
						Type:  ast.NewIdent(repoPackage + "." + repo.RepoFullName + "Interface"),
					})
				}
			}
			structUpdated = true

		// --- constructor: add missing params; replace-or-append return fields
		case *ast.FuncDecl:
			if x.Name.Name != "New"+serviceData.ServiceFullName {
				return true
			}

			// params
			for _, repo := range serviceData.RepoData {
				want := utils.PascalToCamel(repo.RepoFullName)
				paramExists := false
				for _, p := range x.Type.Params.List {
					for _, nm := range p.Names {
						if nm.Name == want {
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
						Names: []*ast.Ident{ast.NewIdent(want)},
						Type: &ast.SelectorExpr{
							X:   ast.NewIdent(repoPackage),
							Sel: ast.NewIdent(repo.RepoFullName + "Interface"),
						},
					})
				}
			}

			// find the return stmt (prefer the last one)
			var ret *ast.ReturnStmt
			for i := len(x.Body.List) - 1; i >= 0; i-- {
				if r, ok := x.Body.List[i].(*ast.ReturnStmt); ok {
					ret = r
					break
				}
			}
			if ret != nil && len(ret.Results) > 0 {
				if ul, ok := ret.Results[0].(*ast.UnaryExpr); ok {
					if cl, ok := ul.X.(*ast.CompositeLit); ok {
						// replace-or-append each repo field
						for _, repo := range serviceData.RepoData {
							key := repo.RepoFullName
							val := ast.NewIdent(utils.PascalToCamel(repo.RepoFullName))
							replaced := false
							for i, elt := range cl.Elts {
								if kv, ok := elt.(*ast.KeyValueExpr); ok {
									if id, ok := kv.Key.(*ast.Ident); ok && id.Name == key {
										kv.Value = val
										cl.Elts[i] = kv
										replaced = true
										break
									}
								}
							}
							if !replaced {
								cl.Elts = append(cl.Elts, &ast.KeyValueExpr{
									Key:   ast.NewIdent(key),
									Value: val,
								})
							}
						}
						constructorUpdated = true
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

	// write service file
	out, err := os.Create(serviceData.ServiceFilePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer out.Close()
	if err := printer.Fprint(out, fileSet, node); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	// --- Update CentralService (replace args with unique centralRepo.<Repo>) ---
	centralPath := path.Join(cli_config.CliConfig.ServicesFolderPath, "central_service.go")
	cset := token.NewFileSet()
	cnode, err := parser.ParseFile(cset, centralPath, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("failed to parse central service file: %w", err)
	}

	centralUpdated := false

	ast.Inspect(cnode, func(n ast.Node) bool {
		fn, ok := n.(*ast.FuncDecl)
		if !ok || fn.Name.Name != "NewCentralService" {
			return true
		}
		// find return &CentralService{ ... }
		var ret *ast.ReturnStmt
		for i := len(fn.Body.List) - 1; i >= 0; i-- {
			if r, ok := fn.Body.List[i].(*ast.ReturnStmt); ok {
				ret = r
				break
			}
		}
		if ret == nil || len(ret.Results) == 0 {
			return true
		}
		ul, ok := ret.Results[0].(*ast.UnaryExpr)
		if !ok {
			return true
		}
		cl, ok := ul.X.(*ast.CompositeLit)
		if !ok {
			return true
		}

		target := serviceData.ServiceFullName // e.g., TestService
		for _, elt := range cl.Elts {
			kv, ok := elt.(*ast.KeyValueExpr)
			if !ok {
				continue
			}
			key, ok := kv.Key.(*ast.Ident)
			if !ok || key.Name != target {
				continue
			}
			call, ok := kv.Value.(*ast.CallExpr)
			if !ok {
				continue
			}

			// Rebuild args uniquely
			seen := map[string]struct{}{}
			newArgs := make([]ast.Expr, 0, len(call.Args)+len(serviceData.RepoData))

			// keep existing unique centralRepo.<Repo>
			for _, a := range call.Args {
				if sel, ok := a.(*ast.SelectorExpr); ok {
					if id, ok := sel.X.(*ast.Ident); ok && id.Name == "centralRepo" {
						name := sel.Sel.Name
						if _, ok := seen[name]; !ok {
							seen[name] = struct{}{}
							newArgs = append(newArgs, a)
						}
					}
				}
			}
			// ensure current repos are present
			for _, repo := range serviceData.RepoData {
				if _, ok := seen[repo.RepoFullName]; ok {
					continue
				}
				seen[repo.RepoFullName] = struct{}{}
				newArgs = append(newArgs, &ast.SelectorExpr{
					X:   ast.NewIdent("centralRepo"),
					Sel: ast.NewIdent(repo.RepoFullName),
				})
			}

			call.Args = newArgs
			centralUpdated = true
			break
		}
		return true
	})

	if !centralUpdated {
		return fmt.Errorf("failed to update NewCentralService constructor call to New%s", serviceData.ServiceFullName)
	}

	out2, err := os.Create(centralPath)
	if err != nil {
		return fmt.Errorf("failed to open central service file: %w", err)
	}
	defer out2.Close()
	if err := printer.Fprint(out2, cset, cnode); err != nil {
		return fmt.Errorf("failed to write central service file: %w", err)
	}

	return nil
}

// CopyRepoMethodsToService copies selected methods from a repository interface to a service interface.
// It also generates proxy methods that call the underlying repository methods directly from the service.
func CopyRepoMethodsToService(serviceData *ServiceData, methodNames []string) error {
	fileSet := token.NewFileSet()

	// Parse the service file
	serviceAst, err := parser.ParseFile(fileSet, serviceData.ServiceFilePath, nil, parser.AllErrors)
	if err != nil {
		return err
	}

	// Helper: ensure an import exists (no-op if already present)
	ensureImport := func(f *ast.File, importPath string) {
		if importPath == "" {
			return
		}
		for _, imp := range f.Imports {
			if strings.Trim(imp.Path.Value, `"`) == importPath {
				return // already imported
			}
		}
		newImport := &ast.ImportSpec{
			Path: &ast.BasicLit{Kind: token.STRING, Value: fmt.Sprintf("%q", importPath)},
		}
		// Try to append to an existing import block
		for _, decl := range f.Decls {
			if gen, ok := decl.(*ast.GenDecl); ok && gen.Tok == token.IMPORT {
				gen.Specs = append(gen.Specs, newImport)
				return
			}
		}
		// No import block; create one at the top
		f.Decls = append([]ast.Decl{
			&ast.GenDecl{Tok: token.IMPORT, Specs: []ast.Spec{newImport}},
		}, f.Decls...)
	}

	// Conditionally import models (existing behavior)
	modelImport := path.Join(cli_config.CliConfig.ProjectName, cli_config.CliConfig.ModelsFolderPath)
	ensureImport(serviceAst, modelImport)

	paginationImport := path.Join(cli_config.CliConfig.ProjectName, cli_config.CliConfig.DatabaseInstancesFolderPath)
	needsPagination := false
	for _, name := range methodNames {
		if strings.Contains(name, "WithPagination") {
			needsPagination = true
			break
		}
	}
	if needsPagination {
		ensureImport(serviceAst, paginationImport)
	}

	for _, repo := range serviceData.RepoData {
		// Parse each repo file
		repoAst, err := parser.ParseFile(fileSet, repo.RepoFilePath, nil, parser.AllErrors)
		if err != nil {
			return err
		}

		// Find repo interface and collect selected methods
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
					if method == nil || len(method.Names) == 0 {
						continue // skip embedded or malformed
					}
					if _, ok := method.Type.(*ast.FuncType); !ok {
						continue // skip non-methods
					}
					for _, n := range method.Names {
						if contains(methodNames, n.Name) {
							methodMap[n.Name] = method
						}
					}
				}
			}
		}

		if len(methodMap) == 0 {
			return fmt.Errorf("no methods found for %s", repo.RepoFullName+"Interface")
		}

		// 4) Update service interface: append selected methods (naive append; dedupe if needed)
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
				for _, methodName := range methodNames {
					if m, ok := methodMap[methodName]; ok {
						ifaceType.Methods.List = append(ifaceType.Methods.List, m)
					}
				}
			}
		}

		// 5) Update service struct methods (proxy methods)
		for _, methodName := range methodNames {
			m, ok := methodMap[methodName]
			if !ok {
				continue
			}
			funcDecl, err := buildProxyFuncDecl(m, serviceData.ServiceFullName, &repo)
			if err != nil {
				return err
			}
			serviceAst.Decls = append(serviceAst.Decls, funcDecl)
		}
	}

	// 6) Write back to service file (fix: create file before deferring Close)
	f, err := os.Create(serviceData.ServiceFilePath)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

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
