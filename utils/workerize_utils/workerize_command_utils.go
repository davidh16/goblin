package workerize_utils

import (
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
	data := &WorkerizeData{}

	data.JobsExists = utils.FileExists(path.Join(cli_config.CliConfig.JobsFolderPath, "job.go"))
	data.JobsManagerExists = utils.FileExists(path.Join(cli_config.CliConfig.JobsFolderPath, "jobs_manager.go"))
	data.OrchestratorExists = utils.FileExists(path.Join(cli_config.CliConfig.WorkersFolderPath, "orchestrator.go"))
	data.WorkerPoolExists = utils.FileExists(path.Join(cli_config.CliConfig.WorkersFolderPath, "worker_pool.go"))
	data.CentralServiceExists = utils.FileExists(path.Join(cli_config.CliConfig.ServicesFolderPath, "central_service.go"))
	return data
}

func ListImplementedDatabases() ([]string, error) {
	var implementedDatabases []string
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
		return nil, err
	}
	return implementedDatabases, nil
}
