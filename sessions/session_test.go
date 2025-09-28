package sessions

import (
	"github.com/alexedwards/scs/v2/memstore"
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"testing"
	"time"
)

func TestSession_InitSession(t *testing.T) {

	// Set the environment to load the test .env file
	t.Setenv("COOKIE_NAME_TEST", "test_session")
	t.Setenv("COOKIE_LIFETIME_MINUTES_TEST", "30")
	t.Setenv("COOKIE_PERSISTENT_TEST", "true")
	t.Setenv("COOKIE_DOMAIN_TEST", "localhost")
	t.Setenv("COOKIE_SECURE_TEST", "false")
	t.Setenv("SESSION_STORE_TEST", "cookie")

	// Initialize the session configuration
	appSessionConfig := &Session{
		CookieName:       os.Getenv("COOKIE_NAME_TEST"),
		CookieLifeTime:   os.Getenv("COOKIE_LIFETIME_MINUTES_TEST"),
		CookiePersistent: os.Getenv("COOKIE_PERSISTENT_TEST"),
		CookieDomain:     os.Getenv("COOKIE_DOMAIN_TEST"),
		CookieSecure:     os.Getenv("COOKIE_SECURE_TEST"),
		SessionStore:     os.Getenv("SESSION_STORE_TEST"),
	}

	sm := appSessionConfig.InitSession()

	// Validate session configuration
	assert.Equal(t, "test_session", sm.Cookie.Name)
	assert.Equal(t, 30*time.Minute, sm.Lifetime)
	assert.True(t, sm.Cookie.Persist)
	assert.False(t, sm.Cookie.Secure)
	assert.Equal(t, "localhost", sm.Cookie.Domain)

	// Validate the session store based on the environment variable
	_, ok := sm.Store.(*memstore.MemStore)
	assert.True(t, ok, "expected store to be memstore")

	// Cleanup: unset environment variable
	if err := unSetAll(); err != nil {
		t.Fatal(err)
	}
	log.Println("unsetting successful")
}

func unSetAll() error {
	vars := []string{
		"COOKIE_NAME_TEST",
		"COOKIE_LIFETIME_MINUTES_TEST",
		"COOKIE_PERSISTENT_TEST",
		"COOKIE_DOMAIN_TEST",
		"COOKIE_SECURE_TEST",
		"SESSION_STORE_TEST",
	}

	for _, key := range vars {
		if err := os.Unsetenv(key); err != nil {
			return err
		}
	}
	return nil
}
