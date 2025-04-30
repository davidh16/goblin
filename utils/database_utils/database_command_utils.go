package database_utils

import (
	"errors"
	"fmt"
	"goblin/cli_config"
	"goblin/utils"
	"os"
	"path"
	"strings"
	"text/template"
)

type DatabaseOption int

const (
	Unspecified DatabaseOption = iota
	PostgresSQL
	MariaDB
	Redis
)

var DatabaseNameOptionsMap = map[string]DatabaseOption{
	"PostgresSQL": PostgresSQL,
	"MariaDB":     MariaDB,
	"Redis":       Redis,
}

var DatabaseOptionNamesMap = map[DatabaseOption]string{
	Unspecified: "Unspecified",
	PostgresSQL: "PostgresSQL",
	MariaDB:     "MariaDB",
	Redis:       "Redis",
}

var DatabaseOptionDefaultPortsMap = map[DatabaseOption]string{
	PostgresSQL: "5432",
	MariaDB:     "3306",
	Redis:       "6379",
}

var DatabaseOptionInstanceDefaultFileNamesMap = map[DatabaseOption]string{
	PostgresSQL: "postgres.go",
	MariaDB:     "mariadb.go",
	Redis:       "redis.go",
}

func GetDatabaseOptionDefaultEnvDataMap(option DatabaseOption) (map[string]string, error) {
	switch option {
	case Unspecified:
		return nil, errors.New("unspecified database")
	case PostgresSQL:
		return map[string]string{
			"POSTGRES_USER":     "postgres",
			"POSTGRES_PASSWORD": "postgres",
			"POSTGRES_DB":       cli_config.CliConfig.ProjectName,
			"POSTGRES_HOST":     "localhost",
			"POSTGRES_PORT":     DatabaseOptionDefaultPortsMap[PostgresSQL],
		}, nil
	case MariaDB:
		return map[string]string{
			"MARIADB_USER":     "mariadb",
			"MARIADB_PASSWORD": "mariadb",
			"MARIADB_DB":       cli_config.CliConfig.ProjectName,
			"MARIADB_HOST":     "localhost",
			"MARIADB_PORT":     DatabaseOptionDefaultPortsMap[MariaDB],
		}, nil
	case Redis:
		return map[string]string{
			"REDIS_USER":     "redis",
			"REDIS_PASSWORD": "redis",
			"REDIS_DB":       cli_config.CliConfig.ProjectName,
			"REDIS_HOST":     "localhost",
			"REDIS_PORT":     DatabaseOptionDefaultPortsMap[Redis],
		}, nil
	default:
		return nil, errors.New("unknown option")
	}
}

type DatabaseData struct {
	DatabaseType DatabaseOption
	Port         string
}

func InitializeDatabaseInstance(database DatabaseData) error {

	tmpl, err := template.ParseFiles(DatabaseOptionTemplatePaths[database.DatabaseType])
	if err != nil {
		utils.HandleError(err, "Error parsing model template")
	}

	err = os.MkdirAll(cli_config.CliConfig.DatabaseInstancesFolderPath, 0755)
	if err != nil {
		if !os.IsExist(err) {
			return err
		}

	}

	file, err := os.Create(path.Join(cli_config.CliConfig.DatabaseInstancesFolderPath, DatabaseOptionInstanceDefaultFileNamesMap[database.DatabaseType]))
	if err != nil {
		if !os.IsExist(err) {
			return err
		}
	}
	defer file.Close()

	templateData := struct {
		DatabasePackage string
	}{
		DatabasePackage: strings.Split(cli_config.CliConfig.DatabaseInstancesFolderPath, "/")[len(strings.Split(cli_config.CliConfig.DatabaseInstancesFolderPath, "/"))-1],
	}

	err = tmpl.Execute(file, templateData)
	if err != nil {
		return err
	}

	if database.DatabaseType != Redis {
		err = generatePaginationFile()
		if err != nil {
			return err
		}
	}

	fmt.Println(fmt.Sprintf("✅ %s instance initialized.", DatabaseOptionNamesMap[database.DatabaseType]))
	return nil
}

func generatePaginationFile() error {
	// if pagination already exists, return early
	file, err := os.Create(path.Join(cli_config.CliConfig.DatabaseInstancesFolderPath, "pagination.go"))
	if err != nil {
		if !os.IsExist(err) {
			return err
		}
		return nil
	}
	defer file.Close()

	tmpl, err := template.ParseFiles(PaginationTemplateFilePath)
	if err != nil {
		utils.HandleError(err, "Error parsing model template")
	}

	templateData := struct {
		DatabasePackage string
	}{
		DatabasePackage: strings.Split(cli_config.CliConfig.DatabaseInstancesFolderPath, "/")[len(strings.Split(cli_config.CliConfig.DatabaseInstancesFolderPath, "/"))-1],
	}

	err = tmpl.Execute(file, templateData)
	if err != nil {
		utils.HandleError(err, "Error executing model template")
	}

	fmt.Println("✅ pagination.go file generated.")
	return nil
}
