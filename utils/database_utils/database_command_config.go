package database_utils

var DatabaseOptionTemplatePaths = map[DatabaseOption]string{
	PostgresSQL: "commands/database/postgres.tmpl",
	MariaDB:     "commands/database/mariadb.tmpl",
	Redis:       "commands/database/redis.tmpl",
}

const (
	PaginationTemplateFilePath = "commands/database/pagination.tmpl"
)
