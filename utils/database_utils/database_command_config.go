package database_utils

var DatabaseOptionTemplatePaths = map[DatabaseOption]string{
	PostgresSQL: "postgres.tmpl",
	MariaDB:     "mariadb.tmpl",
	Redis:       "redis.tmpl",
}

const (
	PaginationTemplateFilePath = "pagination.tmpl"
)
