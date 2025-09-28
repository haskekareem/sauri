package sauri

import (
	"crypto/rand"
	"fmt"
	"io"
	"math/big"
	"os"
	"path/filepath"
)

// CreateDirIfNotExists checks if a directory exists at dirPath and creates it if not
func (s *Sauri) CreateDirIfNotExists(dirPath string) error {
	// Clean the path to ensure it's properly formatted
	cleanPath := filepath.Clean(dirPath)

	// Define directory permissions mode
	const dirMode = 0755

	// Check if the directory exists
	if _, err := os.Stat(cleanPath); os.IsNotExist(err) {
		// Create directory with appropriate permissions (rwxr-xr-x)
		if err := os.MkdirAll(cleanPath, dirMode); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", cleanPath, err)
		}
	} else if err != nil {
		// Return error if Stat fails for any other reason
		return fmt.Errorf("failed to check directory %s: %w", cleanPath, err)
	}
	return nil
}

// CreateFileIfNotExist checks if a file exists at filePath and creates it if not
func (s *Sauri) CreateFileIfNotExist(filePath string) error {
	// Clean and get absolute path
	cleanPath := filepath.Clean(filePath)

	// Check if file exists
	if _, err := os.Stat(cleanPath); os.IsNotExist(err) {
		// Create the file (rwxr--r--)
		file, err := os.Create(cleanPath)
		if err != nil {
			return fmt.Errorf("failed to create file %s: %w", cleanPath, err)
		}
		defer func(file *os.File) {
			err := file.Close()
			if err != nil {
				_ = file.Close()
			}
		}(file)
	} else if err != nil {
		// Return error if Stat fails for any other reason
		return fmt.Errorf("failed to check file %s: %w", cleanPath, err)
	}
	return nil
}

// Loader Holds the list of directories to search for templates.
type Loader struct {
	dirs []string
}

func (l *Loader) Open(name string) (io.ReadCloser, error) {
	for _, dir := range l.dirs {
		//Build full file path by joining the current directory with the template name
		path := filepath.Join(dir, name)

		file, err := os.Open(path)
		//If the file exists and opens successfully, return it
		if err == nil {
			return file, nil
		}
		//If the file just doesn't exist, continue to the next directory
		if !os.IsNotExist(err) {
			return nil, err
		}
	}

	//After trying all directories, if the file wasn’t found, return
	return nil, os.ErrNotExist
}

func (l *Loader) Exists(name string) bool {
	for _, dir := range l.dirs {
		path := filepath.Join(dir, name)
		if _, err := os.Stat(path); err == nil {
			return true
		}

	}
	return false
}

// GenerateRandomString generates a random string of n characters
func (s *Sauri) GenerateRandomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	// hold the generated random characters.
	result := make([]byte, n)

	// The loop iterates n times to fill the result slice with random characters
	for i := 0; i < n; i++ {
		// This generates a cryptographically secure random number between 0 and
		// the length of letters (which is 62, since there are 62 possible characters).
		// The rand.Reader is a secure random number generator.
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return ""
		}
		result[i] = letters[num.Int64()]
	}
	return string(result)

}
