package sessions

import (
	"database/sql"
	"github.com/alexedwards/scs/mysqlstore"
	"github.com/alexedwards/scs/postgresstore"
	"github.com/alexedwards/scs/redisstore"
	"github.com/alexedwards/scs/v2"
	"github.com/gomodule/redigo/redis"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Session struct {
	CookieName       string
	CookieLifeTime   string
	CookiePersistent string
	CookieDomain     string
	CookieSecure     string
	SessionStore     string
	DBConnPool       *sql.DB
	RedisConnPool    *redis.Pool
}

// InitSession initializes and configures a session manager based on the provided
// Session struct.
func (s *Session) InitSession() *scs.SessionManager {
	var secure, persist bool

	// how long should the session lasts
	lifetimeMinutes, err := strconv.Atoi(s.CookieLifeTime)
	if err != nil || lifetimeMinutes <= 0 {
		lifetimeMinutes = 60 // Default to 60 minutes if invalid or missing
	}

	// should cookies persist
	if strings.ToLower(s.CookiePersistent) == "true" {
		persist = true
	}

	// must cookies secure
	if strings.ToLower(s.CookieSecure) == "true" {
		secure = true
	}

	// Initialize a new session manager and configure the session.
	sm := scs.New()
	sm.Lifetime = time.Duration(lifetimeMinutes) * time.Minute
	// cookie settings
	sm.Cookie.Name = s.CookieName
	sm.Cookie.Persist = persist
	sm.Cookie.Secure = secure
	sm.Cookie.Domain = s.CookieDomain
	sm.Cookie.SameSite = http.SameSiteStrictMode

	// which session store
	switch strings.ToLower(s.SessionStore) {
	case "redis":
		// Configure session to use Redis store
		sm.Store = redisstore.New(s.RedisConnPool)
	case "mysql", "mariadb":
		// Configure session to use MySQL/MariaDB store
		sm.Store = mysqlstore.New(s.DBConnPool)
	case "postgres", "postgresql":
		// Configure session to use PostgresSQL store
		sm.Store = postgresstore.New(s.DBConnPool)
	default:
		// No external store specified, default to cookie-based session
	}

	return sm

}
