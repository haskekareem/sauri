package sauri

import (
	"bufio"
	"database/sql"
	"errors"
	"fmt"
	"github.com/CloudyKit/jet/v6"
	"github.com/dgraph-io/badger/v3"
	"github.com/haskekareem/sauri/cache"
	"github.com/haskekareem/sauri/renderer"
	"github.com/haskekareem/sauri/sessions"
	"github.com/haskekareem/sauri/validator"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// InitFolders check if those folders exists and if not create folders for a new project
func (s *Sauri) InitFolders(i initializedFoldersPath) error {
	currentRootPath := i.currentRootPath

	for _, folderName := range i.folderNames {
		// create folders if only they don't already exists
		err := s.CreateDirIfNotExists(filepath.Join(currentRootPath, folderName))
		if err != nil {
			return err
		}
	}
	return nil
}

// checkDotEnvFile check for the existence of a .env file and if not it create one
func (s *Sauri) checkDotEnvFile(pathToCheck string) error {
	err := s.CreateFileIfNotExist(fmt.Sprintf("%s/.env", pathToCheck))
	if err != nil {
		return err
	}
	return nil
}

// LoadAndSetEnv loads the environment variables from the .env file.
func (s *Sauri) LoadAndSetEnv(filePath ...string) error {
	//open the .env file
	envFile, err := os.Open(filePath[0])
	if err != nil {
		return err
	}
	// Ensure the file is closed when the function exits
	defer func(envFile *os.File) {
		err := envFile.Close()
		if err != nil {
			_ = envFile.Close()
		}
	}(envFile)

	// create a scanner to read the .env file line by line
	scanner := bufio.NewScanner(envFile)

	//read the file line by line
	for scanner.Scan() {
		line := scanner.Text()

		//remove any leading and trailing whitespaces
		line = strings.TrimSpace(line)

		//skip empty lines and those starting with "#"
		if line == "" || strings.HasPrefix(line, "#") {
			// Skips the current iteration of the loop and moves to the next line.
			continue
		}

		// Split the line into key and value at the first '=' character
		parts := strings.SplitN(line, "=", 2)
		// Checks if the line was successfully split into exactly two parts
		if len(parts) != 2 {
			continue
		}

		// Trim whitespace from key and value
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Set the environment variable
		if err := os.Setenv(key, value); err != nil {
			return err
		}
	}
	// Check for errors that may have occurred during scanning
	err = scanner.Err()
	if err != nil {
		return err
	}
	return nil
}

// createLoggers creates a customized loggers
func (s *Sauri) createLoggers() (*log.Logger, *log.Logger) {
	var infoLogger *log.Logger
	var errorLogger *log.Logger

	errorLogger = log.New(os.Stderr, "ERROR\t", log.Ltime|log.Ldate|log.Lshortfile)
	infoLogger = log.New(os.Stderr, "INFO\t", log.Ltime|log.Ldate)

	return infoLogger, errorLogger

}

// ListenAndServe creates a web server listening on the given port and serving
func (s *Sauri) ListenAndServe() {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", os.Getenv("PORT")),
		ErrorLog:     s.ErrorLog,
		Handler:      s.Router,
		IdleTimeout:  30 * time.Second,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 600 * time.Second,
	}

	// closing the connection pools
	if s.DBConn.SqlConnPool != nil {
		defer func(SqlConnPool *sql.DB) {
			_ = SqlConnPool.Close()
		}(s.DBConn.SqlConnPool)
	}

	if s.DBConn.PgxConnPool != nil {
		defer s.DBConn.PgxConnPool.Close()
	}

	s.InfoLog.Printf("Listening on port %s", os.Getenv("PORT"))

	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		s.ErrorLog.Fatalf("Could not listen on: %s: %v\n", os.Getenv("PORT"), err)
	}
}

// CreateRenderer creates a new Renderer instance
func (s *Sauri) CreateRenderer() {
	myRenderer := &renderer.Renderer{
		RendererEngine:    s.config.rendererEngine,
		TemplatesRootPath: "resources",
		Port:              s.config.port,
		JetViews:          s.JetViewsSetUp,
		DevelopmentMode:   s.DebugMode,
		Session:           s.Session,
	}
	s.Renderer = myRenderer
}

// InitializeJetSet sets up the Jet template set using the provided directories.
// It supports flexible configuration: either or both of layoutsDir and pagesDir can be provided.
// At least one directory must be non-empty, or an error is returned.
func (s *Sauri) InitializeJetSet(layoutsDir, pagesDir string) (*jet.Set, error) {
	var dirs []string

	// Add layouts directory if provided
	if layoutsDir != "" {
		dirs = append(dirs, layoutsDir)
	}

	// Add pages directory if provided and different from layouts directory
	if pagesDir != "" && pagesDir != layoutsDir {
		dirs = append(dirs, pagesDir)
	}
	// Ensure at least one directory is provided
	if len(dirs) == 0 {
		return nil, errors.New("at least one valid template directory must be provided")
	}

	// Create a loader with the valid directories
	loader := &Loader{dirs: dirs}

	// Create a new Jet template set with the custom loader
	var views *jet.Set
	if s.DebugMode {
		views = jet.NewSet(
			loader,
			jet.InDevelopmentMode())
	} else {
		views = jet.NewSet(loader)
	}

	return views, nil
}

// NewValidator creates a new Validator instance.
func (s *Sauri) NewValidator(data url.Values, FileData map[string]*multipart.FileHeader, rules map[string][]string, dbPool *sql.DB, pgx *pgxpool.Pool) *validator.Validation {
	return &validator.Validation{
		Data:             data,
		Errors:           validator.ErrorContainer{},
		Rules:            rules,
		CustomValidation: make(map[string]validator.CustomValidationFunc),
		CustomMessages:   make(map[string]string),
		AttributeAliases: make(map[string]string),
		FileData:         FileData,
		DIContainer:      map[string]interface{}{},
		StopOnFirstFail:  true, // Set this to true by default to enable stopping on first failure
		DBPool: struct {
			DBPoolSQL *sql.DB
			PoolPGX   *pgxpool.Pool
		}{DBPoolSQL: dbPool, PoolPGX: pgx},
	}
}

// initializeClientRedisCache create a cache redis client by initializing the
// redisCache struct type
func (s *Sauri) initializeClientRedisCache() *cache.RedisCache {
	return &cache.RedisCache{
		Conn:   s.NewRedisConnPool(),
		Prefix: s.config.redis.prefix,
	}
}

// initializeClientBadgerCache create a cache redis client by initializing the
// redisCache struct type
func (s *Sauri) initializeClientBadgerCache() *cache.BadgerCache {
	db, err := badger.Open(badger.DefaultOptions(s.RootPath + "storage/badger"))
	if err != nil {
		return nil
	}
	return &cache.BadgerCache{
		DBConn: db,
		Prefix: s.config.redis.prefix,
	}
}

// popSession initialize and populate the session manager
func (s *Sauri) popSession() {
	appSession := sessions.Session{
		CookieName:       s.config.cookie.name,
		CookieLifeTime:   s.config.cookie.lifetime,
		CookiePersistent: s.config.cookie.persist,
		CookieDomain:     s.config.cookie.domain,
		CookieSecure:     s.config.cookie.secure,
	}

	//populate values based on whether db store or redis is being used
	switch s.config.sessionStoreType {
	case "redis":
		appSession.RedisConnPool = myRedisCache.Conn
	case "mysql", "mariadb", "postgres", "postgresql":
		appSession.DBConnPool = s.DBConn.SqlConnPool
	}

	// initialized and store the session in Gudu type
	s.Session = appSession.InitSession()
}
