package cache

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/dgraph-io/badger/v3"
	"os"
	"testing"
	"time"
)

// TestBadgerCache_Set Validates that the value stored in
// the cache matches the expected value.
func TestBadgerCache_Set(t *testing.T) {
	data := "value"

	err := testBadgerCache.Set("foo", data, 5*time.Minute)
	if err != nil {
		t.Error(err)
	}

	// Verify that the key exists and has the correct value
	result, err := testBadgerCache.Get("foo")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	retrievedData, ok := result.(string)
	if !ok {
		t.Errorf("Expected EntryCache type, got %T", result)
	}

	if retrievedData != "value" {
		t.Errorf("Expected %v, got %v", data, retrievedData)
	}

	err = testBadgerCache.Delete("foo")
	if err != nil {
		t.Error(err)
	}
}

// TestBadgerCache_SetMultiple validates that multiple key-value pairs are stored correctly in the cache.
func TestBadgerCache_SetMultiple(t *testing.T) {
	data := map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
	}

	err := testBadgerCache.SetMultiple(data, 5*time.Minute)
	if err != nil {
		t.Error(err)
	}

	// Verify the keys exist and have correct values
	result1, err := testBadgerCache.Get("key1")
	if err != nil {
		t.Errorf("Expected no error for key1, got %v", err)
	}
	if result1 != "value1" {
		t.Errorf("Expected 'value1', got %v", result1)
	}

	result2, err := testBadgerCache.Get("key2")
	if err != nil {
		t.Errorf("Expected no error for key2, got %v", err)
	}
	if result2 != "value2" {
		t.Errorf("Expected 'value2', got %v", result2)
	}

	err = testBadgerCache.Delete("key1")
	if err != nil {
		t.Error(err)
	}

	err = testBadgerCache.Delete("key2")
	if err != nil {
		t.Error(err)
	}
}

// TestBadgerCache_Get  Confirms the retrieved value matches the expected value.
func TestBadgerCache_Get(t *testing.T) {
	// Set a key-value pair
	data := "school-girl"
	err := testBadgerCache.Set("myKey", data, 10*time.Minute)
	if err != nil {
		t.Fatalf("Failed to set value: %v", err)
	}

	// Retrieve the value
	result, err := testBadgerCache.Get("myKey")
	if err != nil {
		t.Errorf("Failed to get value: %v", err)
	}

	retrievedData, ok := result.(string)
	if !ok {
		t.Errorf("Expected string type, got %T", result)
	}
	if retrievedData != data {
		t.Errorf("Expected %v, got %v", data, retrievedData)
	}
	err = testBadgerCache.Delete("myKey")
	if err != nil {
		t.Error(err)
	}
}

// TestBadgerCache_GetAll tests if all keys and values are retrieved correctly.
func TestBadgerCache_GetAll(t *testing.T) {
	// Set multiple key-value pairs
	err := testBadgerCache.Set("key1", "value1", 5*time.Minute)
	if err != nil {
		t.Error(err)
	}

	err = testBadgerCache.Set("key2", "value2", 5*time.Minute)
	if err != nil {
		t.Error(err)
	}

	// Get all keys and values
	result, err := testBadgerCache.GetAll()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 items, got %d", len(result))
	}

	if result["test-gudu:key1"] != "value1" || result["test-gudu:key2"] != "value2" {
		t.Errorf("Expected correct values for keys")
	}

	err = testBadgerCache.Delete("key1")
	if err != nil {
		t.Error(err)
	}

	err = testBadgerCache.Delete("key2")
	if err != nil {
		t.Error(err)
	}
}

// TestBadgerCache_Delete Checks that the key is deleted and cannot be retrieved.
func TestBadgerCache_Delete(t *testing.T) {
	// Set a key-value pair
	data := 6754
	err := testBadgerCache.Set("home", data, 10*time.Minute)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Delete the key
	err = testBadgerCache.Delete("home")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify that the key no longer exists
	_, err = testBadgerCache.Get("home")
	if err == nil {
		t.Errorf("Expected 'home' to be deleted, but it still exists.")
	}
}

// TestBadgerCache_Empty tests the Empty method.
func TestBadgerCache_Empty(t *testing.T) {
	// Set multiple key-value pairs
	data1 := 23
	data2 := 12

	err := testBadgerCache.Set("keyed1", data1, 10*time.Minute)
	if err != nil {
		t.Errorf("Failed to set keyed1: %v", err)
	}
	err = testBadgerCache.Set("keyed2", data2, 10*time.Minute)
	if err != nil {
		t.Errorf("Failed to set keyed2: %v", err)
	}

	// Empty all keys
	err = testBadgerCache.Empty()
	if err != nil {
		t.Errorf("Failed to empty all keys: %v", err)
	}

	/*keys, err := testBadgerCache.Keys()
	if err != nil {
		t.Errorf("Expected zero keys, got %s", keys)
	}
	if len(keys) != 0 {
		t.Errorf("Expected 0 keys, got %v", keys)
	}*/

	// Verify that the cache is empty
	_, err = testBadgerCache.Get("keyed1")
	if err == nil {
		t.Errorf("Expected error for deleted key, got none")
	}
}

// TestBadgerCache_DeleteMultiple tests the deletion of multiple keys.
func TestBadgerCache_DeleteMultiple(t *testing.T) {
	// Set multiple keys
	err := testBadgerCache.Set("key1", "value1", 5*time.Minute)
	if err != nil {
		t.Error(err)
	}
	err = testBadgerCache.Set("key2", "value2", 5*time.Minute)
	if err != nil {
		t.Error(err)
	}
	err = testBadgerCache.Set("key3", "value3", 5*time.Minute)
	if err != nil {
		t.Error(err)
	}

	// Delete multiple keys
	err = testBadgerCache.DeleteMultiple([]string{"key1", "key2"})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify that the keys are deleted
	_, err = testBadgerCache.Get("key1")
	if err == nil {
		t.Errorf("Expected error, got nil for deleted key1")
	}

	_, err = testBadgerCache.Get("key2")
	if err == nil {
		t.Errorf("Expected error, got nil for deleted key2")
	}

	// Verify non-deleted key still exists
	result, err := testBadgerCache.Get("key3")
	if err != nil {
		t.Errorf("Expected no error for key3, got %v", err)
	}
	if result != "value3" {
		t.Errorf("Expected 'value3', got %v", result)
	}

	err = testBadgerCache.Delete("key3")
	if err != nil {
		t.Error(err)
	}
}

// TestBadgerCache_DropByPrefix checks if all keys with a specific prefix are deleted.
func TestBadgerCache_DropByPrefix(t *testing.T) {
	// Set keys with the prefix "prefix"
	err := testBadgerCache.Set("key1", "value1", 5*time.Minute)
	if err != nil {
		t.Error(err)
	}
	err = testBadgerCache.Set("key2", "value2", 5*time.Minute)
	if err != nil {
		t.Error(err)
	}

	// Drop all keys with the prefix
	err = testBadgerCache.DropByPrefix()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify the keys with the prefix are deleted
	_, err = testBadgerCache.Get("key1")
	if err == nil {
		t.Errorf("Expected error, got nil for deleted key1")
	}
	_, err = testBadgerCache.Get("key2")
	if err == nil {
		t.Errorf("Expected error, got nil for deleted key2")
	}
}

// TestBadgerCache_TTL tests the TTL method of BadgerCache.
func TestBadgerCache_TTL(t *testing.T) {
	// Set a key-value pair with a 5-minute TTL
	err := testBadgerCache.Set("key11", "value1", 5*time.Minute)
	if err != nil {
		t.Fatalf("Failed to set value: %v", err)
	}

	// Get the TTL for the key
	ttl, err := testBadgerCache.TTL("key11")
	if err != nil {
		t.Fatalf("Failed to get TTL: %v", err)
	}

	if ttl.Minutes() <= 0 {
		t.Errorf("Expected a positive TTL, but got %v", ttl.Minutes())
	}

	// Clean up
	err = testBadgerCache.Delete("key11")
	if err != nil {
		t.Error(err)
	}
}

// TestBadgerCache_Expire tests the Expire method of BadgerCache.
func TestBadgerCache_Expire(t *testing.T) {
	// Set a key-value pair with a 5-minute TTL
	err := testBadgerCache.Set("key12", "value12", 5*time.Minute)
	if err != nil {
		t.Fatalf("Failed to set value: %v", err)
	}

	// Update the expiration time to 2 minutes
	err = testBadgerCache.Expire("key12", 2*time.Minute)
	if err != nil {
		t.Fatalf("Failed to update expiration: %v", err)
	}

	// Get the updated TTL
	ttl, err := testBadgerCache.TTL("key12")
	if err != nil {
		t.Fatalf("Failed to get updated TTL: %v", err)
	}

	if ttl.Minutes() > 2 {
		t.Errorf("Expected TTL <= 2 minutes, but got %v minutes", ttl.Minutes())
	}

	// Clean up
	err = testBadgerCache.Delete("key12")
	if err != nil {
		t.Error(err)
	}
}

// TestBadgerCache_Keys Ensures the correct keys are returned for various
// patterns and specific key requests.
func TestBadgerCache_Keys(t *testing.T) {
	err := testBadgerCache.Empty()
	if err != nil {
		t.Fatalf("Failed to empty cache: %v", err)
	}

	// Case 1: Test retrieving no keys (when no keys exist)
	keys, err := testBadgerCache.Keys()
	if err != nil {
		t.Errorf("Failed to retrieve keys: %v", err)
	}
	if len(keys) != 0 {
		t.Errorf("Expected 0 keys, but got %v", len(keys))
	}

	// Set multiple keys
	err = testBadgerCache.Set("key1", "value1", 5*time.Minute)
	if err != nil {
		t.Fatalf("Failed to set key1: %v", err)
	}
	err = testBadgerCache.Set("key2", "value2", 5*time.Minute)
	if err != nil {
		t.Fatalf("Failed to set key2: %v", err)
	}
	err = testBadgerCache.Set("key3", "value3", 5*time.Minute)
	if err != nil {
		t.Fatalf("Failed to set key3: %v", err)
	}

	// Case 2: Retrieve all keys with the given prefix
	keys, err = testBadgerCache.Keys()
	if err != nil {
		t.Errorf("Failed to retrieve keys: %v", err)
	}

	expectedKeys := []string{"test-gudu:key1", "test-gudu:key2", "test-gudu:key3"}
	if len(keys) != len(expectedKeys) {
		t.Errorf("Expected %v keys, got %v", len(expectedKeys), len(keys))
	}

	for _, expectedKey := range expectedKeys {
		found := false
		for _, key := range keys {
			if key == expectedKey {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected key '%s' not found in keys", expectedKey)
		}
	}
	// Case 3: Retrieve a specific key
	keys, err = testBadgerCache.Keys("key1")
	if err != nil {
		t.Errorf("Failed to retrieve specific key: %v", err)
	}
	if len(keys) != 1 || keys[0] != "test-gudu:key1" {
		t.Errorf("Expected 'test-gudu:key1', but got %v", keys)
	}

	// Case 4: Retrieve multiple specific keys
	keys, err = testBadgerCache.Keys("key1", "key2")
	if err != nil {
		t.Errorf("Failed to retrieve specified keys: %v", err)
	}
	expectedKeys = []string{"test-gudu:key1", "test-gudu:key2"}
	if len(keys) != len(expectedKeys) {
		t.Errorf("Expected %v keys, but got %v", len(expectedKeys), len(keys))
	}
	for _, expectedKey := range expectedKeys {
		found := false
		for _, key := range keys {
			if key == expectedKey {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected key '%s' not found in keys", expectedKey)
		}
	}

	// Case 5: Retrieve keys matching a specific pattern ("key*")
	keys, err = testBadgerCache.Keys("key*")
	if err != nil {
		t.Errorf("Failed to retrieve keys: %v", err)
	}
	expectedKeys = []string{"test-gudu:key1", "test-gudu:key2", "test-gudu:key3"}
	if len(keys) != len(expectedKeys) {
		t.Errorf("Expected %v keys with pattern, but got %v", len(expectedKeys), len(keys))
	}

	for _, key := range expectedKeys {
		if !contains(keys, key) {
			t.Errorf("Expected key '%s' not found in keys", key)
		}
	}

	// Clean up after test
	err = testBadgerCache.Delete("key1")
	if err != nil {
		t.Error(err)
	}
	err = testBadgerCache.Delete("key2")
	if err != nil {
		t.Error(err)
	}
	err = testBadgerCache.Delete("key3")
	if err != nil {
		t.Error(err)
	}
}

// TestBadgerCache_Update Ensures the updated value is correctly stored and retrieved.
func TestBadgerCache_Update(t *testing.T) {
	data := 30

	err := testBadgerCache.Set("foo", data, 10*time.Minute)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	updatedData := "john"
	err = testBadgerCache.Update("foo", updatedData)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify that the key exists and has the updated value
	result, err := testBadgerCache.Get("foo")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	retrievedData, ok := result.(string)
	if !ok {
		t.Errorf("Expected string type, got %T", result)
	}

	if retrievedData != updatedData {
		t.Errorf("Expected %v, got %v", updatedData, retrievedData)
	}
}

func TestBadgerCache_KeysWithBatchSize(t *testing.T) {
	// Test case: Retrieve keys with batch size limit
	for i := 0; i < 5; i++ {
		_ = testBadgerCache.Set(fmt.Sprintf("key%d", i), "value", 5*time.Minute)
	}

	keys, err := testBadgerCache.KeysWithBatchSize(3)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(keys) != 3 {
		t.Errorf("Expected 2 keys in batch, got %d", len(keys))
	}

}

// TestBadgerCache_Exists tests Verifies the existence and non-existence of keys.
func TestBadgerCache_Exists(t *testing.T) {
	// Case 1: Check existence of a non-existent key
	exists, err := testBadgerCache.Exists("nonExistentKey")
	if err != nil {
		t.Fatalf("Failed to check existence of 'nonExistentKey': %v", err)
	}
	if exists {
		t.Errorf("Expected 'nonExistentKey' to not exist, but it does.")
	}

	// Case 2: Set a key and check existence
	err = testBadgerCache.Set("existentKey", "someData", 5*time.Minute)
	if err != nil {
		t.Fatalf("Failed to set 'existentKey': %v", err)
	}

	exists, err = testBadgerCache.Exists("existentKey")
	if err != nil {
		t.Fatalf("Failed to check existence of 'existentKey': %v", err)
	}
	if !exists {
		t.Errorf("Expected 'existentKey' to exist, but it does not.")
	}

	// Case 3: Check existence after deleting the key
	err = testBadgerCache.Delete("existentKey")
	if err != nil {
		t.Fatalf("Failed to delete 'existentKey': %v", err)
	}

	exists, err = testBadgerCache.Exists("existentKey")
	if err != nil {
		t.Fatalf("Failed to check existence after deleting 'existentKey': %v", err)
	}
	if exists {
		t.Errorf("Expected 'existentKey' to not exist after deletion, but it still does.")
	}

	// Case 4: Set a key with a short TTL, wait for it to expire, and check existence
	err = testBadgerCache.Set("ttlKey", "someData", 1*time.Second)
	if err != nil {
		t.Fatalf("Failed to set 'ttlKey' with TTL: %v", err)
	}

	// Key should exist before expiration
	exists, err = testBadgerCache.Exists("ttlKey")
	if err != nil {
		t.Fatalf("Failed to check existence of 'ttlKey' before expiration: %v", err)
	}
	if !exists {
		t.Errorf("Expected 'ttlKey' to exist before expiration, but it does not.")
	}

	// Wait for the TTL to expire
	time.Sleep(2 * time.Second)

	// Key should no longer exist after expiration
	exists, err = testBadgerCache.Exists("ttlKey")
	if err != nil {
		t.Fatalf("Failed to check existence of 'ttlKey' after expiration: %v", err)
	}
	if exists {
		t.Errorf("Expected 'ttlKey' to not exist after expiration, but it still does.")
	}

}

// TestBadgerCache_EmptyByMatch tests the EmptyByMatch method.
func TestBadgerCache_EmptyByMatch(t *testing.T) {
	// Test case: Empty all keys matching a pattern
	err := testBadgerCache.Set("test1", "data1", 5*time.Minute)
	if err != nil {
		t.Error(err)
	}

	err = testBadgerCache.Set("test2", "data2", 5*time.Minute)
	if err != nil {
		t.Error(err)
	}

	err = testBadgerCache.Set("otherKey", "data3", 5*time.Minute)
	if err != nil {
		t.Error(err)
	}

	err = testBadgerCache.Set("try", "data3", 5*time.Minute)
	if err != nil {
		t.Error(err)
	}

	err = testBadgerCache.EmptyByMatch("test*")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Test case: Verify keys have been deleted
	_, err = testBadgerCache.Get("test1")
	if err == nil {
		t.Errorf("Expected error for deleted key, got none")
	}
	_, err = testBadgerCache.Get("test2")
	if err == nil {
		t.Errorf("Expected error for deleted key, got none")
	}

	// Test case: Verify that non-matching keys are not deleted
	_, err = testBadgerCache.Get("otherKey")
	if err != nil {
		t.Errorf("Expected no error for key 'otherKey', got %v", err)
	}

	// Verify that non-matching keys are not deleted
	_, err = testBadgerCache.Get("try")
	if err != nil {
		t.Errorf("Expected no error for key 'try', got %v", err)
	}
}

// TestBadgerCache_Backup tests the backup functionality.
func TestBadgerCache_Backup(t *testing.T) {
	// Set some data to back up
	err := testBadgerCache.Set("backupKey", "backupValue", 5*time.Minute)
	if err != nil {
		t.Fatalf("Failed to set data for backup: %v", err)
	}

	// Perform backup
	var buf bytes.Buffer
	_, err = testBadgerCache.Backup(&buf)
	if err != nil {
		t.Fatalf("Failed to perform backup: %v", err)
	}

	// Ensure backup is non-empty
	if buf.Len() == 0 {
		t.Fatalf("Expected non-empty backup, got empty buffer")
	}

	// Clean up
	err = testBadgerCache.Delete("backupKey")
	if err != nil {
		t.Error(err)
	}
}

// TestBadgerCache_Restore tests the restore functionality.
func TestBadgerCache_Restore(t *testing.T) {
	// Set some data and back it up
	err := testBadgerCache.Set("restoreKey", "restoreValue", 5*time.Minute)
	if err != nil {
		t.Fatalf("Failed to set data for restore test: %v", err)
	}

	var buf bytes.Buffer
	_, err = testBadgerCache.Backup(&buf)
	if err != nil {
		t.Fatalf("Failed to backup data: %v", err)
	}

	// Clear the cache and restore from backup
	err = testBadgerCache.Clear()
	if err != nil {
		t.Fatalf("Failed to clear cache before restore: %v", err)
	}

	err = testBadgerCache.Restore(&buf)
	if err != nil {
		t.Fatalf("Failed to restore data: %v", err)
	}

	// Verify the restored data
	result, err := testBadgerCache.Get("restoreKey")
	if err != nil {
		t.Fatalf("Failed to retrieve restored key: %v", err)
	}
	if result != "restoreValue" {
		t.Errorf("Expected 'restoreValue', got %v", result)
	}

	// Clean up
	err = testBadgerCache.Delete("restoreKey")
	if err != nil {
		t.Error(err)
	}
}

// TestBadgerCache_Clear tests the clear functionality of BadgerCache.
func TestBadgerCache_Clear(t *testing.T) {
	// Set multiple key-value pairs
	err := testBadgerCache.Set("key1", "value1", 5*time.Minute)
	if err != nil {
		t.Fatalf("Failed to set key1: %v", err)
	}
	err = testBadgerCache.Set("key2", "value2", 5*time.Minute)
	if err != nil {
		t.Fatalf("Failed to set key2: %v", err)
	}

	// Clear all data
	err = testBadgerCache.Clear()
	if err != nil {
		t.Fatalf("Failed to clear cache: %v", err)
	}

	// Verify that all keys are removed
	_, err = testBadgerCache.Get("key1")
	if err == nil {
		t.Fatalf("Expected key1 to be deleted, but it's still there")
	}

	_, err = testBadgerCache.Get("key2")
	if err == nil {
		t.Fatalf("Expected key2 to be deleted, but it's still there")
	}
}

// TestBadgerCache_RunGC tests the garbage collection functionality.
func TestBadgerCache_RunGC(t *testing.T) {
	// Run garbage collection
	err := testBadgerCache.RunGC(0.5)
	if err != nil && !errors.Is(err, badger.ErrNoRewrite) {
		t.Fatalf("Failed to run garbage collection: %v", err)
	}
}

// TestBadgerCache_Size tests the size retrieval functionality of BadgerCache.
func TestBadgerCache_Size(t *testing.T) {
	/* Open a new BadgerDB instance with SyncWrites set to true, to force data to disk.
	opts := ba.DefaultOptions("").WithInMemory(true).WithSyncWrites(true)
	db, err := badger.Open(opts)
	if err != nil {
		t.Fatalf("Failed to open Badger DB: %v", err)
	}
	defer db.Close()

	// Insert larger values to ensure that the data is flushed to the LSM and VLog.
	largeValue := strings.Repeat("largevalue", 1000) // Create a large value string to force disk writes.

	// Set multiple key-value pairs with large values to populate the LSM and VLog
	for i := 0; i < 300; i++ {
		err := testBadgerCache.Set(fmt.Sprintf("sizeKey%d", i), largeValue, 5*time.Minute)
		if err != nil {
			t.Fatalf("Failed to set sizeKey%d: %v", i, err)
		}
	}

	// Run value log garbage collection to trigger LSM and VLog write flushes.
	err := testBadgerCache.RunGC(0.5)
	if err != nil {
		t.Logf("Garbage collection returned an error or wasn't needed: %v", err)
	}

	// Force BadgerDB to flush its data to disk.
	err = testBadgerCache.Sync()
	if err != nil {
		t.Fatalf("Failed to sync database: %v", err)
	}

	// Allow some time for BadgerDB compaction to finish.
	time.Sleep(2 * time.Second)

	// Get the size of the LSM and VLog
	lsmSize, vlogSize, err := testBadgerCache.Size()
	if err != nil {
		t.Fatalf("Failed to retrieve size: %v", err)
	}

	// Check that LSM and VLog sizes are non-zero
	if lsmSize == 0 || vlogSize == 0 {
		t.Errorf("Expected non-zero LSM and VLog sizes, got LSM=%d, VLog=%d", lsmSize, vlogSize)
	}

	// Clean up the test data
	for i := 0; i < 300; i++ {
		err := testBadgerCache.Delete(fmt.Sprintf("sizeKey%d", i))
		if err != nil {
			t.Error(err)
		}
	}*/
}

// TestBadgerCache_StreamKeys tests the streaming of keys in batches.
func TestBadgerCache_StreamKeys(t *testing.T) {
	// Ensure the cache is empty before starting the test.
	err := testBadgerCache.Empty()
	if err != nil {
		t.Fatalf("Failed to empty cache: %v", err)
	}

	// Scenario 1: Ensure the cache is empty before starting the test
	keys, err := testBadgerCache.StreamKeys(3)
	if err != nil {
		t.Fatalf("StreamKeys failed: %v", err)
	}
	if len(keys) != 0 {
		t.Errorf("Expected no keys, got %v", keys)
	}

	// Set multiple key-value pairs
	for i := 0; i < 10; i++ {
		err := testBadgerCache.Set(fmt.Sprintf("streamKey%d", i), fmt.Sprintf("value%d", i), 5*time.Minute)
		if err != nil {
			t.Fatalf("Failed to set streamKey%d: %v", i, err)
		}
	}

	// Stream the keys
	keys, err = testBadgerCache.StreamKeys(5) // Retrieve keys in batches of 5
	if err != nil {
		t.Fatalf("Failed to stream keys: %v", err)
	}

	// Verify that the correct number of keys are retrieved
	if len(keys) != 5 {
		t.Errorf("Expected 5 keys in batch, but got %d", len(keys))
	}

	// Verify that the keys are correct
	expectedKeys := []string{"test-gudu:streamKey0", "test-gudu:streamKey1", "test-gudu:streamKey2", "test-gudu:streamKey3", "test-gudu:streamKey4"}
	for _, expectedKey := range expectedKeys {
		found := false
		for _, key := range keys {
			if key == expectedKey {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected key '%s' not found in stream result", expectedKey)
		}
	}

	// Clean up
	for i := 0; i < 10; i++ {
		err := testBadgerCache.Delete(fmt.Sprintf("streamKey%d", i))
		if err != nil {
			t.Error(err)
		}
	}

}

// TestBadgerCache_RefreshTTL tests if TTL is successfully refreshed for a key.
func TestBadgerCache_RefreshTTL(t *testing.T) {
	// Set a key with TTL
	err := testBadgerCache.Set("ttlKeyRefresh", "value", 5*time.Minute)
	if err != nil {
		t.Fatalf("Failed to set key: %v", err)
	}

	// Refresh TTL
	err = testBadgerCache.RefreshTTL("ttlKeyRefresh", 10*time.Minute)
	if err != nil {
		t.Fatalf("Failed to refresh TTL: %v", err)
	}

	// Verify the TTL is updated
	ttl, err := testBadgerCache.TTL("ttlKeyRefresh")
	if err != nil {
		t.Fatalf("Failed to retrieve TTL: %v", err)
	}

	if ttl < 9*time.Minute {
		t.Errorf("Expected TTL >= 9 minutes, got %v", ttl)
	}

	// Clean up
	err = testBadgerCache.Delete("ttlKeyRefresh")
	if err != nil {
		t.Error(err)
	}
}

// TestBadgerCache_Sync tests the sync functionality of BadgerCache.
func TestBadgerCache_Sync(t *testing.T) {
	// Perform a sync operation
	err := testBadgerCache.Sync()
	if err != nil {
		t.Fatalf("Failed to sync database: %v", err)
	}
}

// TestBadgerCache_UpdateMultiple tests the UpdateMultiple function by checking whether multiple key-value pairs are correctly updated.
func TestBadgerCache_UpdateMultiple(t *testing.T) {
	// Set multiple initial key-value pairs.
	items := EntryCache{
		"multiKey1": "initialValue1",
		"multiKey2": "initialValue2",
		"multiKey3": "initialValue3",
	}

	err := testBadgerCache.SetMultiple(items)
	if err != nil {
		t.Fatalf("Failed to set multiple keys: %v", err)
	}

	// Update the values for the keys.
	updatedItems := EntryCache{
		"multiKey1": "updatedValue1",
		"multiKey2": "updatedValue2",
		"multiKey3": "updatedValue3",
	}

	err = testBadgerCache.UpdateMultiple(updatedItems)
	if err != nil {
		t.Fatalf("Failed to update multiple keys: %v", err)
	}

	// Verify that the values are updated.
	for key, expectedValue := range updatedItems {
		result, err := testBadgerCache.Get(key)
		if err != nil {
			t.Errorf("Failed to get updated key %s: %v", key, err)
		}

		retrievedValue, ok := result.(string)
		if !ok {
			t.Errorf("Expected string type, got %T", result)
		}

		if retrievedValue != expectedValue {
			t.Errorf("Expected %v, got %v", expectedValue, retrievedValue)
		}
	}

	// Clean up.
	for key := range updatedItems {
		err := testBadgerCache.Delete(key)
		if err != nil {
			t.Errorf("Failed to delete key %s: %v", key, err)
		}
	}
}

// TestBadgerCache_Close ensures that the Badger database closes without errors.
func TestBadgerCache_Close(t *testing.T) {
	// Create a temporary directory
	tempDir := os.TempDir()

	defer func(path string) {
		_ = os.RemoveAll(path)
	}(tempDir) // Clean up the temporary directory when done

	// Open Badger DB in the temporary directory
	db, err := badger.Open(badger.DefaultOptions(tempDir))
	if err != nil {
		t.Fatalf("Failed to open Badger DB: %v", err)
	}

	badgerCache := &BadgerCache{
		DBConn: db,
		Prefix: "test-prefix",
	}

	// Close the Badger database.
	err = badgerCache.Close()
	if err != nil {
		t.Fatalf("Failed to close Badger DB: %v", err)
	}

	// Try to perform an operation on a closed database to confirm it fails.
	_, err = badgerCache.Get("someKey")
	if err == nil {
		t.Errorf("Expected error when accessing closed DB, but got none")
	}
}
