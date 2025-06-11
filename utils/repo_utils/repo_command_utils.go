package repo_utils

import (
	"fmt"
	"github.com/davidh16/goblin/cli_config"
	"github.com/davidh16/goblin/templates"
	"github.com/davidh16/goblin/utils"
	"github.com/davidh16/goblin/utils/model_utils"
	"github.com/jinzhu/inflection"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
)

const (
	RepoTemplatePath = "repo.tmpl"
)

type RepoData struct {
	RepoNameSnakeCase              string // i.e user
	RepoEntity                     string // i.e User
	RepoFullName                   string // i.e UserRepo
	RepoFileName                   string // i.e user_repo.go
	RepoFilePath                   string // i.e repos/user_repo.go
	SelectedRepoMethodsToImplement []string
	CentralRepoExists              bool
	ModelStrategy                  ModelStrategy // i.e ModelStrategyNewModel
	ModelData                      *model_utils.ModelData
}

// NewRepoData returns a pointer to a newly initialized RepoData struct.
//
// It sets up default values including a new instance of ModelData.
func NewRepoData() *RepoData {
	return &RepoData{
		ModelData: model_utils.NewModelData(),
	}
}

type ModelStrategy int

const (
	ModelStrategyUnspecified ModelStrategy = iota
	ModelStrategyNewModel
	ModelStrategyExistingModel
)

var ModelStrategyOptionsMap map[ModelStrategy]string = map[ModelStrategy]string{
	ModelStrategyUnspecified:   "Unspecified",
	ModelStrategyNewModel:      "Create new model",
	ModelStrategyExistingModel: "Use existing model",
}

var ModelOptionsStrategyMap map[string]ModelStrategy = map[string]ModelStrategy{
	"Unspecified":        ModelStrategyUnspecified,
	"Create new model":   ModelStrategyNewModel,
	"Use existing model": ModelStrategyExistingModel,
}

type method struct {
	Signature *ast.Field
	Function  *ast.FuncDecl
}

type Method int

const (
	Create Method = iota + 1
	Update
	Delete
	ListAll
	ListWithPagination
	GetByUuid
)

var RepoRawMethodsMap = map[Method]string{
	Create:             "Create",
	Update:             "Update",
	Delete:             "Delete",
	GetByUuid:          "GetByUuid",
	ListAll:            "ListAll",
	ListWithPagination: "ListWithPagination",
}

// AddNewRepoToCentralRepo injects a new repository into central_repo.go.
// It updates the CentralRepo struct to include the new repository interface,
// and modifies the constructor (NewCentralRepo) to initialize the repository using its constructor.
func AddNewRepoToCentralRepo(repoData *RepoData) error {
	centralRepoFilePath := path.Join(cli_config.CliConfig.RepositoriesFolderPath, "central_repo.go")
	const structName = "CentralRepo"
	const constructorName = "NewCentralRepo"

	centralRepoAttributeDataType := repoData.RepoFullName + "Interface"
	centralRepoAttributeName := repoData.RepoFullName
	repoConstructor := "New" + repoData.RepoFullName

	fileSet := token.NewFileSet()
	node, err := parser.ParseFile(fileSet, centralRepoFilePath, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	ast.Inspect(node, func(n ast.Node) bool {
		// 1. Add field to CentralRepo struct
		if gen, ok := n.(*ast.TypeSpec); ok && gen.Name.Name == structName {
			if structType, ok := gen.Type.(*ast.StructType); ok {
				structType.Fields.List = append(structType.Fields.List, &ast.Field{
					Names: []*ast.Ident{ast.NewIdent(centralRepoAttributeName)},
					Type:  ast.NewIdent(centralRepoAttributeDataType),
				})
				structType.Fields.Opening = token.Pos(1)
			}
		}

		// 2. Update constructor to use db and call NewXxxRepo(db)
		if fn, ok := n.(*ast.FuncDecl); ok && fn.Name.Name == constructorName {
			// Ensure *gorm.DB parameter exists
			hasDB := false
			for _, p := range fn.Type.Params.List {
				if starExpr, ok := p.Type.(*ast.StarExpr); ok {
					if selExpr, ok := starExpr.X.(*ast.SelectorExpr); ok && selExpr.Sel.Name == "DB" {
						hasDB = true
						break
					}
				}
			}
			if !hasDB {
				fn.Type.Params.List = append(fn.Type.Params.List, &ast.Field{
					Names: []*ast.Ident{ast.NewIdent("db")},
					Type: &ast.StarExpr{
						X: &ast.SelectorExpr{
							X:   ast.NewIdent("gorm"),
							Sel: ast.NewIdent("DB"),
						},
					},
				})
			}

			// Modify return statement to use constructor
			if retStmt, ok := fn.Body.List[len(fn.Body.List)-1].(*ast.ReturnStmt); ok {
				if compLit, ok := retStmt.Results[0].(*ast.UnaryExpr).X.(*ast.CompositeLit); ok {
					compLit.Elts = append(compLit.Elts, &ast.KeyValueExpr{
						Key: ast.NewIdent(centralRepoAttributeName),
						Value: &ast.CallExpr{
							Fun:  ast.NewIdent(repoConstructor),
							Args: []ast.Expr{ast.NewIdent("db")},
						},
					})
				}
			}
		}

		return true
	})

	// Write back to file with pretty formatting
	outFile, err := os.Create(centralRepoFilePath)
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

// GenerateImplementRepoMethodsNowQuestion builds a formatted string that
// previews the available repository methods for implementation.
//
// It returns a message string like:
//
//	Do you want to implement repo methods now?
//	--------------------------------------------
//	Available methods:
//	CreateCar
//	DeleteCar
//	ListCarsWithPagination
//	...
//	--------------------------------------------
//
// This is used as a message for survey.Select or other CLI confirmations.
func GenerateImplementRepoMethodsNowQuestion(modelEntity string) string {

	message := "Do you want to implement repository methods now?\n"
	message += "--------------------------------------------\n"
	message += "Available methods:\n"

	for i := 1; i <= len(RepoRawMethodsMap); i++ {
		message += generateRepoMethodName(Method(i), modelEntity) + "\n"
	}

	message += "--------------------------------------------\n"

	return message

}

// ListExistingModels scans all Go files in the models folder to find structs
// that contain a 'Uuid' field of type string. These are considered valid model types.
// Returns a list of matching model metadata.
func ListExistingModels() ([]model_utils.ModelData, error) {
	var models []model_utils.ModelData

	err := filepath.WalkDir(cli_config.CliConfig.ModelsFolderPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil || !strings.HasSuffix(path, ".go") {
			return nil // skip non-Go files
		}

		fileSet := token.NewFileSet()
		node, err := parser.ParseFile(fileSet, path, nil, parser.ParseComments)
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

				structType, ok := typeSpec.Type.(*ast.StructType)
				if !ok {
					continue
				}

				for _, field := range structType.Fields.List {
					if len(field.Names) == 0 {
						continue
					}

					if field.Names[0].Name == "Uuid" {
						if ident, ok := field.Type.(*ast.Ident); ok && ident.Name == "string" {
							model := model_utils.ModelData{
								NameSnakeCase: utils.PascalToSnake(typeSpec.Name.Name),
								ModelFileName: filepath.Base(path),
								ModelFilePath: path,
								ModelEntity:   typeSpec.Name.Name,
							}
							models = append(models, model)
						}
					}
				}
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return models, nil
}

// AddMethodsToRepo appends selected method signatures to the repository interface
// and adds corresponding method implementations to the repository file.
// It also ensures the necessary model import is present and formats the file with gofmt.
func AddMethodsToRepo(repoData *RepoData, wantedRepoMethods []Method) error {
	fileSet := token.NewFileSet()
	node, err := parser.ParseFile(fileSet, repoData.RepoFilePath, nil, parser.AllErrors)
	if err != nil {
		return err
	}

	var newDecls []ast.Decl

	// Prepare model import path
	modelsPackage := path.Join(cli_config.CliConfig.ProjectName, cli_config.CliConfig.ModelsFolderPath)
	quotedModelsPackage := strconv.Quote(modelsPackage)

	importSpec := &ast.ImportSpec{
		Path: &ast.BasicLit{
			Kind:  token.STRING,
			Value: quotedModelsPackage,
		},
	}

	var importAdded bool

	// Check if import block exists and append to it
	for _, decl := range node.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.IMPORT {
			continue
		}

		// Check if already imported
		for _, spec := range genDecl.Specs {
			if importSpecTyped, ok := spec.(*ast.ImportSpec); ok && importSpecTyped.Path.Value == quotedModelsPackage {
				importAdded = true
				break
			}
		}

		// Append to existing import block if not present
		if !importAdded {
			genDecl.Specs = append(genDecl.Specs, importSpec)
			importAdded = true
			break
		}
	}

	// If no import block exists, create a new one at the top
	if !importAdded {
		newImportDecl := &ast.GenDecl{
			Tok: token.IMPORT,
			Specs: []ast.Spec{
				importSpec,
			},
		}
		node.Decls = append([]ast.Decl{newImportDecl}, node.Decls...)
	}

	// Update interface with method signatures
	ast.Inspect(node, func(n ast.Node) bool {
		if gen, ok := n.(*ast.GenDecl); ok {
			for _, spec := range gen.Specs {
				if iface, ok := spec.(*ast.TypeSpec); ok {
					if iface.Name.Name == repoData.RepoFullName+"Interface" {
						if ifaceType, ok := iface.Type.(*ast.InterfaceType); ok {
							ifaceType.Methods.Opening = token.Pos(1) // force multiline
							for _, repoMethod := range wantedRepoMethods {
								ifaceType.Methods.List = append(
									ifaceType.Methods.List,
									NewRepoMethod(repoMethod, repoData.RepoEntity, repoData.ModelData.ModelEntity).Signature,
								)
							}
						}
					}
				}
			}
		}
		return true
	})

	// Add method implementations at the end
	for _, repoMethod := range wantedRepoMethods {
		newDecls = append(newDecls, NewRepoMethod(repoMethod, repoData.RepoEntity, repoData.ModelData.ModelEntity).Function)
	}
	node.Decls = append(node.Decls, newDecls...)

	// Write back to file
	outFile, err := os.Create(repoData.RepoFilePath)
	if err != nil {
		utils.HandleError(err)
	}
	defer outFile.Close()

	cfg := &printer.Config{Mode: printer.UseSpaces | printer.TabIndent, Tabwidth: 4}
	if err = cfg.Fprint(outFile, fileSet, node); err != nil {
		utils.HandleError(err)
	}

	// Format the file with gofmt
	_ = exec.Command("gofmt", "-w", repoData.RepoFilePath).Run()
	return nil
}

// NewRepoMethod generates both the interface method signature and the
// implementation (function declaration) for a given repository method.
//
// It dynamically constructs the method name based on the provided Method
// enum and the entity name, following naming conventions such as:
// - CreateCar
// - ListCarsWithPagination
//
// The returned method struct contains:
// - Signature: to be inserted into the interface
// - Function:  to be appended to the implementation file
func NewRepoMethod(repoMethod Method, repoEntity, modelEntity string) method {
	methodName := generateRepoMethodName(repoMethod, modelEntity)

	modelType := strings.Split(cli_config.CliConfig.ModelsFolderPath, "/")[len(strings.Split(cli_config.CliConfig.ModelsFolderPath, "/"))-1] + "." + modelEntity

	return method{
		Signature: &ast.Field{
			Names: []*ast.Ident{ast.NewIdent(methodName)},
			Type: &ast.FuncType{
				Params: &ast.FieldList{
					List: generateMethodParams(repoMethod, modelEntity, modelType),
				},
				Results: &ast.FieldList{
					List: generateMethodResults(repoMethod, modelEntity, modelType),
				},
			},
		},
		Function: &ast.FuncDecl{
			Recv: &ast.FieldList{
				List: []*ast.Field{
					{
						Names: []*ast.Ident{ast.NewIdent("r")},
						Type:  &ast.StarExpr{X: ast.NewIdent(repoEntity + "Repo")},
					},
				},
			},
			Name: ast.NewIdent(methodName),
			Type: &ast.FuncType{
				Params: &ast.FieldList{
					List: generateMethodParams(repoMethod, modelEntity, modelType),
				},
				Results: &ast.FieldList{
					List: generateMethodResults(repoMethod, modelEntity, modelType),
				},
			},
			Body: &ast.BlockStmt{
				List: generateMethodBody(repoMethod, modelEntity, modelType),
			},
		},
	}
}

// generateRepoMethodName generates a method name based on the method type
// and the provided model entity.
//
// It handles pluralization for list-style methods and applies naming
// conventions such as:
// - "ListCars"
// - "ListCarsWithPagination"
// - "GetCarByUuid"
// - "CreateCar", etc.
func generateRepoMethodName(method Method, modelEntity string) string {
	switch method {
	case ListAll:
		return RepoRawMethodsMap[method] + inflection.Plural(modelEntity)
	case ListWithPagination:
		return "List" + inflection.Plural(modelEntity) + "WithPagination"
	case GetByUuid:
		return "Get" + modelEntity + "ByUuid"
	default:
		return RepoRawMethodsMap[method] + modelEntity
	}
}

// GenerateRepoMethodNamesMap returns a map of method name strings
// (e.g. "CreateCar") to their corresponding Method enum values.
//
// This is useful for lookup and display in selection prompts or
// when resolving user choices back to actual Method values.
func GenerateRepoMethodNamesMap(modelEntity string) map[string]Method {
	methodNames := map[string]Method{}
	for i := 0; i <= len(RepoRawMethodsMap); i++ {
		methodNames[generateRepoMethodName(Method(i), modelEntity)] = Method(i)
	}

	return methodNames
}

// GenerateSortedRepoMethodNames returns a sorted slice of method name
// strings based on the Method enum index.
//
// It is typically used to ensure consistent ordering when displaying
// available methods in the UI.
func GenerateSortedRepoMethodNames(modelEntity string) []string {
	var methodNames []string
	for i := 1; i <= len(RepoRawMethodsMap); i++ {
		methodNames = append(methodNames, generateRepoMethodName(Method(i), modelEntity))
	}

	return methodNames
}

// CreateRepo generates a new repository source file using a predefined template (repo.tmpl).
// It fills in the repo entity name and package path and writes the resulting code to disk.
func CreateRepo(repoData *RepoData) error {
	tmpl, err := template.ParseFS(templates.Files, RepoTemplatePath)
	if err != nil {
		return err
	}

	f, err := os.Create(repoData.RepoFilePath)
	if err != nil {
		return err
	}
	defer f.Close()

	templateData := struct {
		RepoPackage string
		RepoEntity  string
	}{
		RepoPackage: strings.Split(cli_config.CliConfig.RepositoriesFolderPath, "/")[len(strings.Split(cli_config.CliConfig.RepositoriesFolderPath, "/"))-1],
		RepoEntity:  repoData.RepoEntity,
	}

	err = tmpl.Execute(f, templateData)
	if err != nil {
		return err
	}
	return nil
}

func AddCentralRepoToCentralServiceConstructor() error {
	fset := token.NewFileSet()

	centralServiceFilePath := path.Join(cli_config.CliConfig.ServicesFolderPath, "central_service.go")

	node, err := parser.ParseFile(fset, centralServiceFilePath, nil, parser.AllErrors)
	if err != nil {
		return err
	}

	// Add import if not already present
	importPath := path.Join(cli_config.CliConfig.ProjectName, cli_config.CliConfig.RepositoriesFolderPath)
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
		if !ok || fn.Name.Name != "NewCentralService" {
			return true
		}

		// Check if centralService already exists
		for _, param := range fn.Type.Params.List {
			if len(param.Names) > 0 && param.Names[0].Name == "centralRepo" {
				return false
			}
		}

		// Add parameter: centralService *services.CentralService
		param := &ast.Field{
			Names: []*ast.Ident{ast.NewIdent("centralRepo")},
			Type: &ast.StarExpr{
				X: &ast.SelectorExpr{
					X:   ast.NewIdent("repos"),
					Sel: ast.NewIdent("CentralRepo"),
				},
			},
		}
		fn.Type.Params.List = append(fn.Type.Params.List, param)

		return false
	})

	// Write back to file
	file, err := os.Create(centralServiceFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	return printer.Fprint(file, fset, node)
}
