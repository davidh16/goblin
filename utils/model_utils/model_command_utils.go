package model_utils

import (
	"fmt"
	"github.com/AlecAivazis/survey/v2"
	"goblin/cli_config"
	"goblin/utils"
	"goblin/utils/migration_utils"
	"os"
	"path"
	"reflect"
	"strings"
	"text/template"
	"time"
)

type ModelData struct {
	NameSnakeCase string
	ModelFileName string
	ModelFilePath string
	ModelEntity   string
}

type MigrationColumn struct {
	Name         string
	SQLType      string
	Nullable     bool
	IsPrimaryKey bool
	IsUnique     bool
	HasDefault   bool
	DefaultExpr  string // e.g. "uuid_generate_v4()"
}

func NewModelData() *ModelData {
	return &ModelData{}
}

func TriggerGetModelNameFlow() (*ModelData, error) {
	modelData := &ModelData{}

	for {
		for {
			if err := survey.AskOne(&survey.Input{
				Message: "Please type the model file name (snake_case) :",
				Default: "my_model_file",
			}, &modelData.NameSnakeCase); err != nil {
				return nil, err
			}

			if !utils.IsSnakeCase(modelData.NameSnakeCase) {
				fmt.Printf("🛑 %s is not in snake case\n", modelData.NameSnakeCase)
			} else {
				break
			}
		}

		modelData.ModelEntity = utils.SnakeToPascal(modelData.NameSnakeCase)
		modelData.ModelFileName = modelData.NameSnakeCase + ".go"
		modelData.ModelFilePath = path.Join(cli_config.CliConfig.ModelsFolderPath, modelData.ModelFileName)

		var confirm bool
		confirmPrompt := &survey.Confirm{
			Message: fmt.Sprintf("You are about to create a model file named %s, do you want to continue ?", modelData.ModelFileName),
		}
		if err := survey.AskOne(confirmPrompt, &confirm); err != nil {
			return nil, err
		}

		if !confirm {
			continue
		}

		for {
			if utils.FileExists(modelData.ModelFilePath) {
				var overwriteConfirmed bool
				confirmPrompt = &survey.Confirm{
					Message: fmt.Sprintf("%s model already exists. Do you want to overwrite it ?", modelData.ModelFileName),
					Default: false,
				}
				if err := survey.AskOne(confirmPrompt, &overwriteConfirmed); err != nil {
					return nil, err
				}

				if overwriteConfirmed {
					confirmPrompt = &survey.Confirm{
						Message: fmt.Sprintf("Are you sure you want to overwrite %s model ?", modelData.ModelFileName),
						Default: false,
					}
					if err := survey.AskOne(confirmPrompt, &overwriteConfirmed); err != nil {
						return nil, err
					}
				}

				if !overwriteConfirmed {
					continue
				}
			}
			break
		}
		break
	}
	return modelData, nil
}

func CreateMigrationForUserModel(selectedAttributes []string) error {
	var columnDefs []MigrationColumn
	for _, name := range selectedAttributes {
		typ, ok := AllPossibleUserModelAttributes[name]
		if !ok {
			continue
		}

		sqlType := mapGoTypeToSQL(typ)
		nullable := typ.Kind() == reflect.Ptr

		column := MigrationColumn{
			Name:     strings.ToLower(utils.PascalToSnake(name)),
			SQLType:  sqlType,
			Nullable: nullable,
		}

		if name == "Uuid" {
			column.HasDefault = true
			column.DefaultExpr = "uuid_generate_v4()"
			column.Nullable = false
			column.IsUnique = true
			column.IsPrimaryKey = true
		}

		columnDefs = append(columnDefs, column)
	}

	usersUpMigrationPath := path.Join(cli_config.CliConfig.MigrationsFolderPath, time.Now().Format("20060102150405")+"_users_up.sql")
	usersDownMigrationPath := path.Join(cli_config.CliConfig.MigrationsFolderPath, time.Now().Format("20060102150405")+"_users_down.sql")

	f, err := os.Create(usersUpMigrationPath)
	if err != nil {
		return err
	}

	funcMap := template.FuncMap{
		"add": func(a, b int) int {
			return a + b
		},
		"len": func(a []MigrationColumn) int {
			return len(a)
		},
	}
	templateData := struct {
		MigrationColumns []MigrationColumn
		TableName        string
	}{
		MigrationColumns: columnDefs,
		TableName:        "users",
	}
	tmpl := template.Must(template.New(migration_utils.CustomMigrationUpTemplateName).Funcs(funcMap).ParseFiles(migration_utils.CustomMigrationUpTemplatePath))

	err = tmpl.Execute(f, templateData)
	if err != nil {
		return err
	}

	f, err = os.Create(usersDownMigrationPath)
	if err != nil {
		return err
	}

	tmpl, err = template.ParseFiles(migration_utils.CustomMigrationDownTemplatePath)
	if err != nil {
		return err
	}

	err = tmpl.Execute(f, templateData)
	if err != nil {
		return err
	}

	fmt.Println("✅ Users migration generated successfully.")

	return nil
}

func mapGoTypeToSQL(t reflect.Type) string {
	// Handle pointer types
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	switch t.Kind() {

	case reflect.String:
		return "TEXT"
	case reflect.Int, reflect.Int64:
		return "BIGINT"
	case reflect.Int32:
		return "INTEGER"
	case reflect.Int16:
		return "SMALLINT"
	case reflect.Int8:
		return "SMALLINT"
	case reflect.Uint, reflect.Uint64:
		return "BIGINT" // no unsigned in Postgres
	case reflect.Uint32:
		return "INTEGER"
	case reflect.Uint16, reflect.Uint8:
		return "SMALLINT"
	case reflect.Float32:
		return "REAL"
	case reflect.Float64:
		return "DOUBLE PRECISION"
	case reflect.Bool:
		return "BOOLEAN"
	}

	// Special case for time.Time
	if t.PkgPath() == "time" && t.Name() == "Time" {
		return "TIMESTAMP"
	}

	// Common case for UUID (e.g., from github.com/google/uuid)
	if t.Name() == "UUID" {
		return "UUID"
	}

	// Fallback for unknown types
	return "TEXT"
}
