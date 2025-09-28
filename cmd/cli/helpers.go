package main

import (
	"fmt"
	"github.com/fatih/color"
	_ "github.com/go-sql-driver/mysql"
	"io"
	"path/filepath"
	"strings"

	"os"
)

func setUp(arg2 string) {
	// do not need a .env file when running the new, help and version cmd so don't initialize the .env
	if arg2 != "new" && arg2 != "help" && arg2 != "version" {
		path, err := os.Getwd()
		if err != nil {
			exitGracefully(err)
		}
		sauri2.RootPath = path

		pathToSearch := filepath.Join(sauri2.RootPath, ".env")

		// 	load .env file
		err = sauri2.LoadAndSetEnv(pathToSearch)
		if err != nil {
			exitGracefully(err)
		}

		sauri2.DBConn.DatabaseType = os.Getenv("DATABASE_TYPE")
	}

}

func getDSN() (string, error) {
	dbType := sauri2.DBConn.DatabaseType

	// dsn holds the connection string
	var dsn string

	// Retrieve environment variables
	host := os.Getenv("DATABASE_HOST")
	port := os.Getenv("DATABASE_PORT")
	user := os.Getenv("DATABASE_USER")
	password := os.Getenv("DATABASE_PASS")
	dbname := os.Getenv("DATABASE_NAME")
	sslMode := os.Getenv("DATABASE_SSL_MODE")

	// convert my default jackc driver to the package used by the migrate package
	if dbType == "pgx" {
		dbType = "postgres"
	}

	// check database type and build a connection string
	switch dbType {
	case "postgresql", "postgres":
		// Use default ssl mode if not set
		if sslMode == "" {
			sslMode = "disable"
		}

		// with password configuration
		if password != "" {
			dsn = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s", user, password, host, port, dbname, sslMode)
		} else {
			// without password configuration
			dsn = fmt.Sprintf("postgres://%s@%s:%s/%s?sslmode=%s", user, host, port, dbname, sslMode)

		}
	case "mysql", "mariadb":
		// Use default ssl mode if not set
		if sslMode == "" {
			sslMode = "false"
		}
		// Build MySQL DSN
		dsn = fmt.Sprintf("mysql://%s:%s@/%s?parseTime=True&loc=Local", user, password, dbname)

	default:
		// Unsupported database type
		return "", fmt.Errorf("unsupported database type: %s", dbType)
	}

	return dsn, nil
}

func showHelp() {
	color.Yellow(`Available commands:

	help                      -show the help command
	version                   -show the version command
	migrate                   -run all up migration that have not been previously run
	migrate down              -reverse the most recently run migration
	migrate down all          -remove all migration previously run
	migrate reset             -run all down migration in reverse order then run run all up migration
	make migration <name>     -create two files, one for up migration and the other for down migration
	make controllers <name>   -create a stub controller in the controllers folder
	make models <name>        -create a new model in the data folder
	make auth 				  -create and run migration for authentication tables, models and middlewares
	make controllers          -create a stub controllers in the controllers folder
	make models				  -create a new models in the data folder
	make session              -create a table in the database to be used as a session store

`)
}

// exitGracefully Helper function to handle errors gracefully
func exitGracefully(err error, msg ...string) {
	message := ""
	if len(msg) > 0 {
		message = msg[0]
	}
	if err != nil {
		color.Blue("Error: %v\n", err)
	}
	if len(message) > 0 {
		color.Yellow(message)
	} else {
		color.Green("finished!")
	}
	os.Exit(0)
}

// copyFile Helper function to copy files
func copyFile(sourcePath, destPath string) error {
	source, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer func(source *os.File) {
		_ = source.Close()
	}(source)

	dest, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer func(dest *os.File) {
		_ = dest.Close()
	}(dest)

	_, err = io.Copy(dest, source)
	return err
}

// walkFuncUpdateSourceFiles scans through files in a directory tree, looks for Go source files (*.go), and replaces
// every occurrence of the string "myapp" with the value of appURL inside those files, saving the changes back to the same file.
func walkFuncUpdateSourceFiles(path string, fi os.FileInfo, err error) error {
	// check for error before doing anything
	// If an error occurred while traversing the file system,
	// stop processing this file/directory and return the error.
	if err != nil {
		return err
	}

	// check if the current file is directory
	// If the current path is a directory, do nothing and move on.
	if fi.IsDir() {
		return nil
	}

	// only check go files
	// Uses a glob pattern (*.go) to check if the file is a Go source file. If it’s not a .go file, skip it.
	matchedGoFiles, err := filepath.Match("*.go", fi.Name())
	if err != nil {
		return err
	}

	// now we have a matching go files, so read its content
	if matchedGoFiles {
		// Reads the file contents into memory.
		r, err := os.ReadFile(path)
		if err != nil {
			exitGracefully(err)
		}

		//Replaces all (-1) occurrences of the substring "myapp" with the global variable appURL.
		newContent := strings.Replace(string(r), "myapp", appURL, -1)

		// Writes the updated contents back to the same file.
		err = os.WriteFile(path, []byte(newContent), 0)
		if err != nil {
			exitGracefully(err)
		}
	}
	return nil
}

// updateSource used to update go files with user specified name
func updateSource() {
	// walk through the entire project including folder directories and subfolders
	err := filepath.Walk(".", walkFuncUpdateSourceFiles)
	if err != nil {
		exitGracefully(err)
	}
}

func copyDataToFile(data []byte, to string) error {
	err := os.WriteFile(to, data, 0644)
	if err != nil {
		return err
	}

	return nil
}

func fileExists(fileToCheck string) bool {

	if _, err := os.Stat(fileToCheck); os.IsNotExist(err) {
		return false
	}

	return true
}
