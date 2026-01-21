package main

import (
	"sync"
	"time"
)

// Costanti HTTP
const (
	contentTypeJSON     = "application/json"
	contentTypeHeader   = "Content-Type"
	methodNotAllowedMsg = "Method not allowed"
)

// Versione applicazione
const AppVersion = "1.0.4"

// Costante per posizione personalizzata
const customLocationLabel = "Posizione personalizzata"

// Variabili globali - Config
var (
	serverPort            string
	telegramBotToken      string
	telegramChatID        string
	notificationInterval  time.Duration
	notificationStartHour int
	notificationEndHour   int

	configMutex sync.RWMutex
)

// Variabili globali - Stato notifiche
var (
	notificationsEnabled = false
	notificationsMutex   sync.RWMutex
	ticker               *time.Ticker
	stopChan             chan bool
)

// Variabili globali - Posizione personalizzata
var (
	customLat     float64
	customLon     float64
	useCustom     bool
	locationMutex sync.RWMutex
)

// GeoLocation rappresenta una posizione geografica
type GeoLocation struct {
	Lat     float64 `json:"lat"`
	Lon     float64 `json:"lon"`
	City    string  `json:"city"`
	Country string  `json:"country"`
}

// WeatherData contiene i dati meteo per il template
type WeatherData struct {
	City                 string
	Country              string
	Lat                  float64
	Lon                  float64
	Time                 string
	CurrentCondition     string
	CurrentTemp          float64
	Humidity             float64
	WindSpeed            float64
	Visibility           float64
	Precipitation        float64
	TodayMax             float64
	TodayMin             float64
	TodayCondition       string
	TomorrowMax          float64
	TomorrowMin          float64
	TomorrowCondition    string
	NotificationsEnabled bool
	IntervalMinutes      int
	StartHour            int
	EndHour              int
	Version              string
}

// UpdateConfigRequest rappresenta una richiesta di aggiornamento configurazione
type UpdateConfigRequest struct {
	IntervalMinutes int `json:"interval_minutes"`
	StartHour       int `json:"start_hour"`
	EndHour         int `json:"end_hour"`
}

// ConfigResponse rappresenta la risposta di configurazione
type ConfigResponse struct {
	IntervalMinutes int  `json:"interval_minutes"`
	StartHour       int  `json:"start_hour"`
	EndHour         int  `json:"end_hour"`
	NotificationsOn bool `json:"notifications_on"`
}

// SetLocationRequest rappresenta una richiesta di impostazione posizione
type SetLocationRequest struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}
