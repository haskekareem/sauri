package cache

import (
	"errors"
	"fmt"
	"github.com/dgraph-io/badger/v3"
	"log"
	"strings"
)

// ============================ HELPER FUNCTIONS ============================
// deleteKeysMatchingPattern deletes keys in batches of 10,000 that match the specified pattern.
// This function will handle deletion for both Empty() (using just the prefix) and
// EmptyByMatch() (using the pattern).
func (b *BadgerCache) deleteKeysMatchingPattern(txn *badger.Txn, pattern string, batchSize int) (int, error) {
	opts := badger.DefaultIteratorOptions
	opts.PrefetchValues = false
	it := txn.NewIterator(opts)
	defer it.Close()

	deleted := 0
	prefix := strings.Split(pattern, "*")[0] // Get the prefix from the pattern

	for it.Seek([]byte(prefix)); it.ValidForPrefix([]byte(prefix)); it.Next() {
		item := it.Item()
		key := string(item.Key())

		// If the user specifies a pattern like "key*", apply wildcard matching
		if strings.Contains(pattern, "*") {
			if !matchWildcard(key, pattern) {
				continue
			}
		}

		err := txn.Delete([]byte(key))
		if err != nil {
			return deleted, fmt.Errorf("failed to delete key %s: %w", key, err)
		}
		deleted++

		if deleted >= batchSize {
			break // Stop when batch size is reached
		}
	}

	return deleted, nil
}

// emptyWithRetries handles retry logic and batch processing for key deletion.
func (b *BadgerCache) emptyWithRetries(
	deleteFunc func(*badger.Txn, int) (int, error), batchSize int, maxRetries int) error {

	for {
		retries := 0
		for retries < maxRetries {
			err := b.DBConn.Update(func(txn *badger.Txn) error {
				deleted, err := deleteFunc(txn, batchSize)
				if err != nil {
					return err
				}

				if deleted == 0 {
					log.Println("No more keys to delete")
					return nil // Stop if no more keys are deleted
				}
				return nil
			})

			if err != nil {
				if errors.Is(err, badger.ErrConflict) {
					retries++
					log.Printf("Transaction conflict occurred. Retrying... (%d/%d)", retries, maxRetries)
					continue // Retry the transaction
				}
				return fmt.Errorf("failed to empty keys: %w", err) // Return on non-conflict errors
			}
			break // Exit retry loop if the transaction succeeds
		}

		if retries >= maxRetries {
			return fmt.Errorf("failed to empty keys after %d retries", maxRetries)
		}
		break // Exit the main loop if no more keys are deleted
	}
	return nil
}

// matchWildcard is a helper function that matches strings against wildcard patterns (like "key*").
func matchWildcard(str, pattern string) bool {
	// Split the pattern by "*" to get the parts before and after the wildcard.
	parts := strings.Split(pattern, "*")

	if len(parts) == 1 {
		// If there's no "*", the pattern should exactly match the string.
		return str == pattern
	}

	// Match the beginning part before the wildcard.
	if !strings.HasPrefix(str, parts[0]) {
		return false
	}

	// If there's a second part (after "*"), match the end of the string.
	if len(parts) > 1 && parts[1] != "" && !strings.HasSuffix(str, parts[1]) {
		return false
	}

	// If only prefix is present (like "test*"), and the key starts with "test", it's a match.
	return true
}
