package migration_utils

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
