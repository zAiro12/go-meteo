package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

// toggleNotificationsHandler gestisce l'attivazione/disattivazione delle notifiche
func toggleNotificationsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, methodNotAllowedMsg, http.StatusMethodNotAllowed)
		return
	}

	notificationsMutex.RLock()
	current := notificationsEnabled
	notificationsMutex.RUnlock()

	if current {
		stopNotifications()
	} else {
		startNotifications()
	}

	notificationsMutex.RLock()
	newState := notificationsEnabled
	notificationsMutex.RUnlock()

	w.Header().Set(contentTypeHeader, contentTypeJSON)
	_ = json.NewEncoder(w).Encode(map[string]bool{"enabled": newState})
}

// getConfigHandler restituisce la configurazione corrente
func getConfigHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, methodNotAllowedMsg, http.StatusMethodNotAllowed)
		return
	}

	configMutex.RLock()
	interval := int(notificationInterval / time.Minute)
	start := notificationStartHour
	end := notificationEndHour
	configMutex.RUnlock()

	notificationsMutex.RLock()
	on := notificationsEnabled
	notificationsMutex.RUnlock()

	w.Header().Set(contentTypeHeader, contentTypeJSON)
	_ = json.NewEncoder(w).Encode(ConfigResponse{
		IntervalMinutes: interval,
		StartHour:       start,
		EndHour:         end,
		NotificationsOn: on,
	})
}

// updateConfigHandler aggiorna la configurazione delle notifiche
func updateConfigHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, methodNotAllowedMsg, http.StatusMethodNotAllowed)
		return
	}

	var req UpdateConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	if req.IntervalMinutes <= 0 {
		req.IntervalMinutes = 5
	}
	if req.StartHour < 0 || req.StartHour > 23 {
		req.StartHour = 7
	}
	if req.EndHour < 0 || req.EndHour > 23 {
		req.EndHour = 18
	}

	configMutex.Lock()
	notificationInterval = time.Duration(req.IntervalMinutes) * time.Minute
	notificationStartHour = req.StartHour
	notificationEndHour = req.EndHour
	configMutex.Unlock()

	notificationsMutex.Lock()
	if notificationsEnabled && ticker != nil {
		ticker.Stop()
		ticker = time.NewTicker(notificationInterval)
		log.Printf("🔁 Intervallo notifiche aggiornato a %d minuti", req.IntervalMinutes)
	}
	on := notificationsEnabled
	notificationsMutex.Unlock()

	w.Header().Set(contentTypeHeader, contentTypeJSON)
	_ = json.NewEncoder(w).Encode(ConfigResponse{
		IntervalMinutes: req.IntervalMinutes,
		StartHour:       req.StartHour,
		EndHour:         req.EndHour,
		NotificationsOn: on,
	})
}

// setLocationHandler imposta una posizione personalizzata per il meteo
func setLocationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, methodNotAllowedMsg, http.StatusMethodNotAllowed)
		return
	}

	var req SetLocationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	// Valida coordinate
	if req.Lat < -90 || req.Lat > 90 || req.Lon < -180 || req.Lon > 180 {
		http.Error(w, "Invalid coordinates", http.StatusBadRequest)
		return
	}
	// Require admin authorization to change global default
	if req.AdminUsername == "" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// verify user is admin
	u, err := GetUserByUsername(req.AdminUsername)
	if err != nil || u == nil || !u.IsAdmin {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// persist global default location in DB (collection app_config)
	coll := MongoDB.Collection("app_config")
	ctx := context.Background()
	filter := bson.M{"key": "default_location"}
	// attempt to get human-readable name for coordinates
	city, province, country := getCityNameFromCoordinates(req.Lat, req.Lon)
	pcode := provinceCodeFromName(province)
	displayName := city
	if pcode != "" {
		displayName = fmt.Sprintf("%s (%s), %s", city, pcode, country)
	} else if city != "" && country != "" {
		displayName = fmt.Sprintf("%s, %s", city, country)
	}

	value := bson.M{"lat": req.Lat, "lon": req.Lon, "city": city, "country": country, "display": displayName}
	_, err = coll.UpdateOne(ctx, filter, bson.M{"$set": bson.M{"value": value}}, optionsUpdateUpsert())
	if err != nil {
		log.Printf("Errore salvataggio default location su DB: %v", err)
		http.Error(w, "errore salvataggio", http.StatusInternalServerError)
		return
	}

	// update in-memory values too
	locationMutex.Lock()
	customLat = req.Lat
	customLon = req.Lon
	useCustom = true
	locationMutex.Unlock()

	log.Printf("📍 Posizione globale impostata da admin %s: %.4f, %.4f", req.AdminUsername, req.Lat, req.Lon)

	w.Header().Set(contentTypeHeader, contentTypeJSON)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"lat":     req.Lat,
		"lon":     req.Lon,
		"city":    city,
		"country": country,
		"display": displayName,
	})
}

// resetLocationHandler ripristina la geolocalizzazione automatica
func resetLocationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, methodNotAllowedMsg, http.StatusMethodNotAllowed)
		return
	}

	locationMutex.Lock()
	useCustom = false
	locationMutex.Unlock()

	log.Println("📍 Ripristinata geolocalizzazione automatica")

	w.Header().Set(contentTypeHeader, contentTypeJSON)
	_ = json.NewEncoder(w).Encode(map[string]bool{"success": true})
}
