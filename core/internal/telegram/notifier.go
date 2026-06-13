package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

const (
	defaultTimeout   = 10 * time.Second
	sendMessagePath  = "https://api.telegram.org/bot%s/sendMessage"
	parseModeHTML    = "HTML"
)

type Config struct {
	BotToken string
	ChatID   string
}

func (c Config) Enabled() bool {
	return strings.TrimSpace(c.BotToken) != "" && strings.TrimSpace(c.ChatID) != ""
}

type TelegramNotifier struct {
	botToken string
	chatID   string
	enabled  bool
	client   *http.Client
}

func NewNotifier(cfg Config) *TelegramNotifier {
	return &TelegramNotifier{
		botToken: strings.TrimSpace(cfg.BotToken),
		chatID:   strings.TrimSpace(cfg.ChatID),
		enabled:  cfg.Enabled(),
		client: &http.Client{
			Timeout: defaultTimeout,
		},
	}
}

func (n *TelegramNotifier) Enabled() bool {
	return n.enabled
}

func (n *TelegramNotifier) SendMessage(ctx context.Context, message string) error {
	if !n.enabled {
		return fmt.Errorf("telegram notifier is not configured")
	}

	trimmed := strings.TrimSpace(message)
	if trimmed == "" {
		return fmt.Errorf("message is required")
	}

	payload := sendMessageRequest{
		ChatID:    n.chatID,
		Text:      trimmed,
		ParseMode: parseModeHTML,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal telegram payload: %w", err)
	}

	endpoint := fmt.Sprintf(sendMessagePath, n.botToken)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create telegram request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("telegram request failed: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read telegram response: %w", err)
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("telegram API error (%d): %s", resp.StatusCode, strings.TrimSpace(string(responseBody)))
	}

	var apiResponse sendMessageResponse
	if err := json.Unmarshal(responseBody, &apiResponse); err != nil {
		return fmt.Errorf("decode telegram response: %w", err)
	}
	if !apiResponse.OK {
		return fmt.Errorf("telegram API returned ok=false: %s", apiResponse.Description)
	}

	log.Printf("telegram: message sent to chat %s", n.chatID)
	return nil
}

// SendImportantNews sends a formatted alert when important news is detected.
// Example: SendImportantNews(ctx, "MU", "AI/HBM report")
func (n *TelegramNotifier) SendImportantNews(ctx context.Context, symbol, headline string) error {
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	headline = strings.TrimSpace(headline)
	if symbol == "" || headline == "" {
		return fmt.Errorf("symbol and headline are required")
	}

	message := fmt.Sprintf("🚨 Important News: %s mentioned in %s", EscapeHTML(symbol), EscapeHTML(headline))
	return n.SendMessage(ctx, message)
}

// SendAlert sends a generic stock monitor alert to Telegram.
func (n *TelegramNotifier) SendAlert(ctx context.Context, title, body string) error {
	title = strings.TrimSpace(title)
	body = strings.TrimSpace(body)
	if title == "" {
		return fmt.Errorf("title is required")
	}

	var message string
	if body == "" {
		message = fmt.Sprintf("🚨 <b>%s</b>", EscapeHTML(title))
	} else {
		message = fmt.Sprintf("🚨 <b>%s</b>\n%s", EscapeHTML(title), EscapeHTML(body))
	}

	return n.SendMessage(ctx, message)
}

func EscapeHTML(value string) string {
	replacer := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
	)
	return replacer.Replace(value)
}

type sendMessageRequest struct {
	ChatID    string `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode"`
}

type sendMessageResponse struct {
	OK          bool   `json:"ok"`
	Description string `json:"description"`
}
