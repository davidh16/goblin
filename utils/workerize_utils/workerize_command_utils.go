package workerize_utils

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"goblin/cli_config"
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
