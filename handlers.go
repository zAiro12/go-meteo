package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
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

	locationMutex.Lock()
	customLat = req.Lat
	customLon = req.Lon
	useCustom = true
	locationMutex.Unlock()

	log.Printf("📍 Posizione personalizzata impostata: %.4f, %.4f", req.Lat, req.Lon)

	w.Header().Set(contentTypeHeader, contentTypeJSON)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"lat":     req.Lat,
		"lon":     req.Lon,
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
