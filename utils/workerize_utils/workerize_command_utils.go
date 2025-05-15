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
	JobNameSnakeCase    string
	JobNamePascalCase   string
	JobNameCamelCase    string
	JobTypeName         string
	JobFilePath         string
	JobFileName         string
	JobMetadataFileName string
	JobMetadataName     string
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
		JobsPackage string
	}{
		JobsPackage: strings.Split(cli_config.CliConfig.JobsFolderPath, "/")[len(strings.Split(cli_config.CliConfig.JobsFolderPath, "/"))-1],
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

	// Step 1: Add new JobType constant
	newConst := &ast.ValueSpec{
		Names: []*ast.Ident{ast.NewIdent(customJobData.JobNamePascalCase)},
		Values: []ast.Expr{&ast.BinaryExpr{
			X:  ast.NewIdent("iota"),
			Op: token.ADD,
			Y:  ast.NewIdent("1"), // Assumes appending to previous iota block
		}},
	}

	// Find const declaration of JobType and append
	for _, decl := range node.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.CONST {
			continue
		}
		for _, spec := range genDecl.Specs {
			valSpec, ok := spec.(*ast.ValueSpec)
			if ok && valSpec.Names[0].Name == "JobTypeUnspecified" {
				genDecl.Specs = append(genDecl.Specs, newConst)
				break
			}
		}
	}

	ast.Inspect(node, func(n ast.Node) bool {
		compositeLit, ok := n.(*ast.CompositeLit)
		if !ok {
			return true
		}

		switch x := compositeLit.Type.(type) {
		case *ast.MapType:
			// jobTypesMap
			_, ok := x.Key.(*ast.Ident)
			if ok {
				_, ok := x.Value.(*ast.StructType)
				if ok {
					compositeLit.Elts = append(compositeLit.Elts, &ast.KeyValueExpr{
						Key:   ast.NewIdent(customJobData.JobTypeName),
						Value: &ast.CompositeLit{Type: ast.NewIdent("struct{}")},
					})
				}
				// jobTypeMetadataMap
				_, ok = x.Key.(*ast.Ident)
				if ok {

					_, ok := x.Value.(*ast.Ident)
					if ok {
						compositeLit.Elts = append(compositeLit.Elts, &ast.KeyValueExpr{
							Key: ast.NewIdent(customJobData.JobTypeName),
							Value: &ast.CallExpr{
								Fun:  ast.NewIdent("reflect.TypeOf"),
								Args: []ast.Expr{&ast.CompositeLit{Type: ast.NewIdent(customJobData.JobMetadataName)}},
							},
						})
					}
				}
				return true
			}
		}
		return true
	})

	// Write back to file
	file, err := os.Create("job.go")
	if err != nil {
		return err
	}
	defer file.Close()

	if err = printer.Fprint(file, fset, node); err != nil {
		return err
	}

	return err
}
