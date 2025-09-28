package cache

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/dgraph-io/badger/v3"
	"github.com/dgraph-io/ristretto/z"
	"io"
	"strings"
	"time"
)

// BadgerCache struct holds the Badger database instance and key prefix.
type BadgerCache struct {
	DBConn *badger.DB
	Prefix string
}

// ============================ METHODS ============================

// Close closes the badger connection pool.
func (b *BadgerCache) Close() error {
	if err := b.DBConn.Close(); err != nil {
		return err
	}
	return nil
}

// prefixedKey returns the key with the specified prefix.
func (b *BadgerCache) prefixedKey(key string) string {
	return fmt.Sprintf("%s:%s", b.Prefix, key)
}

// Set adds a key-value pair to the Badger cache with a prefixed key.
// It handles optional expiration time.
func (b *BadgerCache) Set(keyStr string, value interface{}, expires ...time.Duration) error {

	finalPrefixedKey := b.prefixedKey(keyStr)

	// Start a BadgerDB transaction and check for key existence
	return b.DBConn.Update(func(txn *badger.Txn) error {
		//Preparing the Entry for Storage
		itemEntry := EntryCache{}
		itemEntry[finalPrefixedKey] = value

		// Encode the value to a byte slice.
		// Converts the value into a byte array because BadgerDB stores data as binary (byte arrays).
		encodedValue, err := encodeValue(itemEntry)
		if err != nil {
			return fmt.Errorf("failed to encode value: %w", err)
		}

		//prepares an entry for insertion into BadgerDB.
		newEntry := badger.NewEntry([]byte(finalPrefixedKey), encodedValue)

		// Set TTL if provided.
		if len(expires) > 0 {
			newEntry.WithTTL(expires[0])
		}
		//insertion into BadgerDB.
		//Takes the entry and writes it to the database within the transaction.
		return txn.SetEntry(newEntry) //returns nil if successful or an error if something goes wrong.
	})
}

// SetMultiple allows for batch setting of multiple key-value pairs at once.
func (b *BadgerCache) SetMultiple(items EntryCache, expires ...time.Duration) error {
	wb := b.DBConn.NewWriteBatch() // Create a write batch
	defer wb.Cancel()

	for keyStr, value := range items {
		finalPrefixedKey := b.prefixedKey(keyStr)
		itemEntry := EntryCache{}
		itemEntry[finalPrefixedKey] = value

		encodedValue, err := encodeValue(itemEntry)
		if err != nil {
			return fmt.Errorf("failed to encode value: %w", err)
		}

		newEntry := badger.NewEntry([]byte(finalPrefixedKey), encodedValue)

		if len(expires) > 0 {
			newEntry.WithTTL(expires[0])
		}

		if err := wb.SetEntry(newEntry); err != nil {
			return err
		}
	}

	return wb.Flush()
}

// Get retrieves the value for a given prefixed key from the Badger cache
// and decodes it into an EntryCache.
func (b *BadgerCache) Get(keyStr string) (interface{}, error) {
	var result []byte
	prefixedKey := b.prefixedKey(keyStr)

	// Start a read-only transaction to view the database without modifying it
	err := b.DBConn.View(func(txn *badger.Txn) error {
		// Try to get the item (key-value pair) from the database
		item, err := txn.Get([]byte(prefixedKey))
		if err != nil {
			// Return the error if key is not found or another issue occurs
			return err
		}
		// Retrieve the actual value associated with the key
		err = item.Value(func(val []byte) error {
			// Copy the value into the result variable /The value must be copied because it’s only valid inside
			// the transaction
			result = append(result[:0], val...)
			return nil
		})
		return err
	})

	// error from the transaction
	if err != nil {
		// If the key was not found or an error occurred, return a user-friendly error
		if errors.Is(err, badger.ErrKeyNotFound) {
			return nil, fmt.Errorf("key not found")
		}
		return nil, fmt.Errorf("transaction to get the value failed: %w", err)
	}

	// Decode the retrieved byte data into its original type (could be any type like string, number, etc.)
	decoded, err := decodeValue(result)
	if err != nil {
		return nil, err
	}

	// retrieve the item from the map and return it
	item, exists := decoded[prefixedKey]
	if !exists {
		return nil, fmt.Errorf("key %s not found in decoded value", prefixedKey)
	}

	return item, nil
}

// GetAll retrieves all key-value pairs stored in Badger.
func (b *BadgerCache) GetAll() (EntryCache, error) {
	//This method can iterate through all keys and fetch their corresponding values.
	//This is useful when you need to scan the entire cache or a subset of it.
	results := EntryCache{}

	err := b.DBConn.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			var result []byte
			err := item.Value(func(val []byte) error {
				result = append(result[:0], val...)
				return nil
			})
			if err != nil {
				return err
			}

			decoded, err := decodeValue(result)
			if err != nil {
				return err
			}

			key := string(item.Key())
			results[key] = decoded[key]
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to retrieve all key-value pairs: %w", err)
	}

	return results, nil
}

// Update updates an existing key-value pair in the Badger cache, with an optional expiration time.
func (b *BadgerCache) Update(keyStr string, value interface{}, expires ...time.Duration) error {
	prefixedKey := b.prefixedKey(keyStr)

	// initiate a EntryCache instance and store the value in it
	entry := EntryCache{}
	entry[prefixedKey] = value

	encoded, err := encodeValue(entry)
	if err != nil {
		return fmt.Errorf("failed to encode value: %w", err)
	}
	// Update value in Badger with optional TTL
	return b.DBConn.Update(func(txn *badger.Txn) error {
		e := badger.NewEntry([]byte(prefixedKey), encoded)
		if len(expires) > 0 {
			e.WithTTL(expires[0])
		}
		return txn.SetEntry(e)
	})
}

// UpdateMultiple allows batch updating of multiple key-value pairs with optional TTL.
func (b *BadgerCache) UpdateMultiple(items EntryCache, expires ...time.Duration) error {
	wb := b.DBConn.NewWriteBatch()
	defer wb.Cancel()

	for keyStr, value := range items {
		finalPrefixedKey := b.prefixedKey(keyStr)
		entry := EntryCache{}
		entry[finalPrefixedKey] = value

		encoded, err := encodeValue(entry)
		if err != nil {
			return fmt.Errorf("failed to encode value: %w", err)
		}

		newEntry := badger.NewEntry([]byte(finalPrefixedKey), encoded)
		if len(expires) > 0 {
			newEntry.WithTTL(expires[0])
		}

		if err := wb.SetEntry(newEntry); err != nil {
			return err
		}
	}

	return wb.Flush()
}

// Exists checks if a key exists in the Badger cache.
func (b *BadgerCache) Exists(keyStr string) (bool, error) {
	prefixedKey := b.prefixedKey(keyStr)

	// If not in cache, check Badger database
	err := b.DBConn.View(func(txn *badger.Txn) error {
		_, err := txn.Get([]byte(prefixedKey))
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				// Key not found in Badger, ensure it's also removed from the cache
				return badger.ErrKeyNotFound
			}
			return err
		}

		return nil // Key exists and is not expired
	})

	if errors.Is(err, badger.ErrKeyNotFound) {
		// Key does not exist
		return false, nil
	}

	return err == nil, err // Return true if key exists, otherwise the error
}

// Delete removes a key-value pair with a prefixed key from the Badger cache.
func (b *BadgerCache) Delete(keyStr string) error {
	prefixedKey := b.prefixedKey(keyStr)
	// Delete the key in Badger
	return b.DBConn.Update(func(txn *badger.Txn) error {
		err := txn.Delete([]byte(prefixedKey))
		if err != nil {
			return fmt.Errorf("failed to delete key %s: %w", prefixedKey, err)
		}
		return nil
	})
}

// Backup performs a full backup of the Badger database.
func (b *BadgerCache) Backup(w io.Writer) (uint64, error) {
	ts, err := b.DBConn.Backup(w, 0) // 0 means backup all entries
	if err != nil {
		return 0, fmt.Errorf("failed to perform backup: %w", err)
	}
	return ts, nil
}

// Restore loads a backup from the provided reader into the Badger database.
func (b *BadgerCache) Restore(r io.Reader) error {
	return b.DBConn.Load(r, 10000) // Loads with a batch size of 10,000
}

// RunGC triggers garbage collection for the value log to reclaim disk space.
func (b *BadgerCache) RunGC(discardRatio float64) error {
	return b.DBConn.RunValueLogGC(discardRatio)
}

// DeleteMultiple allows batch deletion of multiple keys.
func (b *BadgerCache) DeleteMultiple(keys []string) error {
	wb := b.DBConn.NewWriteBatch()
	defer wb.Cancel()

	for _, keyStr := range keys {
		prefixedKey := b.prefixedKey(keyStr)
		if err := wb.Delete([]byte(prefixedKey)); err != nil {
			return fmt.Errorf("failed to delete keys %s: %w", prefixedKey, err)
		}
	}

	return wb.Flush()
}

// Clear drops all data in the database.
func (b *BadgerCache) Clear() error {
	// allows you to clear the entire cache efficiently.
	return b.DBConn.DropAll()
}

// StreamKeys retrieves keys in a stream, useful for large datasets.
func (b *BadgerCache) StreamKeys(batchSize int) ([]string, error) {
	// hold the retrieved keys
	var keys []string

	stream := b.DBConn.NewStream()
	stream.NumGo = 8
	stream.Prefix = []byte(b.Prefix)

	stream.Send = func(buf *z.Buffer) error {

		return nil
	}

	// Pass a valid context to Orchestrate (context.Background())
	if err := stream.Orchestrate(context.Background()); err != nil {
		return nil, err
	}

	return keys, nil
}

// Keys retrieves all keys matching a certain pattern, a specific key, or a list of keys.
func (b *BadgerCache) Keys(patternOrKey ...string) ([]string, error) {
	var keys []string

	if err := b.DBConn.View(func(txn *badger.Txn) error {
		// If no argument is provided, scan all keys with the prefix
		if len(patternOrKey) == 0 {
			opts := badger.DefaultIteratorOptions
			opts.PrefetchValues = false
			it := txn.NewIterator(opts)
			defer it.Close()

			prefixedPattern := fmt.Sprintf("%s:", b.Prefix)

			for it.Rewind(); it.Valid(); it.Next() {
				item := it.Item()
				if bytes.HasPrefix(item.Key(), []byte(prefixedPattern)) {
					keys = append(keys, string(item.Key()))
				}
			}
		} else if len(patternOrKey) == 1 {
			// todo; If the user specifies a wildcard pattern (e.g., "key*"), the Keys method
			//  iterates through all keys, removing the prefix and matching the key names against the pattern.

			// Handle single pattern or key
			pattern := patternOrKey[0]

			if strings.Contains(pattern, "*") {
				// Wildcard pattern, iterate over all keys and filter them
				opts := badger.DefaultIteratorOptions
				opts.PrefetchValues = false
				it := txn.NewIterator(opts)
				defer it.Close()

				prefixedPattern := b.prefixedKey("")
				for it.Rewind(); it.Valid(); it.Next() {
					item := it.Item()
					key := string(item.Key())
					if bytes.HasPrefix(item.Key(), []byte(prefixedPattern)) {
						trimmedKey := strings.TrimPrefix(key, b.Prefix+":")
						//compare the keys with the pattern.
						if matchWildcard(trimmedKey, pattern) {
							keys = append(keys, key)
						}
					}

				}
			} else {
				// Single key, check if it exists
				prefixedKey := b.prefixedKey(pattern)
				_, err := txn.Get([]byte(prefixedKey))
				if err == nil {
					keys = append(keys, prefixedKey)
				} else if !errors.Is(err, badger.ErrKeyNotFound) {
					return err
				}
			}

		} else {
			// Handle multiple specific keys
			for _, k := range patternOrKey {
				prefixedKey := b.prefixedKey(k)
				_, err := txn.Get([]byte(prefixedKey))
				if err == nil {
					keys = append(keys, prefixedKey)
				} else if !errors.Is(err, badger.ErrKeyNotFound) {
					return err
				}
			}
		}

		return nil

	}); err != nil {
		return nil, fmt.Errorf("failed to retrieve keys: %w", err)
	}
	return keys, nil
}

// KeysWithBatchSize retrieves all keys matching a certain pattern, a specific key, or a list of keys,
// with pagination support.
func (b *BadgerCache) KeysWithBatchSize(batchSize int, patternOrKey ...string) ([]string, error) {
	if batchSize == 0 {
		batchSize = 1000 // Default batch size, or any appropriate value
	}

	var keys []string

	if err := b.DBConn.View(func(txn *badger.Txn) error {
		// If no argument is provided, scan all keys with the prefix
		if len(patternOrKey) == 0 {
			opts := badger.DefaultIteratorOptions
			opts.PrefetchValues = false
			opts.PrefetchSize = batchSize
			it := txn.NewIterator(opts)
			defer it.Close()

			prefixedPattern := fmt.Sprintf("%s:", b.Prefix)

			for it.Rewind(); it.Valid(); it.Next() {
				item := it.Item()
				if bytes.HasPrefix(item.Key(), []byte(prefixedPattern)) {
					keys = append(keys, string(item.Key()))
					if len(keys) >= batchSize {
						// Stop when we reach the batch size
						break
					}
				}
			}
		} else if len(patternOrKey) == 1 {
			// todo; If the user specifies a wildcard pattern (e.g., "key*"), the Keys method
			//  iterates through all keys, removing the prefix and matching the key names against the pattern.

			// Handle single pattern or key
			pattern := patternOrKey[0]

			if strings.Contains(pattern, "*") {
				// Wildcard pattern, iterate over all keys and filter them
				opts := badger.DefaultIteratorOptions
				opts.PrefetchValues = false
				opts.PrefetchSize = batchSize
				it := txn.NewIterator(opts)
				defer it.Close()

				prefixedPattern := b.prefixedKey("")
				for it.Rewind(); it.Valid(); it.Next() {
					item := it.Item()
					key := string(item.Key())
					if bytes.HasPrefix(item.Key(), []byte(prefixedPattern)) {
						trimmedKey := strings.TrimPrefix(key, b.Prefix+":")
						//compare the keys with the pattern.
						if matchWildcard(trimmedKey, pattern) {
							keys = append(keys, key)
							if len(keys) >= batchSize {
								break
							}
						}
					}

				}
			} else {
				// Single key, check if it exists
				prefixedKey := b.prefixedKey(pattern)
				_, err := txn.Get([]byte(prefixedKey))
				if err == nil {
					keys = append(keys, prefixedKey)
				} else if !errors.Is(err, badger.ErrKeyNotFound) {
					return err
				}
			}

		} else {
			// Handle multiple specific keys
			for _, k := range patternOrKey {
				prefixedKey := b.prefixedKey(k)
				_, err := txn.Get([]byte(prefixedKey))
				if err == nil {
					keys = append(keys, prefixedKey)
				} else if !errors.Is(err, badger.ErrKeyNotFound) {
					return err
				}

				if len(keys) >= batchSize {
					break // Stop when we reach the batch size
				}
			}
		}

		return nil

	}); err != nil {
		return nil, fmt.Errorf("failed to retrieve keys: %w", err)
	}
	return keys, nil
}

// Expire sets a timeout on a key.
func (b *BadgerCache) Expire(keyStr string, expiration time.Duration) error {
	prefixedKey := b.prefixedKey(keyStr)

	// Update expiration time in Badger
	return b.DBConn.Update(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(prefixedKey))
		if err != nil {
			return err
		}

		// Get the existing value and set a new TTL
		return item.Value(func(val []byte) error {
			e := badger.NewEntry([]byte(prefixedKey), val).WithTTL(expiration)
			return txn.SetEntry(e)
		})
	})
}

// TTL retrieves the time-to-live of a key.
func (b *BadgerCache) TTL(keyStr string) (time.Duration, error) {
	prefixedKey := b.prefixedKey(keyStr)
	var ttl time.Duration

	if err := b.DBConn.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(prefixedKey))
		if err != nil {
			return err
		}
		// set the time to live for the key
		//Ensure that keys without TTL return a meaningful result (e.g., 0 duration).
		if item.ExpiresAt() > 0 {
			ttl = time.Until(time.Unix(int64(item.ExpiresAt()), 0))
		} else {
			ttl = 0
		}

		return nil
	}); err != nil {
		return 0, fmt.Errorf("failed to retrieve TTL: %w", err)
	}

	return ttl, nil
}

// RefreshTTL updates the expiration time (TTL) of an existing key.
func (b *BadgerCache) RefreshTTL(keyStr string, newTTL time.Duration) error {
	prefixedKey := b.prefixedKey(keyStr)

	return b.DBConn.Update(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(prefixedKey))
		if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			e := badger.NewEntry([]byte(prefixedKey), val).WithTTL(newTTL)
			return txn.SetEntry(e)
		})
	})
}

// EmptyByMatch deletes all keys matching a specific pattern
func (b *BadgerCache) EmptyByMatch(pattern string) error {
	// Extract the prefix from the pattern (e.g., "key")
	prefix := strings.Split(pattern, "*")[0]

	prefixedPattern := fmt.Sprintf("%s%s", b.Prefix+":", prefix)
	batchSize := 10000 // Default batch size for deleting keys
	maxRetries := 3

	return b.emptyWithRetries(func(txn *badger.Txn, batchSize int) (int, error) {
		return b.deleteKeysMatchingPattern(txn, prefixedPattern, batchSize)
	}, batchSize, maxRetries) // Batch size 10,000 and 3 max retries

}

// DropByPrefix drops all keys that match the prefix using Badger's DropPrefix.
func (b *BadgerCache) DropByPrefix() error {
	prefixedPattern := []byte(b.Prefix)
	return b.DBConn.DropPrefix(prefixedPattern)
}

// Empty deletes all keys with the specific prefix using a pipeline
func (b *BadgerCache) Empty() error {
	prefixedPattern := fmt.Sprintf("%s:", b.Prefix) // e.g., "gudu:"
	batchSize := 10000                              // Default batch size for deleting keys
	maxRetries := 3                                 // Max retries for handling transaction conflicts

	return b.emptyWithRetries(func(txn *badger.Txn, batchSize int) (int, error) {
		return b.deleteKeysMatchingPattern(txn, prefixedPattern, batchSize)
	}, batchSize, maxRetries)

}

// Sync flushes the database content to disk.
func (b *BadgerCache) Sync() error {
	return b.DBConn.Sync()
}

// Size returns the size of the LSM and value log files.
func (b *BadgerCache) Size() (int64, int64, error) {
	lsmSize, vlogSize := b.DBConn.Size()
	return lsmSize, vlogSize, nil
}
