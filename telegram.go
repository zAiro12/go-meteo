package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

const (
	telegramTokenNotConfiguredMsg = "telegram bot token non configurato"
)

// sendTelegramToChat invia il messaggio al singolo chat id
func sendTelegramToChat(chatID string, data *WeatherData) error {
	if telegramBotToken == "" {
		return fmt.Errorf(telegramTokenNotConfiguredMsg)
	}

	// Anteponi data/ora corrente in formato italiano (gg/mm/aaaa HH:MM)
	nowStr := time.Now().Format("02/01/2006 15:04")
	message := fmt.Sprintf(
		"%s\n🌤️ *Meteo %s*\n\n"+
			"🕐 %s\n\n"+
			"*Condizioni Attuali*\n"+
			"%s\n"+
			"🌡️ Temperatura: %.1f°C\n"+
			"💧 Umidità: %.0f%%\n"+
			"💨 Vento: %.1f km/h\n"+
			"🌧️ Precipitazioni: %.1f mm\n\n"+
			"*Oggi*\n"+
			"Max: %.1f°C | Min: %.1f°C",
		nowStr,
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

// sendTelegramText invia un messaggio di testo semplice al chat id
func sendTelegramText(chatID, text string) error {
	if telegramBotToken == "" {
		return fmt.Errorf(telegramTokenNotConfiguredMsg)
	}
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", telegramBotToken)
	payload := map[string]any{
		"chat_id": chatID,
		"text":    text,
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
		buf := new(bytes.Buffer)
		_, _ = buf.ReadFrom(resp.Body)
		return fmt.Errorf("telegram API status %d: %s", resp.StatusCode, buf.String())
	}
	return nil
}

// sendTelegramNotification invia notifiche a tutti gli utenti con notify=true e telegram_user impostato
func sendTelegramNotification(data *WeatherData) error {
	if telegramBotToken == "" {
		return fmt.Errorf(telegramTokenNotConfiguredMsg)
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

// ---------- Telegram registration helper (deep-link + polling) ----------

type telegramTokenDoc struct {
	Token     string    `bson:"token"`
	Username  string    `bson:"username"`
	CreatedAt time.Time `bson:"created_at"`
	Used      bool      `bson:"used"`
	ChatID    string    `bson:"chat_id,omitempty"`
}

// generateTelegramToken crea e salva un token per l'username
func generateTelegramToken(username string) (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	token := hex.EncodeToString(b)

	coll := MongoDB.Collection("telegram_tokens")
	ctx := context.Background()
	doc := telegramTokenDoc{Token: token, Username: username, CreatedAt: time.Now(), Used: false}
	if _, err := coll.InsertOne(ctx, doc); err != nil {
		return "", err
	}
	return token, nil
}

// getBotUsername queries Telegram getMe to obtain bot username (cached)
func getBotUsername() string {
	if telegramBotToken == "" {
		return ""
	}
	resp, err := http.Get(fmt.Sprintf("https://api.telegram.org/bot%s/getMe", telegramBotToken))
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	var out struct {
		Ok     bool `json:"ok"`
		Result struct {
			Username string `json:"username"`
		} `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return ""
	}
	return out.Result.Username
}

// generateTelegramTokenHandler HTTP: crea token e ritorna deep-link
func generateTelegramTokenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Username string `json:"username"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Username == "" {
		http.Error(w, "username richiesto", http.StatusBadRequest)
		return
	}
	// verify user exists
	u, err := GetUserByUsername(req.Username)
	if err != nil || u == nil {
		http.Error(w, "user non trovato", http.StatusBadRequest)
		return
	}
	token, err := generateTelegramToken(req.Username)
	if err != nil {
		http.Error(w, "errore generazione token", http.StatusInternalServerError)
		return
	}
	botUser := getBotUsername()
	link := ""
	if botUser != "" {
		// build t.me/<bot>?start=<token>
		link = fmt.Sprintf("https://t.me/%s?start=%s", url.PathEscape(botUser), token)
	} else {
		// fallback to telling user to send token to bot
		link = token
	}
	w.Header().Set(contentTypeHeader, contentTypeJSON)
	_ = json.NewEncoder(w).Encode(map[string]string{"token": token, "link": link})
}

// pollTelegramUpdates periodicamente chiama getUpdates e associa token->chat_id
func pollTelegramUpdates() {
	if telegramBotToken == "" {
		log.Println("telegram bot token non impostato: skip polling")
		return
	}
	log.Println("🔁 Avviato poller Telegram getUpdates")
	offset := 0
	client := &http.Client{Timeout: 30 * time.Second}
	for {
		out, err := fetchTelegramUpdates(client, offset)
		if err != nil {
			log.Printf("Errore getUpdates: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		for _, upd := range out.Result {
			offset = upd.UpdateId + 1
			processUpdate(upd)
		}

		// small sleep to avoid hammering if no results
		time.Sleep(500 * time.Millisecond)
	}
}

// fetchTelegramUpdates richiama l'API getUpdates e ritorna la struttura decodificata
func fetchTelegramUpdates(client *http.Client, offset int) (struct {
	Ok     bool `json:"ok"`
	Result []struct {
		UpdateId int `json:"update_id"`
		Message  struct {
			Chat struct {
				Id int64 `json:"id"`
			} `json:"chat"`
			Text string `json:"text"`
		} `json:"message"`
	} `json:"result"`
}, error) {
	var out struct {
		Ok     bool `json:"ok"`
		Result []struct {
			UpdateId int `json:"update_id"`
			Message  struct {
				Chat struct {
					Id int64 `json:"id"`
				} `json:"chat"`
				Text string `json:"text"`
			} `json:"message"`
		} `json:"result"`
	}

	urlStr := fmt.Sprintf("https://api.telegram.org/bot%s/getUpdates?offset=%d&timeout=20", telegramBotToken, offset)
	resp, err := client.Get(urlStr)
	if err != nil {
		return out, err
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	if err := json.Unmarshal(body, &out); err != nil {
		return out, fmt.Errorf("parsing getUpdates: %w body=%s", err, string(body))
	}
	return out, nil
}

// processUpdate gestisce un singolo update Telegram (attualmente solo /start token)
func processUpdate(upd struct {
	UpdateId int `json:"update_id"`
	Message  struct {
		Chat struct {
			Id int64 `json:"id"`
		} `json:"chat"`
		Text string `json:"text"`
	} `json:"message"`
}) {
	text := strings.TrimSpace(upd.Message.Text)
	if !strings.HasPrefix(text, "/start") {
		return
	}
	parts := strings.Fields(text)
	if len(parts) < 2 {
		return
	}
	token := parts[1]
	handleStartToken(token, upd.Message.Chat.Id)
}

// handleStartToken cerca il token nel DB e aggiorna l'utente con telegram_user+notify
func handleStartToken(token string, chatID int64) {
	coll := MongoDB.Collection("telegram_tokens")
	ctx := context.Background()
	var doc telegramTokenDoc
	if err := coll.FindOne(ctx, bson.M{"token": token, "used": false}).Decode(&doc); err != nil {
		log.Printf("Token non trovato o già usato: %s", token)
		return
	}

	chatIDStr := fmt.Sprintf("%d", chatID)
	userColl := MongoDB.Collection("users")
	if _, err := userColl.UpdateOne(ctx, bson.M{"username": doc.Username}, bson.M{"$set": bson.M{"telegram_user": chatIDStr, "notify": true}}); err != nil {
		log.Printf("Errore aggiornamento user telegram_user: %v", err)
		// try to notify user about failure (best-effort)
		_ = sendTelegramText(chatIDStr, "Errore durante la registrazione delle notifiche. Riprova più tardi.")
	} else {
		log.Printf("Registrato telegram_user per %s => %s", doc.Username, chatIDStr)
		// mark token used
		_, _ = coll.UpdateOne(ctx, bson.M{"token": token}, bson.M{"$set": bson.M{"used": true, "chat_id": chatIDStr}})
		// send confirmation message to user
		_ = sendTelegramText(chatIDStr, "✅ Registrazione completata! Riceverai le notifiche meteo su questo account.")
	}
}
