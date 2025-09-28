package sauri

import (
	"database/sql"
	"github.com/jackc/pgx/v5/pgxpool"
)

// initializedFoldersPath used to list the folders in the current working directory
type initializedFoldersPath struct {
	currentRootPath string
	folderNames     []string
}

// redisConfig configs for redis cache
type redisConfig struct {
	host     string
	password string
	prefix   string
}

// sauriConfigs set the sauri package configurations and not exported
type sauriConfigs struct {
	port             string
	rendererEngine   string
	cookie           cookieConfig
	sessionStoreType string
	dBConfig         dataBaseConfig
	redis            redisConfig
}
type dataBaseConfig struct {
	dsn          string
	dataBaseType string
}

type DatabaseConn struct {
	DatabaseType string
	SqlConnPool  *sql.DB
	PgxConnPool  *pgxpool.Pool
}

// cookieConfig for session configurations
type cookieConfig struct {
	name     string
	lifetime string
	persist  string
	secure   string
	domain   string
}
