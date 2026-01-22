package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
)

// sendTelegramToChat invia il messaggio al singolo chat id
func sendTelegramToChat(chatID string, data *WeatherData) error {
	if telegramBotToken == "" {
		return fmt.Errorf("telegram bot token non configurato")
	}

	message := fmt.Sprintf(
		"🌤️ *Meteo %s*\n\n"+
			"🕐 %s\n\n"+
			"*Condizioni Attuali*\n"+
			"%s\n"+
			"🌡️ Temperatura: %.1f°C\n"+
			"💧 Umidità: %.0f%%\n"+
			"💨 Vento: %.1f km/h\n"+
			"🌧️ Precipitazioni: %.1f mm\n\n"+
			"*Oggi*\n"+
			"Max: %.1f°C | Min: %.1f°C",
		data.City,
		data.Time,
		data.CurrentCondition,
		data.CurrentTemp,
		data.Humidity,
		data.WindSpeed,
		data.Precipitation,
		data.TodayMax,
		data.TodayMin,
	)

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", telegramBotToken)

	payload := map[string]any{
		"chat_id":    chatID,
		"text":       message,
		"parse_mode": "Markdown",
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := http.Post(url, contentTypeJSON, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// read body for detailed error
		buf := new(bytes.Buffer)
		_, _ = buf.ReadFrom(resp.Body)
		bodyStr := buf.String()
		return fmt.Errorf("telegram API status %d: %s", resp.StatusCode, bodyStr)
	}

	return nil
}

// sendTelegramNotification invia notifiche a tutti gli utenti con notify=true e telegram_user impostato
func sendTelegramNotification(data *WeatherData) error {
	if telegramBotToken == "" {
		return fmt.Errorf("telegram bot token non configurato")
	}

	coll := MongoDB.Collection("users")
	ctx := context.Background()
	cursor, err := coll.Find(ctx, bson.M{"notify": true, "telegram_user": bson.M{"$ne": ""}})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	sent := 0
	for cursor.Next(ctx) {
		var u User
		if err := cursor.Decode(&u); err != nil {
			log.Printf("Errore decode user: %v", err)
			continue
		}
		if u.TelegramUser == "" {
			continue
		}
		if err := sendTelegramToChat(u.TelegramUser, data); err != nil {
			log.Printf("Errore invio a %s (%s): %v", u.Username, u.TelegramUser, err)
			continue
		}
		sent++
		log.Printf("Notifica inviata a %s (%s)", u.Username, u.TelegramUser)
	}

	if sent == 0 {
		log.Println("Nessun utente con notifiche attive e Telegram impostato")
	}
	return nil
}
