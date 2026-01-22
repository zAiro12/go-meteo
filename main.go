package main

import (
	"io"
	"fmt"
	
	"log"
	"net/http"
	"strings"
	"time"
)

// loggingResponseWriter wraps http.ResponseWriter to capture status
type loggingResponseWriter struct {
	http.ResponseWriter
	status int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.status = code
	lrw.ResponseWriter.WriteHeader(code)
}

// logMiddleware logs incoming requests and responses
func logMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lrw := &loggingResponseWriter{ResponseWriter: w, status: 200}

		// Read small request body for logging (without consuming for handlers)
		var bodySnippet string
		if r.ContentLength > 0 && (r.Method == "POST" || r.Method == "PUT") {
			data, err := io.ReadAll(r.Body)
			if err == nil {
				// restore Body for next reader
				r.Body = io.NopCloser(strings.NewReader(string(data)))
				if len(data) > 1024 {
					bodySnippet = string(data[:1024]) + "..."
				} else {
					bodySnippet = string(data)
				}
			}
		}

		log.Printf("--> %s %s from=%s body=%q", r.Method, r.URL.RequestURI(), r.RemoteAddr, bodySnippet)

		next(lrw, r)

		duration := time.Since(start)
		log.Printf("<-- %s %s status=%d dur=%s", r.Method, r.URL.RequestURI(), lrw.status, duration)
	}
}

func main() {
	loadConfig()
	InitMongo()

	// Attiva notifiche di default
	startNotifications()

	http.HandleFunc("/", logMiddleware(HomeHandler))
	http.HandleFunc("/toggle-notification", logMiddleware(toggleNotificationsHandler))
	http.HandleFunc("/config", logMiddleware(getConfigHandler))
	http.HandleFunc("/config/update", logMiddleware(updateConfigHandler))
	http.HandleFunc("/location/set", logMiddleware(setLocationHandler))
	http.HandleFunc("/location/reset", logMiddleware(resetLocationHandler))
	http.HandleFunc("/meteo/user/profile", logMiddleware(UserProfileHandler))
	http.HandleFunc("/meteo/user/update", logMiddleware(UserProfileUpdateHandler))
	http.HandleFunc("/meteo/admin/notifications", logMiddleware(adminAuth(AdminNotificationsHandler)))

	// Auth endpoints
	http.HandleFunc("/meteo/auth/register", logMiddleware(RegisterHandler))
	http.HandleFunc("/meteo/auth/login", logMiddleware(LoginHandler))

	// legacy endpoints kept for compatibility
	http.HandleFunc("/meteo/user/login", logMiddleware(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }))
	http.HandleFunc("/meteo/user/logout", logMiddleware(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }))

	fmt.Printf("🌐 Server su http://localhost%s\n", serverPort)
	log.Fatal(http.ListenAndServe(serverPort, nil))
}
