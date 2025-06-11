package database_utils

var DatabaseOptionTemplatePaths = map[DatabaseOption]string{
	PostgresSQL: "templates/postgres.tmpl",
	MariaDB:     "templates/mariadb.tmpl",
	Redis:       "templates/redis.tmpl",
}

const (
	PaginationTemplateFilePath = "templates/pagination.tmpl"
)
