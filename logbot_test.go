package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

// --- handleConviction ---

func TestHandleConviction_MethodNotAllowed(t *testing.T) {
	for _, method := range []string{http.MethodGet, http.MethodPut, http.MethodDelete} {
		req := httptest.NewRequest(method, "/convinced", nil)
		w := httptest.NewRecorder()
		handleConviction(w, req)
		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("%s: expected 405, got %d", method, w.Code)
		}
	}
}

func TestHandleConviction_InvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/convinced", strings.NewReader("not json"))
	w := httptest.NewRecorder()
	handleConviction(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandleConviction_EmptyBody(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/convinced", strings.NewReader(""))
	w := httptest.NewRecorder()
	handleConviction(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandleConviction_Success(t *testing.T) {
	fakeTelegram := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	}))
	defer fakeTelegram.Close()

	old := telegramAPIBase
	telegramAPIBase = fakeTelegram.URL
	defer func() { telegramAPIBase = old }()

	os.Setenv("TELEGRAM_BOT_TOKEN", "test-token")
	os.Setenv("TELEGRAM_CHAT_ID", "123456")
	defer os.Unsetenv("TELEGRAM_BOT_TOKEN")
	defer os.Unsetenv("TELEGRAM_CHAT_ID")

	payload := ConvictionPayload{FullName: "Іван Іванов", ConvincedAt: "2024-01-01T10:00:00Z"}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/convinced", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handleConviction(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "logged") {
		t.Errorf("expected body to contain 'logged', got: %s", w.Body.String())
	}
}

func TestHandleConviction_TelegramFailure(t *testing.T) {
	fakeTelegram := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer fakeTelegram.Close()

	old := telegramAPIBase
	telegramAPIBase = fakeTelegram.URL
	defer func() { telegramAPIBase = old }()

	os.Setenv("TELEGRAM_BOT_TOKEN", "bad-token")
	os.Setenv("TELEGRAM_CHAT_ID", "0")
	defer os.Unsetenv("TELEGRAM_BOT_TOKEN")
	defer os.Unsetenv("TELEGRAM_CHAT_ID")

	payload := ConvictionPayload{FullName: "Test", ConvincedAt: "2024-01-01T10:00:00Z"}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/convinced", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	// sendTelegramMessage does not check HTTP status from Telegram, so we expect 200
	handleConviction(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 (Telegram status not checked), got %d", w.Code)
	}
}

// --- sendTelegramMessage ---

func TestSendTelegramMessage_SendsCorrectPayload(t *testing.T) {
	var receivedBody map[string]string
	var receivedPath string

	fakeTelegram := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedPath = r.URL.Path
		json.NewDecoder(r.Body).Decode(&receivedBody)
		w.WriteHeader(http.StatusOK)
	}))
	defer fakeTelegram.Close()

	old := telegramAPIBase
	telegramAPIBase = fakeTelegram.URL
	defer func() { telegramAPIBase = old }()

	os.Setenv("TELEGRAM_BOT_TOKEN", "mytoken")
	os.Setenv("TELEGRAM_CHAT_ID", "999")
	defer os.Unsetenv("TELEGRAM_BOT_TOKEN")
	defer os.Unsetenv("TELEGRAM_CHAT_ID")

	err := sendTelegramMessage("привіт")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(receivedPath, "mytoken") {
		t.Errorf("expected token in path, got: %s", receivedPath)
	}
	if receivedBody["chat_id"] != "999" {
		t.Errorf("expected chat_id=999, got: %s", receivedBody["chat_id"])
	}
	if receivedBody["text"] != "привіт" {
		t.Errorf("expected text='привіт', got: %s", receivedBody["text"])
	}
}

func TestSendTelegramMessage_NetworkError(t *testing.T) {
	old := telegramAPIBase
	telegramAPIBase = "http://127.0.0.1:1" // нічого не слухає
	defer func() { telegramAPIBase = old }()

	err := sendTelegramMessage("test")
	if err == nil {
		t.Error("expected error on network failure, got nil")
	}
}
