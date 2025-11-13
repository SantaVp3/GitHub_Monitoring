package notify

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github-monitor/db/models"
)

// Message represents a notification message
type Message struct {
	Title   string
	Content string
	URL     string
}

// Notifier interface for different notification types
type Notifier interface {
	Send(config *models.NotificationConfig, message Message) error
}

// WeCom implements企业微信notification
type WeCom struct{}

func (w *WeCom) Send(config *models.NotificationConfig, message Message) error {
	payload := map[string]interface{}{
		"msgtype": "markdown",
		"markdown": map[string]string{
			"content": fmt.Sprintf("## %s\n\n%s\n\n[查看详情](%s)", message.Title, message.Content, message.URL),
		},
	}

	return sendWebhook(config.WebhookURL, payload)
}

// DingTalk implements钉钉notification
type DingTalk struct{}

func (d *DingTalk) Send(config *models.NotificationConfig, message Message) error {
	timestamp := time.Now().UnixMilli()
	sign := ""

	if config.Secret != "" {
		sign = generateDingTalkSign(config.Secret, timestamp)
	}

	url := config.WebhookURL
	if sign != "" {
		url = fmt.Sprintf("%s&timestamp=%d&sign=%s", url, timestamp, sign)
	}

	payload := map[string]interface{}{
		"msgtype": "markdown",
		"markdown": map[string]interface{}{
			"title": message.Title,
			"text":  fmt.Sprintf("## %s\n\n%s\n\n[查看详情](%s)", message.Title, message.Content, message.URL),
		},
	}

	return sendWebhook(url, payload)
}

func generateDingTalkSign(secret string, timestamp int64) string {
	stringToSign := fmt.Sprintf("%d\n%s", timestamp, secret)
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(stringToSign))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// Feishu implements飞书notification
type Feishu struct{}

func (f *Feishu) Send(config *models.NotificationConfig, message Message) error {
	timestamp := time.Now().Unix()
	sign := ""

	if config.Secret != "" {
		sign = generateFeishuSign(config.Secret, timestamp)
	}

	payload := map[string]interface{}{
		"msg_type": "interactive",
		"card": map[string]interface{}{
			"header": map[string]interface{}{
				"title": map[string]string{
					"tag":     "plain_text",
					"content": message.Title,
				},
				"template": "red",
			},
			"elements": []interface{}{
				map[string]interface{}{
					"tag":  "div",
					"text": map[string]string{
						"tag":     "lark_md",
						"content": message.Content,
					},
				},
				map[string]interface{}{
					"tag": "action",
					"actions": []interface{}{
						map[string]interface{}{
							"tag": "button",
							"text": map[string]string{
								"tag":     "plain_text",
								"content": "查看详情",
							},
							"type": "primary",
							"url":  message.URL,
						},
					},
				},
			},
		},
	}

	if sign != "" {
		payload["timestamp"] = fmt.Sprintf("%d", timestamp)
		payload["sign"] = sign
	}

	return sendWebhook(config.WebhookURL, payload)
}

func generateFeishuSign(secret string, timestamp int64) string {
	stringToSign := fmt.Sprintf("%d\n%s", timestamp, secret)
	h := hmac.New(sha256.New, []byte(stringToSign))
	h.Write([]byte(stringToSign))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// Webhook implements generic webhook notification
type Webhook struct{}

func (wh *Webhook) Send(config *models.NotificationConfig, message Message) error {
	payload := map[string]interface{}{
		"title":   message.Title,
		"content": message.Content,
		"url":     message.URL,
		"time":    time.Now().Format(time.RFC3339),
	}

	return sendWebhook(config.WebhookURL, payload)
}

// sendWebhook sends a POST request to the webhook URL
func sendWebhook(url string, payload interface{}) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("webhook returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// GetNotifier returns the appropriate notifier based on type
func GetNotifier(notifType string) Notifier {
	switch notifType {
	case "wecom":
		return &WeCom{}
	case "dingtalk":
		return &DingTalk{}
	case "feishu":
		return &Feishu{}
	case "webhook":
		return &Webhook{}
	default:
		return &Webhook{}
	}
}

// SendNotification sends a notification using the specified config
func SendNotification(config *models.NotificationConfig, message Message) error {
	if !config.Enabled {
		return nil // Skip if disabled
	}

	notifier := GetNotifier(config.Type)
	return notifier.Send(config, message)
}
