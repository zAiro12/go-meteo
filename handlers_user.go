package main

import (
	"encoding/json"
	"net/http"
	"os"
)

// Middleware per autenticazione admin tramite password (da .env)
func adminAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pass := r.Header.Get("X-Admin-Password")
		if pass == "" || pass != GetAdminPassword() {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

// Recupera la password admin da variabile d'ambiente
func GetAdminPassword() string {
	return GetEnv("ADMIN_PASSWORD", "")
}

// Handler per ottenere il profilo utente
func UserProfileHandler(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("username")
	if username == "" {
		http.Error(w, "username richiesto", http.StatusBadRequest)
		return
	}
	user, err := GetUserByUsername(username)
	if err != nil {
		// Se non trovato, crea profilo base
		user = &User{Username: username, Notify: false}
		_ = UpsertUser(user)
	}
	json.NewEncoder(w).Encode(user)
}

// Handler per aggiornare il profilo utente (città, telegram, notifiche)
func UserProfileUpdateHandler(w http.ResponseWriter, r *http.Request) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, invalidJSONMsg, http.StatusBadRequest)
		return
	}
	if user.Username == "" {
		http.Error(w, "username richiesto", http.StatusBadRequest)
		return
	}
	// Aggiorna solo i campi Telegram e Notify
	dbUser, _ := GetUserByUsername(user.Username)
	if dbUser == nil {
		dbUser = &User{Username: user.Username}
	}
	// Aggiorna Telegram, Notify e opzionalmente lat/lon/city per posizione personale
	dbUser.TelegramUser = user.TelegramUser
	dbUser.Notify = user.Notify
	if user.Lat != 0 || user.Lon != 0 {
		dbUser.Lat = user.Lat
		dbUser.Lon = user.Lon
	}
	if user.City != "" {
		dbUser.City = user.City
	}
	// if lat/lon provided but city empty, try reverse geocode server-side
	if (dbUser.City == "" || dbUser.City == customLocationLabel) && (dbUser.Lat != 0 || dbUser.Lon != 0) {
		city, _, _ := getCityNameFromCoordinates(dbUser.Lat, dbUser.Lon)
		dbUser.City = city
	}
	err := UpsertUser(dbUser)
	if err != nil {
		http.Error(w, "errore salvataggio", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// Handler protetto per gestire notifiche (solo admin)
func AdminNotificationsHandler(w http.ResponseWriter, r *http.Request) {
	// ...gestione logica notifiche...
	w.Write([]byte("Gestione notifiche admin OK"))
}

// RegisterHandler crea un nuovo utente con username e password
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var payload struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, invalidJSONMsg, http.StatusBadRequest)
		return
	}
	if payload.Username == "" || payload.Password == "" {
		http.Error(w, "username e password richiesti", http.StatusBadRequest)
		return
	}
	if err := CreateUser(payload.Username, payload.Password); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

// LoginHandler verifica credenziali e ritorna il profilo utente
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var payload struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, invalidJSONMsg, http.StatusBadRequest)
		return
	}
	user, err := CheckUserPassword(payload.Username, payload.Password)
	if err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}
	// return user profile
	json.NewEncoder(w).Encode(user)
}

// Utility per leggere variabili d'ambiente con default
func GetEnv(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}
