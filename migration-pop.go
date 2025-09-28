package sauri

import (
	"fmt"
	"github.com/gobuffalo/pop/v5"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// popConnect takes the name of a connection, default is "development",
// and will return that connection from the available `Connections`
func (s *Sauri) popConnect() (*pop.Connection, error) {
	txConn, err := pop.Connect("development")
	if err != nil {
		return nil, err
	}

	return txConn, nil
}

// CreatePopMigration creates both up and down migrations for the provided content
func (s *Sauri) CreatePopMigration(up, down []byte, migrationName string, migrationType string) error {
	// Validate migration type
	validTypes := []string{"fizz", "sql"}
	if !isValidMigrationType(migrationType, validTypes) {
		return fmt.Errorf("invalid migration type: %s, expected one of %v", migrationType, validTypes)
	}

	// Define the migration path
	migrationPath := filepath.Join(s.RootPath, "migrations")

	// Ensure the migration directory exists
	if _, err := os.Stat(migrationPath); os.IsNotExist(err) {
		if err := os.MkdirAll(migrationPath, os.ModePerm); err != nil {
			return fmt.Errorf("failed to create migration directory %s: %v", migrationPath, err)
		}
	}

	// Generate Unix microsecond timestamp for the migration filename prefix
	timestamp := time.Now().UnixMicro()

	// Clean the migration name to avoid invalid characters
	migrationName = cleanMigrationName(migrationName)

	// Construct file names for up and down migrations
	upFileName := fmt.Sprintf("%d_%s.%s.up.%s", timestamp, migrationName, migrationType, migrationType)
	downFileName := fmt.Sprintf("%d_%s.%s.down.%s", timestamp, migrationName, migrationType, migrationType)

	// Create full file paths
	upFilePath := filepath.Join(migrationPath, upFileName)
	downFilePath := filepath.Join(migrationPath, downFileName)

	// Write up migration file
	if err := writeFile(upFilePath, up); err != nil {
		return fmt.Errorf("failed to write up migration file %s: %v", upFilePath, err)
	}

	// Write down migration file
	if err := writeFile(downFilePath, down); err != nil {
		return fmt.Errorf("failed to write down migration file %s: %v", downFilePath, err)
	}

	fmt.Printf("Migration created successfully:\n  Up: %s\n  Down: %s\n", upFileName, downFileName)
	return nil

}

// cleanMigrationName removes or replaces invalid characters in the migration name
func cleanMigrationName(name string) string {
	// Replace spaces with underscores and remove any special characters
	return strings.Map(func(r rune) rune {
		if r == ' ' {
			return '_'
		} else if r == '.' || r == '/' || r == '\\' || r == ':' || r == '*' || r == '?' || r == '"' || r == '<' || r == '>' || r == '|' {
			return -1
		}
		return r
	}, name)
}

// isValidMigrationType checks if the provided migration type is valid
func isValidMigrationType(migrationType string, validTypes []string) bool {
	for _, v := range validTypes {
		if v == migrationType {
			return true
		}
	}
	return false
}

// writeFile writes the given content to the specified file path
func writeFile(filePath string, content []byte) error {
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		return err
	}
	return nil
}

func (s *Sauri) RunUpPopMigration(txn *pop.Connection) error {
	var migrationPath = s.RootPath + "/migrations"

	fileMigrator, err := pop.NewFileMigrator(migrationPath, txn)
	if err != nil {
		return err
	}

	err = fileMigrator.Up()
	if err != nil {
		return err
	}

	return nil
}

func (s *Sauri) RunDownPopMigration(txn *pop.Connection, steps ...int) error {
	var migrationPath = s.RootPath + "/migrations"

	step := 1
	if len(steps) > 0 {
		step = steps[0]
	}

	fileMigrator, err := pop.NewFileMigrator(migrationPath, txn)
	if err != nil {
		return err
	}

	err = fileMigrator.Down(step)
	if err != nil {
		return err
	}
	return nil
}

func (s *Sauri) RunResetPopMigration(txn *pop.Connection) error {
	var migrationPath = s.RootPath + "/migrations"

	fileMigrator, err := pop.NewFileMigrator(migrationPath, txn)
	if err != nil {
		return err
	}

	err = fileMigrator.Reset()
	if err != nil {
		return err
	}
	return nil
}
