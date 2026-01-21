package main

import (
	"log"
	"time"
)

// notificationWorker gestisce l'invio periodico delle notifiche
func notificationWorker() {
	for {
		select {
		case <-stopChan:
			log.Println("📢 Notifiche disattivate")
			return
		case <-ticker.C:
			now := time.Now()
			configMutex.RLock()
			start := notificationStartHour
			end := notificationEndHour
			configMutex.RUnlock()

			hour := now.Hour()
			if hour < start || hour >= end {
				log.Printf("⏱️ Fuori fascia (%02d:00–%02d:00), ora=%02d", start, end, hour)
				continue
			}

			data, err := getWeather()
			if err != nil {
				log.Printf("❌ Errore meteo: %v", err)
				continue
			}

			if err := sendTelegramNotification(data); err != nil {
				log.Printf("❌ Errore notifica: %v", err)
			} else {
				log.Println("✅ Notifica inviata")
			}
		}
	}
}

// startNotifications avvia il sistema di notifiche periodiche
func startNotifications() {
	notificationsMutex.Lock()
	defer notificationsMutex.Unlock()

	if notificationsEnabled {
		return
	}

	configMutex.RLock()
	interval := notificationInterval
	configMutex.RUnlock()

	notificationsEnabled = true
	ticker = time.NewTicker(interval)
	stopChan = make(chan bool)

	log.Printf("📢 Notifiche attivate (intervallo: %v)", interval)

	go func() {
		data, err := getWeather()
		if err != nil {
			log.Printf("❌ Errore meteo iniziale: %v", err)
			return
		}
		if err := sendTelegramNotification(data); err != nil {
			log.Printf("❌ Errore notifica iniziale: %v", err)
		} else {
			log.Println("✅ Notifica iniziale inviata")
		}
	}()

	go notificationWorker()
}

// stopNotifications ferma il sistema di notifiche periodiche
func stopNotifications() {
	notificationsMutex.Lock()
	defer notificationsMutex.Unlock()

	if !notificationsEnabled {
		return
	}

	notificationsEnabled = false
	if ticker != nil {
		ticker.Stop()
	}
	if stopChan != nil {
		close(stopChan)
	}

	log.Println("📢 Notifiche disattivate")
}
