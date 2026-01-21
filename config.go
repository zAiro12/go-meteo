package main

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// loadConfig carica la configurazione da .env e imposta i valori di default
func loadConfig() {
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8321"
	}
	serverPort = ":" + port

	telegramBotToken = os.Getenv("TELEGRAM_BOT_TOKEN")
	telegramChatID = os.Getenv("TELEGRAM_CHAT_ID")

	intervalMinutes := os.Getenv("NOTIFICATION_INTERVAL_MINUTES")
	if intervalMinutes == "" {
		intervalMinutes = "5"
	}
	minutes, err := strconv.Atoi(intervalMinutes)
	if err != nil || minutes <= 0 {
		minutes = 5
	}

	startHourStr := os.Getenv("NOTIFICATION_START_HOUR")
	endHourStr := os.Getenv("NOTIFICATION_END_HOUR")
	if startHourStr == "" {
		startHourStr = "7"
	}
	if endHourStr == "" {
		endHourStr = "18"
	}
	startHour, err := strconv.Atoi(startHourStr)
	if err != nil || startHour < 0 || startHour > 23 {
		startHour = 7
	}
	endHour, err := strconv.Atoi(endHourStr)
	if err != nil || endHour < 0 || endHour > 23 {
		endHour = 18
	}

	configMutex.Lock()
	notificationInterval = time.Duration(minutes) * time.Minute
	notificationStartHour = startHour
	notificationEndHour = endHour
	configMutex.Unlock()

	log.Printf("✅ Config caricata: port=%s, interval=%dmin, range=%02d-%02d",
		serverPort, minutes, startHour, endHour)
}
