package workerize_utils

import (
	"errors"
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"goblin/cli_config"
	"goblin/commands/database"
	central_service "goblin/commands/service/flags/central-service"
	"goblin/utils"
	"goblin/utils/database_utils"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"
)

type WorkerizeData struct {
	JobsExists            bool
	JobsOverwrite         bool
	JobsManagerExists     bool
	JobsManagerOverwrite  bool
	OrchestratorExists    bool
	OrchestratorOverwrite bool
	WorkerPoolExists      bool
	WorkerPoolOverwrite   bool
	CentralServiceExists  bool
}

type CustomJobData struct {
	JobNameSnakeCase          string
	JobNamePascalCase         string
	JobNameCamelCase          string
	JobTypeName               string
	JobFilePath               string
	JobFileName               string
	JobMetadataFileName       string
	JobMetadataName           string
	AlreadyExists             bool
	CreateWorkerPool          bool
	WorkerPoolNameSnakeCase   string
	WorkerPoolNamePascalCase  string
	WorkerPoolNameCamelCase   string
	WorkerPoolFileName        string
	WorkerPoolOverwrite       bool
	WorkerPoolExists          bool
	WorkerPoolSize            int
	WorkerPoolNumberOfRetries int
	WorkerName                string
	ServicesToImplement       []string
}

func InitBoilerplateWorkerizeData() *WorkerizeData {
	data := &WorkerizeData{
		JobsOverwrite:         true,
		JobsManagerOverwrite:  true,
		OrchestratorOverwrite: true,
		WorkerPoolOverwrite:   true,
	}

	data.JobsExists = utils.FileExists(path.Join(cli_config.CliConfig.JobsFolderPath, "job.go"))
	data.JobsManagerExists = utils.FileExists(path.Join(cli_config.CliConfig.JobsFolderPath, "jobs_manager.go"))
	data.OrchestratorExists = utils.FileExists(path.Join(cli_config.CliConfig.WorkersFolderPath, "orchestrator.go"))
	data.WorkerPoolExists = utils.FileExists(path.Join(cli_config.CliConfig.WorkersFolderPath, "worker_pool.go"))
	data.CentralServiceExists = utils.FileExists(path.Join(cli_config.CliConfig.ServicesFolderPath, "central_service.go"))
	return data
}

func ListImplementedDatabases() ([]string, error) {
	var implementedDatabases []string
	err := filepath.WalkDir(cli_config.CliConfig.DatabaseInstancesFolderPath, func(repoPath string, d os.DirEntry, err error) error {
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
			if !ok || funcDecl.Name == nil {
				continue
			}

			funcName := funcDecl.Name.Name

			switch {
			case strings.HasPrefix(funcName, "ConnectToMariaDB"):
				implementedDatabases = append(implementedDatabases, database_utils.DatabaseOptionNamesMap[database_utils.MariaDB])
			case strings.HasPrefix(funcName, "ConnectToPostgres"):
				implementedDatabases = append(implementedDatabases, database_utils.DatabaseOptionNamesMap[database_utils.PostgresSQL])
			case strings.HasPrefix(funcName, "ConnectToRedis"):
				implementedDatabases = append(implementedDatabases, database_utils.DatabaseOptionNamesMap[database_utils.Redis])
			}
		}

		return nil
	})
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}
	return implementedDatabases, nil
}

func ImplementJobsLogic(data *WorkerizeData) error {
	if data.JobsOverwrite {

		tmpl, err := template.ParseFiles(JobTemplateFilePath)
		if err != nil {
			return err
		}

		jobPath := path.Join(cli_config.CliConfig.JobsFolderPath, "job.go")

		f, err := os.Create(jobPath)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				err = os.Mkdir(cli_config.CliConfig.JobsFolderPath, 0755) // 0755 = rwxr-xr-x
				if err != nil {
					return err
				}
				f, err = os.Create(jobPath)
				if err != nil {
					return err
				}
			}
		}
		defer f.Close()

		templateData := struct {
			JobsPackage string
		}{
			JobsPackage: strings.Split(cli_config.CliConfig.JobsFolderPath, "/")[len(strings.Split(cli_config.CliConfig.JobsFolderPath, "/"))-1],
		}

		err = tmpl.Execute(f, templateData)
		if err != nil {
			return err
		}

		fmt.Println("✅ Jobs logic generated successfully.")
	}

	if data.JobsManagerOverwrite {
		tmpl, err := template.ParseFiles(JobsManagerTemplateFilePath)
		if err != nil {
			utils.HandleError(err)
		}

		jobPath := path.Join(cli_config.CliConfig.JobsFolderPath, "jobs_manager.go")

		f, err := os.Create(jobPath)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				err = os.Mkdir(cli_config.CliConfig.JobsFolderPath, 0755) // 0755 = rwxr-xr-x
				if err != nil {
					return err
				}
				f, err = os.Create(jobPath)
				if err != nil {
					return err
				}
			}
		}
		defer f.Close()

		templateData := struct {
			JobsPackage string
		}{
			JobsPackage: strings.Split(cli_config.CliConfig.JobsFolderPath, "/")[len(strings.Split(cli_config.CliConfig.JobsFolderPath, "/"))-1],
		}

		err = tmpl.Execute(f, templateData)
		if err != nil {
			return err
		}

		fmt.Println("✅ Jobs manager logic generated successfully.")

	}

	return nil
}

func ImplementWorkersLogic(data *WorkerizeData) error {
	if data.WorkerPoolOverwrite {

		tmpl, err := template.ParseFiles(WorkerPoolTemplateFilePath)
		if err != nil {
			return err
		}

		jobPath := path.Join(cli_config.CliConfig.WorkersFolderPath, "worker_pool.go")

		f, err := os.Create(jobPath)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				err = os.Mkdir(cli_config.CliConfig.WorkersFolderPath, 0755) // 0755 = rwxr-xr-x
				if err != nil {
					return err
				}
				f, err = os.Create(jobPath)
				if err != nil {
					return err
				}
			}
		}
		defer f.Close()

		templateData := struct {
			WorkersPackage string
			JobsPackage    string
		}{
			WorkersPackage: strings.Split(cli_config.CliConfig.WorkersFolderPath, "/")[len(strings.Split(cli_config.CliConfig.WorkersFolderPath, "/"))-1],
			JobsPackage:    cli_config.CliConfig.JobsFolderPath,
		}

		err = tmpl.Execute(f, templateData)
		if err != nil {
			return err
		}

		fmt.Println("✅ Worker pool logic generated successfully.")
	}

	if data.OrchestratorOverwrite {
		tmpl, err := template.ParseFiles(OrchestratorTemplateFilePath)
		if err != nil {
			utils.HandleError(err)
		}

		jobPath := path.Join(cli_config.CliConfig.WorkersFolderPath, "orchestrator.go")

		f, err := os.Create(jobPath)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				err = os.Mkdir(cli_config.CliConfig.WorkersFolderPath, 0755) // 0755 = rwxr-xr-x
				if err != nil {
					return err
				}
				f, err = os.Create(jobPath)
				if err != nil {
					return err
				}
			}
		}
		defer f.Close()

		templateData := struct {
			WorkersPackage string
			JobsPackage    string
			LoggerPackage  string
		}{
			WorkersPackage: strings.Split(cli_config.CliConfig.WorkersFolderPath, "/")[len(strings.Split(cli_config.CliConfig.WorkersFolderPath, "/"))-1],
			JobsPackage:    cli_config.CliConfig.JobsFolderPath,
			LoggerPackage:  cli_config.CliConfig.LoggerFolderPath,
		}

		err = tmpl.Execute(f, templateData)
		if err != nil {
			return err
		}

		fmt.Println("✅ Orchestrator worker logic generated successfully.")

	}

	return nil
}

func IfWorkerizeIsInitialized() bool {
	exists := utils.FileExists(path.Join(cli_config.CliConfig.JobsFolderPath, "job.go"))
	if !exists {
		return false
	}

	exists = utils.FileExists(path.Join(cli_config.CliConfig.WorkersFolderPath, "orchestrator.go"))
	if !exists {
		return false
	}

	exists = utils.FileExists(path.Join(cli_config.CliConfig.WorkersFolderPath, "worker_pool.go"))
	if !exists {
		return false
	}
	return true
}

func WorkerizeCmdHandlerCopy() {
	implementedDatabases, err := ListImplementedDatabases()
	if err != nil {
		utils.HandleError(err, "Unable to list implemented databases")
	}

	if len(implementedDatabases) > 1 {
		var redisImplemented bool
		for _, impl := range implementedDatabases {
			if impl == database_utils.DatabaseOptionNamesMap[database_utils.Redis] {
				redisImplemented = true
				break
			}
		}

		if !redisImplemented {
			var confirmContinue bool
			confirmContinuePrompt := &survey.Confirm{
				Message: "For implementing background jobs and workers, Redis needs to be implemented, do you wish to continue with Redis implementation?",
				Default: false,
			}
			err = survey.AskOne(confirmContinuePrompt, &confirmContinue)
			if err != nil {
				utils.HandleError(err)
			}
			if !confirmContinue {
				return
			}

			err = database.ImplementRedis()
			if err != nil {
				utils.HandleError(err)
			}
		}

	} else {

		var confirmContinue bool
		confirmContinuePrompt := &survey.Confirm{
			Message: "For implementing background jobs and workers, one persistent database and Redis need to be implemented, do you wish to continue with database implementations?",
			Default: false,
		}
		err = survey.AskOne(confirmContinuePrompt, &confirmContinue)
		if err != nil {
			utils.HandleError(err)
		}
		if !confirmContinue {
			return
		}
		err = database.ImplementRedisAndOtherGormDb()
		if err != nil {
			utils.HandleError(err)
		}
	}

	data := InitBoilerplateWorkerizeData()

	if data.JobsExists {
		confirmOverwritePrompt := &survey.Confirm{
			Message: "job.go already exists, do you wish to overwrite?",
			Default: false,
		}
		err = survey.AskOne(confirmOverwritePrompt, &data.JobsOverwrite)
		if err != nil {
			utils.HandleError(err)
		}
	}

	if data.JobsManagerExists {
		confirmOverwritePrompt := &survey.Confirm{
			Message: "jobs_manager.go already exists, do you wish to overwrite?",
			Default: false,
		}
		err = survey.AskOne(confirmOverwritePrompt, &data.JobsManagerOverwrite)
		if err != nil {
			utils.HandleError(err)
		}
	}

	if !data.CentralServiceExists {
		central_service.GenerateCentralService()
	}

	if data.WorkerPoolExists {
		confirmOverwritePrompt := &survey.Confirm{
			Message: "worker_pool.go already exists, do you wish to overwrite?",
			Default: false,
		}
		err = survey.AskOne(confirmOverwritePrompt, &data.WorkerPoolOverwrite)
		if err != nil {
			utils.HandleError(err)
		}
	}

	if data.OrchestratorExists {
		confirmOverwritePrompt := &survey.Confirm{
			Message: "orchestrator.go already exists, do you wish to overwrite?",
			Default: false,
		}
		err = survey.AskOne(confirmOverwritePrompt, &data.OrchestratorOverwrite)
		if err != nil {
			utils.HandleError(err)
		}
	}

	err = ImplementJobsLogic(data)
	if err != nil {
		utils.HandleError(err)
	}

	err = ImplementWorkersLogic(data)
	if err != nil {
		utils.HandleError(err)
	}

	return
}

func GenerateCustomJobMetadataFile(customJobData *CustomJobData) error {
	tmpl, err := template.ParseFiles(CustomJobMetadataTemplateFilePath)
	if err != nil {
		return err
	}

	jobMetadataPath := path.Join(cli_config.CliConfig.JobsFolderPath, customJobData.JobMetadataFileName)

	f, err := os.Create(jobMetadataPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err = os.Mkdir(cli_config.CliConfig.JobsFolderPath, 0755) // 0755 = rwxr-xr-x
			if err != nil {
				return err
			}
			f, err = os.Create(jobMetadataPath)
			if err != nil {
				return err
			}
		}
	}
	defer f.Close()

	templateData := struct {
		JobsPackage     string
		JobMetadataName string
	}{
		JobsPackage:     strings.Split(cli_config.CliConfig.JobsFolderPath, "/")[len(strings.Split(cli_config.CliConfig.JobsFolderPath, "/"))-1],
		JobMetadataName: customJobData.JobMetadataName,
	}

	err = tmpl.Execute(f, templateData)
	if err != nil {
		return err
	}

	fmt.Println(fmt.Sprintf("✅ %s generated successfully.", customJobData.JobMetadataFileName))
	return nil
}

func AddCustomJobToBaseJob(customJobData *CustomJobData) error {
	baseJobFilePath := path.Join(cli_config.CliConfig.JobsFolderPath, "job.go")

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, baseJobFilePath, nil, parser.AllErrors)
	if err != nil {
		return err
	}

	// Track if const already exists
	jobTypeExists := false

	// Step 1: Find const declaration of JobType and conditionally append
	for _, decl := range node.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.CONST {
			continue
		}
		for _, spec := range genDecl.Specs {
			valSpec, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			for _, ident := range valSpec.Names {
				if ident.Name == customJobData.JobTypeName {
					jobTypeExists = true
					break
				}
			}
		}

		if !jobTypeExists {
			for _, spec := range genDecl.Specs {
				valSpec, ok := spec.(*ast.ValueSpec)
				if ok && valSpec.Names[0].Name == "JobTypeUnspecified" {
					newConst := &ast.ValueSpec{
						Names: []*ast.Ident{ast.NewIdent(customJobData.JobTypeName)},
					}
					genDecl.Specs = append(genDecl.Specs, newConst)
					break
				}
			}
		}
	}

	// Track seen keys to avoid duplicate additions
	jobTypesMapSeen := map[string]bool{}
	jobTypeMetadataMapSeen := map[string]bool{}

	// Step 2: Update jobTypesMap and jobTypeMetadataMap
	ast.Inspect(node, func(n ast.Node) bool {
		genDecl, ok := n.(*ast.GenDecl)
		if !ok {
			return true
		}

		for _, spec := range genDecl.Specs {
			valSpec, ok := spec.(*ast.ValueSpec)
			if !ok || len(valSpec.Names) == 0 || len(valSpec.Values) == 0 {
				continue
			}

			name := valSpec.Names[0].Name
			cl, ok := valSpec.Values[0].(*ast.CompositeLit)
			if !ok {
				continue
			}

			mapType, ok := cl.Type.(*ast.MapType)
			if !ok {
				continue
			}

			keyIdent, ok := mapType.Key.(*ast.Ident)
			if !ok || keyIdent.Name != "JobType" {
				continue
			}

			// Track existing keys
			for _, elt := range cl.Elts {
				kv, ok := elt.(*ast.KeyValueExpr)
				if !ok {
					continue
				}
				keyIdent, ok := kv.Key.(*ast.Ident)
				if !ok {
					continue
				}
				if name == "jobTypesMap" {
					jobTypesMapSeen[keyIdent.Name] = true
				} else if name == "jobTypeMetadataMap" {
					jobTypeMetadataMapSeen[keyIdent.Name] = true
				}
			}

			switch name {
			case "jobTypesMap":
				if !jobTypesMapSeen[customJobData.JobTypeName] {
					cl.Elts = append(cl.Elts, &ast.KeyValueExpr{
						Key:   ast.NewIdent(customJobData.JobTypeName),
						Value: &ast.CompositeLit{Type: ast.NewIdent("struct{}")},
					})
				}
			case "jobTypeMetadataMap":
				if !jobTypeMetadataMapSeen[customJobData.JobTypeName] {
					cl.Elts = append(cl.Elts, &ast.KeyValueExpr{
						Key: ast.NewIdent(customJobData.JobTypeName),
						Value: &ast.CallExpr{
							Fun: ast.NewIdent("reflect.TypeOf"),
							Args: []ast.Expr{&ast.CompositeLit{
								Type: ast.NewIdent(customJobData.JobMetadataName),
							}},
						},
					})
				}
			}
		}
		return true
	})

	// Step 3: Write changes back
	file, err := os.Create(baseJobFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	return printer.Fprint(file, fset, node)
}

func GenerateCustomWorkerPool(customJobData *CustomJobData) error {
	funcMap := template.FuncMap{
		"GenerateWorkerStructFields": GenerateWorkerStructFields,
		"GenerateImplementations":    GenerateImplementations,
	}

	tmpl, err := template.New(CustomWorkerPoolTemplateName).Funcs(funcMap).ParseFiles(CustomWorkerPoolTemplateFilePath)
	if err != nil {
		return err
	}

	customWorkerPoolPath := path.Join(cli_config.CliConfig.WorkersFolderPath, customJobData.WorkerPoolFileName)

	f, err := os.Create(customWorkerPoolPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err = os.Mkdir(cli_config.CliConfig.WorkersFolderPath, 0755) // 0755 = rwxr-xr-x
			if err != nil {
				return err
			}
			f, err = os.Create(customWorkerPoolPath)
			if err != nil {
				return err
			}
		}
	}
	defer f.Close()

	templateData := struct {
		WorkersPackage        string
		JobsPackageImport     string
		ServicePackageImport  string
		ServicePackage        string
		LoggerPackageImport   string
		LoggerPackage         string
		WorkerPoolName        string
		WorkerName            string
		WorkerPoolSize        int
		NumberOfRetries       int
		CustomJobMetadataName string
		ServicesToImplement   []string
	}{
		WorkersPackage:        strings.Split(cli_config.CliConfig.WorkersFolderPath, "/")[len(strings.Split(cli_config.CliConfig.WorkersFolderPath, "/"))-1],
		JobsPackageImport:     path.Join(cli_config.CliConfig.ProjectName, cli_config.CliConfig.JobsFolderPath),
		ServicePackageImport:  path.Join(cli_config.CliConfig.ProjectName, cli_config.CliConfig.ServicesFolderPath),
		ServicePackage:        strings.Split(cli_config.CliConfig.ServicesFolderPath, "/")[len(strings.Split(cli_config.CliConfig.ServicesFolderPath, "/"))-1],
		LoggerPackageImport:   path.Join(cli_config.CliConfig.ProjectName, cli_config.CliConfig.LoggerFolderPath),
		LoggerPackage:         strings.Split(cli_config.CliConfig.LoggerFolderPath, "/")[len(strings.Split(cli_config.CliConfig.LoggerFolderPath, "/"))-1],
		WorkerPoolName:        customJobData.WorkerPoolNamePascalCase,
		WorkerName:            customJobData.WorkerName,
		WorkerPoolSize:        customJobData.WorkerPoolSize,
		NumberOfRetries:       customJobData.WorkerPoolNumberOfRetries,
		CustomJobMetadataName: customJobData.JobMetadataName,
		ServicesToImplement:   customJobData.ServicesToImplement,
	}

	err = tmpl.Execute(f, templateData)
	if err != nil {
		return err
	}

	fmt.Println(fmt.Sprintf("✅ %s generated successfully.", customJobData.WorkerPoolFileName))
	return nil
}

func GenerateImplementations(services []string) string {
	if len(services) == 0 {
		return ""
	}

	var builder strings.Builder
	for _, service := range services {
		fmt.Fprintf(&builder, "%s: centralService.%s,\n", utils.PascalToCamel(service), service)
	}
	return builder.String()
}

func GenerateWorkerStructFields(services []string) string {
	if len(services) == 0 {
		return ""
	}

	var builder strings.Builder
	for _, service := range services {
		fmt.Fprintf(&builder, "%s %s.%sInterface\n", utils.PascalToCamel(service), strings.Split(cli_config.CliConfig.ServicesFolderPath, "/")[len(strings.Split(cli_config.CliConfig.ServicesFolderPath, "/"))-1], service)
	}
	return builder.String()
}
