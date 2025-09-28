package sauri

import (
	"github.com/CloudyKit/jet/v6"
	"github.com/alexedwards/scs/v2"
	"github.com/dgraph-io/badger/v3"
	"github.com/go-chi/chi/v5"
	"github.com/haskekareem/sauri/cache"
	"github.com/haskekareem/sauri/renderer"
	"log"
	"os"
	"path/filepath"
	"strconv"
)

const version = "1.0.0"

var myRedisCache *cache.RedisCache
var myBadgerCache *cache.BadgerCache
var badgerPool *badger.DB

type Sauri struct {
	AppName       string
	DebugMode     bool
	Version       string
	InfoLog       *log.Logger
	ErrorLog      *log.Logger
	RootPath      string
	config        sauriConfigs
	EncryptionKey string
	Cache         cache.Cache
	Router        *chi.Mux
	Renderer      *renderer.Renderer  // Go Rendering engine
	JetViewsSetUp *jet.Set            // Jet rendering engine
	Session       *scs.SessionManager // session management
	DBConn        DatabaseConn
	Responses     *Response
	//Mailer        *mails.Mailer
}

// NewApp is the main project setup
func (s *Sauri) NewApp(currentRootPath string) error {
	//todo: create empty folders for a new project if those folders do not exists

	// create empty folders for a new project
	folderPathConfig := initializedFoldersPath{
		currentRootPath: currentRootPath,
		folderNames: []string{
			"cmd/server",          // main entry point
			"internal/controller", // route handlers
			"internal/model",      // business and database models
			"internal/route",      // route definitions
			"internal/config",     // app configuration
			"internal/migration",  // Database migration
			"internal/mailer",     // mailer logic
			"internal/middleware", // middleware
			"pkg/utils",           // shared utility functions
			"public",              // static files (CSS/JS/images)
			"resources/views",     // template files
			"storage/logs",        // log storage
			"storage/uploads",     // file uploads
			"test",                // test files
			"tmp",                 // temporary files for the project
		},
	}
	// initialize folders during project setup
	err := s.InitFolders(folderPathConfig)
	if err != nil {
		return err
	}
	//initialize nested folders under resources/views
	viewSubFolders := initializedFoldersPath{
		currentRootPath: filepath.Join(currentRootPath, "resources", "views"),
		folderNames:     []string{"layouts", "pages"},
	}
	err = s.InitFolders(viewSubFolders)
	if err != nil {
		return err
	}

	//todo: check if there is .env file in the current wd and if doesn't create one
	err = s.checkDotEnvFile(currentRootPath)
	if err != nil {
		return err
	}

	// todo: if there is a .env file then read its content and put it in the env variable
	err = s.LoadAndSetEnv(filepath.Join(currentRootPath, ".env"))
	if err != nil {
		return err
	}

	//todo: create customised loggers for the project
	infoLog, errorLog := s.createLoggers()

	s.Responses = s.NewResponse()

	// todo: call OpenDBConnectionPool to connect to the DB
	//// Build DSN based on environment variables
	dsn, err := s.BuildDSN()
	if err != nil {
		errorLog.Println("can not build DSN: ", err)
		return err
	}
	dbDriverType := os.Getenv("DATABASE_TYPE")
	dbUse, _ := strconv.ParseBool(os.Getenv("DATABASE_USE"))
	if dbUse {
		sqlDB, pgxPool, err := s.OpenDBConnectionPool(dbDriverType, dsn)
		if err != nil {
			errorLog.Println("can not open DB connection pool: ", err)
			os.Exit(1)
		}
		// populate database in the sauri structure
		s.DBConn = DatabaseConn{
			DatabaseType: dbDriverType,
			SqlConnPool:  sqlDB,
			PgxConnPool:  pgxPool,
		}
	}

	// todo connect to redis server
	if os.Getenv("CACHE") == "redis" || os.Getenv("SESSION_STORE_TYPE") == "redis" {
		myRedisCache = s.initializeClientRedisCache()
		s.Cache = myRedisCache
	}

	// todo connect to badger database
	if os.Getenv("CACHE") == "badger" {
		myBadgerCache = s.initializeClientBadgerCache()
		s.Cache = myBadgerCache
		badgerPool = myBadgerCache.DBConn
		// set periodic garbage collection once a day
		//_, err = s.Mailer.Scheduler.C.AddFunc("@daily", func() {
		//	_ = myBadgerCache.DBConn.RunValueLogGC(0.7)
		//})
	}

	/*if err != nil {
		return err
	}*/

	s.InfoLog = infoLog
	s.ErrorLog = errorLog
	s.DebugMode, _ = strconv.ParseBool(os.Getenv("DEBUG_MODE"))
	s.Version = version
	s.RootPath = currentRootPath

	//todo: populating the package configurations using values from env file
	s.config = sauriConfigs{
		port:           os.Getenv("PORT"),
		rendererEngine: os.Getenv("RENDER_ENGINE"),
		cookie: cookieConfig{
			name:     os.Getenv("COOKIE_NAME"),
			lifetime: os.Getenv("COOKIE_LIFETIME"),
			persist:  os.Getenv("COOKIE_PERSIST"),
			secure:   os.Getenv("COOKIE_SECURE"),
			domain:   os.Getenv("COOKIE_DOMAIN"),
		},
		sessionStoreType: os.Getenv("SESSION_STORE_TYPE"),
		dBConfig: dataBaseConfig{
			dsn:          dsn,
			dataBaseType: dbDriverType,
		},
		redis: redisConfig{
			host:     os.Getenv("REDIS_HOST"),
			password: os.Getenv("REDIS_PASSWORD"),
			prefix:   os.Getenv("REDIS_PREFIX"),
		},
	}

	// todo: router populate
	s.Router = s.defaultRouter().(*chi.Mux)

	// todo Session Initialization and setup
	s.popSession()

	//setting the jet template engine
	viewsDir := filepath.Join(currentRootPath, "resources", "views")
	s.JetViewsSetUp, _ = s.InitializeJetSet(viewsDir, "")

	// creates a new Renderer instance for Go template and initialize its fields
	s.CreateRenderer()

	// Listen for incoming emails on the emailQueue channel
	//go s.Mailer.ListenForEmails()

	return nil
}
