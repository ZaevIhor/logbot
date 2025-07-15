package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type ConvictionPayload struct {
	FullName    string `json:"full_name"`
	ConvincedAt string `json:"convinced_at"`
}

type TelegramMessage struct {
	ChatID string `json:"chat_id"`
	Text   string `json:"text"`
}

func main() {
	http.HandleFunc("/convinced", handleConviction)

	port := ":8080"
	fmt.Println("🚀 Listening on", port)
	log.Fatal(http.ListenAndServe(port, nil))
}

func handleConviction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload ConvictionPayload
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	message := fmt.Sprintf("🧠 Користувача переконали!\n👤 %s\n⏰ %s", payload.FullName, payload.ConvincedAt)
	err = sendTelegramMessage(message)
	if err != nil {
		http.Error(w, "Failed to send Telegram message", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"logged"}`))
}

func sendTelegramMessage(msg string) error {
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	chatID := os.Getenv("TELEGRAM_CHAT_ID")

	telegramURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botToken)

	payload := TelegramMessage{
		ChatID: chatID,
		Text:   msg,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := http.Post(telegramURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
