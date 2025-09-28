package main

import (
	"errors"
	"fmt"
	"github.com/fatih/color"
	"github.com/gertd/go-pluralize"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// doAuth build the subcommand of authentication for make command
func doAuth() error {
	// make migration
	dbType := sauri2.DBConn.DatabaseType
	fileName := fmt.Sprintf("%d_create_auth_table", time.Now().UnixMicro())

	targetUpFilePath := filepath.Join(sauri2.RootPath, "internal", "migration", fileName+"."+dbType+".up.sql")
	targetDownFilePath := filepath.Join(sauri2.RootPath, "internal", "migration", fileName+"."+dbType+".down.sql")

	// templates for the migration (existing contents embed to be copied to the target folders
	tempPathUp := "templates/migrations/auth_table." + dbType + ".up.sql"
	tempPathDown := "drop table if exists users cascade; drop table if exists tokens cascade; drop table if exists remember_tokens;"

	err := copyFilesFromTemplate(tempPathUp, targetUpFilePath)
	if err != nil {
		exitGracefully(err)
	}

	err = copyDataToFile([]byte(tempPathDown), targetDownFilePath)
	if err != nil {
		exitGracefully(err)
	}

	//run up migration by adding migrate command directly
	err = doMigrate("up", "")
	if err != nil {
		exitGracefully(err)
	}

	targetDir := filepath.Join(sauri2.RootPath, "internal")

	// copy the data models and its methods and also make sure existing files don't overwrite
	err = copyFilesFromTemplate("templates/data/user.go.txt", filepath.Join(targetDir, "model", "user.go"))
	if err != nil {
		exitGracefully(err)
	}

	err = copyFilesFromTemplate("templates/data/token.go.txt", filepath.Join(targetDir, "model", "token.go"))
	if err != nil {
		exitGracefully(err)
	}

	//copy the  middleware
	err = copyFilesFromTemplate("templates/middleware/auth-web.go.txt", filepath.Join(targetDir, "middleware", "auth-web.go"))
	if err != nil {
		exitGracefully(err)
	}

	err = copyFilesFromTemplate("templates/middleware/auth-token.go.txt", filepath.Join(targetDir, "middleware", "auth-token.go"))
	if err != nil {
		exitGracefully(err)
	}

	//display message feedback to end users
	color.Yellow("   -users, tokens and remember_tokens migration created and executed")
	color.Yellow("   -user and token models created!!")
	color.Yellow("   -auth middleware created!!")
	color.Yellow("")
	color.Red(" -dont forget to add user and token models in internal/model/models.go " +
		"and add appropriate middleware to your routes")

	return nil
}

// doMigration build the subcommand of migration for make command that create two files for up and down
// migrations
func doMigration(arg4 string) error {
	dbType := sauri2.DBConn.DatabaseType
	// checking for migration name
	if arg4 == "" {
		exitGracefully(errors.New("must give the migration a name"))
	}

	migrationFileName := fmt.Sprintf("%d_%s", time.Now().UnixMicro(), arg4)

	// path the up and down migration folders
	targetUpFilePath := filepath.Join(sauri2.RootPath, "internal", "migration", migrationFileName+"."+dbType+".up.sql")
	targetDownFilePath := filepath.Join(sauri2.RootPath, "internal", "migration", migrationFileName+"."+dbType+".down.sql")

	// templates for the migration (existing contents embed to be copied to the target folders
	tempPathUp := "templates/migrations/migration." + dbType + ".up.sql"
	tempPathDown := "templates/migrations/migration." + dbType + ".down.sql"

	err := copyFilesFromTemplate(tempPathUp, targetUpFilePath)
	if err != nil {
		exitGracefully(err)
	}

	err = copyFilesFromTemplate(tempPathDown, targetDownFilePath)
	if err != nil {
		exitGracefully(err)
	}

	return nil
}

// doControllers build the subcommand of handlers for make command
func doControllers(arg4 string) error {
	// Check for empty controller name
	if arg4 == "" {
		exitGracefully(errors.New("must give the controller a name"))
	}

	// Convert input to proper CamelCase
	controllerName := convertInput(arg4)

	// Get raw lowercase and process it for file name
	raw := strings.ToLower(arg4)
	fileBase := normalizeSeparators(splitCompoundWord(raw))

	// Only use splitCompoundWord result if it's a compound word
	// A compound word is one that gets split (i.e., has a dash after processing)
	var fileName string
	if strings.Contains(fileBase, "-") {
		fileName = fileBase + ".go"
	} else {
		fileName = raw + ".go"
	}

	targetControl := filepath.Join(sauri2.RootPath, "internal", "controller", fileName)
	if fileExists(targetControl) {
		exitGracefully(errors.New(targetControl + " file already exists"))
	}

	// Read template file (assumed to exist in templates/controllers/controller.go.txt)
	data, err := templateFS.ReadFile("templates/controllers/controller.go.txt")
	if err != nil {
		exitGracefully(err)
	}

	// Replace placeholder in template
	controller := strings.ReplaceAll(string(data), "$CONTROLLERNAME$", controllerName)

	// Write the file
	err = os.WriteFile(targetControl, []byte(controller), 0644)
	if err != nil {
		exitGracefully(err)
	}

	return nil
}

// doModels build the subcommand of models for make command
func doModels(arg4 string) error {
	// checking for model name
	if arg4 == "" {
		exitGracefully(errors.New("must give the model a name"))
	}

	data, err := templateFS.ReadFile("templates/data/model.go.txt")
	if err != nil {
		exitGracefully(err)
	}
	model := string(data)

	plur := pluralize.NewClient()

	var modelName = arg4
	var tableName = arg4

	if plur.IsPlural(arg4) {
		modelName = plur.Singular(arg4)
		tableName = strings.ToLower(tableName)
	} else {
		tableName = strings.ToLower(plur.Plural(arg4))
	}

	// target file
	// Convert input to proper CamelCase
	caseModelName := convertInput(arg4)
	fileName := modelName + ".go"

	targetFile := filepath.Join(sauri2.RootPath, "internal", "model", strings.ToLower(fileName))
	if fileExists(targetFile) {
		exitGracefully(errors.New(targetFile + " file already exists"))
	}

	// final version of data going to the target file
	model = strings.ReplaceAll(model, "$MODELNAME$", caseModelName)
	model = strings.ReplaceAll(model, "$TABLENAME$", tableName)

	// copy data to the files
	err = copyDataToFile([]byte(model), targetFile)
	if err != nil {
		exitGracefully(err)
	}

	return nil
}

// doSessionTable build the subcommand for session store for make command
func doSessionTable() error {
	dbType := sauri2.DBConn.DatabaseType

	// configuring database type
	switch dbType {
	case "postgres", "postgresql":
		dbType = "postgres"

	case "mysql", "mariadb":
		dbType = "mysql"
	}

	fileName := fmt.Sprintf("%d_create_session_table", time.Now().UnixMicro())

	// path the up and down migration folders
	targetUpFilePath := filepath.Join(sauri2.RootPath, "internal", "migration", fileName+"."+dbType+".up.sql")
	targetDownFilePath := filepath.Join(sauri2.RootPath, "internal", "migration", fileName+"."+dbType+".down.sql")

	// templates for the migration (existing contents embed to be copied to the target folders
	tempPathUp := "templates/migrations/session_table." + dbType + ".up.sql"
	tempPathDown := "drop table session"

	err := copyFilesFromTemplate(tempPathUp, targetUpFilePath)
	if err != nil {
		exitGracefully(err)
	}

	err = copyDataToFile([]byte(tempPathDown), targetDownFilePath)
	if err != nil {
		exitGracefully(err)
	}

	//run up migration by adding migrate command directly
	err = doMigrate("up", "")
	if err != nil {
		exitGracefully(err)
	}

	return nil
}
