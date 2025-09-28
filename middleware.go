package sauri

import (
	"github.com/justinas/nosurf"
	"net/http"
	"strconv"
)

// SessionLoad takes care of loading and committing session data to the session store, and
// communicating the session token to/from the client in a cookie as necessary.
func (s *Sauri) SessionLoad(next http.Handler) http.Handler {
	return s.Session.LoadAndSave(next)
}

func (s *Sauri) NoSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)
	secure, _ := strconv.ParseBool(s.config.cookie.secure)

	//csrfHandler.ExemptGlob("/api/*")

	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
		Domain:   s.config.cookie.domain,
	})

	return csrfHandler
}
