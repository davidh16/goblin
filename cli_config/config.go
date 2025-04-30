package cli_config

import (
	"fmt"
	"goblin/utils"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"text/template"
)

const configFilePath = ".goblin/cli_config.yaml"

type Config struct {
	ModelsFolderPath       string `yaml:"models_folder_path"`       // path for folder where model go files are located
	ControllersFolderPath  string `yaml:"controllers_folder_path"`  // path for folder where controller go files are located
	ServicesFolderPath     string `yaml:"services_folder_path"`     // path for folder where service go files are located
	RepositoriesFolderPath string `yaml:"repositories_folder_path"` // path for folder where repo go files are located
	ProjectName            string `yaml:"project_name"`
}

var CliConfig *Config

// CreateConfigFile generates a new CLI configuration file in the user's home directory.
//
// It creates the necessary directory (if not already present), loads a configuration template,
// and writes it to the config file. After creation, it reloads the configuration into global variable CliConfig.
//
// Returns an error if any file operations or template executions fail.
func CreateConfigFile() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	err = os.Mkdir(filepath.Join(home, ".goblin"), 0755)
	if err != nil {
		if !os.IsExist(err) {
			return err
		}
	}

	funcMap := template.FuncMap{
		"GetProjectName": utils.GetProjectName,
	}

	tmpl, err := template.New("cli_config.tmpl").Funcs(funcMap).ParseFiles("cli_config/cli_config.tmpl")
	if err != nil {
		return err
	}

	fmt.Println("xxx")

	configYamlFile, err := os.Create(filepath.Join(home, configFilePath))
	if err != nil {
		return err
	}

	err = tmpl.Execute(configYamlFile, nil)
	if err != nil {
		return err
	}

	err = LoadConfig()
	if err != nil {
		return err
	}

	return nil
}

// UpdateConfigFile takes a config represented as a map[string]string and writes it to the config file.
// It serializes the map to YAML format and writes it to the user's config file path.
//
// Returns an error if the file cannot be marshaled or written.
func UpdateConfigFile(cfg map[string]string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	updatedYAML, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	if err = os.WriteFile(filepath.Join(home, configFilePath), updatedYAML, 0644); err != nil {
		return err
	}

	return nil
}

// LoadConfigAsMap loads the config file from the user's home directory and unmarshals it into a map[string]string.
//
// Returns the config map and any error encountered during reading or parsing.
func LoadConfigAsMap() (map[string]string, error) {
	var cfg map[string]string

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(filepath.Join(home, configFilePath))
	if err != nil {
		return cfg, err
	}

	err = yaml.Unmarshal(data, &cfg)
	return cfg, err
}

// LoadConfig attempts to read the CLI configuration file from the user's home directory.
//
// If the configuration file does not exist, it creates a new one using a default template.
// It unmarshals the YAML data into the global CliConfig variable.
//
// Returns an error if reading, creating, or unmarshaling the config fails.
func LoadConfig() error {

	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	data, err := os.ReadFile(filepath.Join(home, configFilePath))
	if err != nil {
		if !os.IsNotExist(err) {
			utils.HandleError(err)
		}

		err = CreateConfigFile()
		if err != nil {
			utils.HandleError(err)
		}
	}

	return yaml.Unmarshal(data, &CliConfig)
}

// PrintConfigMap dynamically prints cli_config map as key: value pairs
func PrintConfigMap(cfgMap map[string]string) {
	fmt.Println("📦 CLI Config:")
	fmt.Println("────────────────────────────────────────────")
	for k, v := range cfgMap {
		fmt.Printf("%-30s %s\n", k+":", v)
	}
	fmt.Println("────────────────────────────────────────────")
}
