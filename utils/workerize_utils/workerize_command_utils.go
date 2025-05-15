package workerize_utils

import (
	"go/ast"
	"go/parser"
	"go/token"
	"goblin/cli_config"
	"goblin/utils/database_utils"
	"os"
	"path/filepath"
	"strings"
)

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
