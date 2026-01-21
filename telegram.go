package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// sendTelegramNotification invia una notifica meteo su Telegram
func sendTelegramNotification(data *WeatherData) error {
	if telegramBotToken == "" || telegramChatID == "" {
		return fmt.Errorf("telegram non configurato")
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

	payload := map[string]interface{}{
		"chat_id":    telegramChatID,
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
		return fmt.Errorf("telegram API status %d", resp.StatusCode)
	}

	return nil
}
