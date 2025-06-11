package migration_utils

import (
	"github.com/davidh16/goblin/cli_config"
	"github.com/davidh16/goblin/utils"
	"os"
	"path"
	"text/template"
	"time"
)

type MigrationData struct {
	MigrationNameSnakeCase    string
	MigrationUpFileName       string
	MigrationDownFileName     string
	MigrationUpFileFullPath   string
	MigrationDownFileFullPath string
}

func NewMigrationData() *MigrationData {
	return &MigrationData{}
}

func GenerateMigrationDataFromName(name string) *MigrationData {

	migrationData := NewMigrationData()
	migrationData.MigrationNameSnakeCase = name
	migrationData.MigrationUpFileName = time.Now().Format("20060102150405") + "_" + name + "_up.sql"
	migrationData.MigrationDownFileName = time.Now().Format("20060102150405") + "_" + name + "_down.sql"
	migrationData.MigrationUpFileFullPath = path.Join(cli_config.CliConfig.MigrationsFolderPath, migrationData.MigrationUpFileName)
	migrationData.MigrationDownFileFullPath = path.Join(cli_config.CliConfig.MigrationsFolderPath, migrationData.MigrationDownFileName)

	return migrationData
}

func GenerateMigrationFiles(migrationData *MigrationData) error {

	if exists := utils.FileExists(cli_config.CliConfig.MigrationsFolderPath); !exists {
		err := os.MkdirAll(cli_config.CliConfig.MigrationsFolderPath, 0755) // 0755 = rwxr-xr-x
		if err != nil {
			return err
		}
	}

	if exists := utils.FileExists(path.Join(cli_config.CliConfig.MigrationsFolderPath, "uuid_ossp_up.sql")); !exists {
		err := createUuidOsspMigrations()
		if err != nil {
			return err
		}
	}

	f, err := os.Create(migrationData.MigrationUpFileFullPath)
	if err != nil {
		return err
	}

	tmpl, err := template.ParseFiles(MigrationUpTemplatePath)
	if err != nil {
		return err
	}

	templateData := struct {
		TableName string
	}{
		TableName: migrationData.MigrationNameSnakeCase,
	}

	err = tmpl.Execute(f, templateData)
	if err != nil {
		return err
	}

	f, err = os.Create(migrationData.MigrationDownFileFullPath)
	if err != nil {
		return err
	}
	defer f.Close()

	tmpl, err = template.ParseFiles(MigrationDownTemplatePath)
	if err != nil {
		return err
	}

	err = tmpl.Execute(f, templateData)
	if err != nil {
		return err
	}

	return nil
}

func createUuidOsspMigrations() error {
	f, err := os.Create(path.Join(cli_config.CliConfig.MigrationsFolderPath, "uuid_ossp_up.sql"))
	if err != nil {
		return err
	}
	defer f.Close()

	tmpl, err := template.ParseFiles(UuidOsspUpTemplatePath)
	if err != nil {
		return err
	}

	err = tmpl.Execute(f, nil)
	if err != nil {
		return err
	}

	f, err = os.Create(path.Join(cli_config.CliConfig.MigrationsFolderPath, "uuid_ossp_down.sql"))
	if err != nil {
		return err
	}
	defer f.Close()

	tmpl, err = template.ParseFiles(UuidOsspDownTemplatePath)
	if err != nil {
		return err
	}

	err = tmpl.Execute(f, nil)
	if err != nil {
		return err
	}

	return nil
}
