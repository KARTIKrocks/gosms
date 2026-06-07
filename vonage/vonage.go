// Package vonage provides a Vonage (Nexmo) SMS provider for gosms.
package vonage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	gosms "github.com/KARTIKrocks/gosms"
)

const defaultBaseURL = "https://rest.nexmo.com"

// MessageType represents the type of message.
type MessageType string

const (
	TypeText    MessageType = "text"
	TypeUnicode MessageType = "unicode"
	TypeBinary  MessageType = "binary"
)

// Config holds Vonage-specific configuration.
type Config struct {
	// APIKey is the Vonage API key.
	APIKey string

	// APISecret is the Vonage API secret.
	APISecret string

	// From is the default sender ID or phone number.
	From string

	// Type is the message type (text or unicode).
	Type MessageType

	// TTL is the message time-to-live in milliseconds.
	TTL int

	// StatusReportRequired enables delivery receipts.
	StatusReportRequired bool

	// CallbackURL is the URL for delivery receipts.
	CallbackURL string

	// HTTPClient is a custom HTTP client (optional).
	HTTPClient *http.Client

	// BaseURL overrides the API base URL (for testing).
	BaseURL string
}

// Provider implements the gosms.Provider interface for Vonage.
type Provider struct {
	config Config
}

// NewProvider creates a new Vonage provider.
func NewProvider(config Config) (*Provider, error) {
	if config.APIKey == "" || config.APISecret == "" {
		return nil, fmt.Errorf("%w: API key and secret required", gosms.ErrInvalidConfig)
	}

	if config.HTTPClient == nil {
		config.HTTPClient = &http.Client{Timeout: 30 * time.Second}
	}

	if config.BaseURL == "" {
		config.BaseURL = defaultBaseURL
	}

	if config.Type == "" {
		config.Type = TypeText
	}

	return &Provider{config: config}, nil
}

// Name returns the provider name.
func (p *Provider) Name() string {
	return "vonage"
}

// Send sends an SMS message via Vonage.
func (p *Provider) Send(ctx context.Context, msg *gosms.Message) (*gosms.Result, error) {
	endpoint := fmt.Sprintf("%s/sms/json", p.config.BaseURL)

	from := msg.From
	if from == "" {
		from = p.config.From
	}
	if from == "" {
		return nil, fmt.Errorf("%w: sender (From) is required", gosms.ErrInvalidConfig)
	}

	reqBody := sendRequest{
		APIKey:    p.config.APIKey,
		APISecret: p.config.APISecret,
		To:        msg.To,
		From:      from,
		Text:      msg.Body,
		Type:      string(p.config.Type),
	}

	if p.config.TTL > 0 {
		reqBody.TTL = p.config.TTL
	}

	if p.config.StatusReportRequired {
		reqBody.StatusReportReq = "1"
	}

	if p.config.CallbackURL != "" {
		reqBody.CallbackURL = p.config.CallbackURL
	}

	if msg.Reference != "" {
		reqBody.ClientRef = msg.Reference
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := p.config.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", gosms.ErrSendFailed, err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}

	var vonageResp sendResponse
	if err := json.Unmarshal(body, &vonageResp); err != nil {
		return nil, err
	}

	if len(vonageResp.Messages) == 0 {
		return nil, fmt.Errorf("%w: no response from Vonage", gosms.ErrSendFailed)
	}

	message := vonageResp.Messages[0]

	if message.Status != "0" {
		return nil, parseError(message.Status, message.ErrorText)
	}

	return &gosms.Result{
		MessageID: message.MessageID,
		To:        message.To,
		Status:    gosms.StatusAccepted,
		Provider:  p.Name(),
		Cost:      message.MessagePrice,
		Currency:  message.Currency,
		Segments:  vonageResp.MessageCount,
		SentAt:    time.Now(),
		Raw: map[string]any{
			"message_id":        message.MessageID,
			"status":            message.Status,
			"remaining_balance": message.RemainingBalance,
			"network":           message.Network,
		},
	}, nil
}

// SendBulk sends multiple SMS messages.
func (p *Provider) SendBulk(ctx context.Context, msgs []*gosms.Message) ([]*gosms.Result, error) {
	return gosms.SendEach(ctx, p.Name(), msgs, p.Send), nil
}

// GetStatus retrieves the delivery status of a message.
// Note: Vonage uses delivery receipts (webhooks) for status updates.
func (p *Provider) GetStatus(_ context.Context, _ string) (*gosms.Status, error) {
	return nil, fmt.Errorf("%w: Vonage uses webhooks for delivery status", gosms.ErrUnsupported)
}

// ParseWebhook parses a Vonage delivery receipt webhook.
func ParseWebhook(r *http.Request) (*gosms.Status, error) {
	if r.Method == http.MethodGet {
		return parseQueryParams(r), nil
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		return nil, err
	}

	var dlr deliveryReceipt
	if err := json.Unmarshal(body, &dlr); err != nil {
		return nil, fmt.Errorf("failed to parse webhook body: %w", err)
	}

	return &gosms.Status{
		MessageID: dlr.MessageID,
		Status:    mapStatus(dlr.Status),
		UpdatedAt: time.Now(),
		ErrorCode: dlr.ErrCode,
		Raw: map[string]any{
			"msisdn":            dlr.MSISDN,
			"to":                dlr.To,
			"network":           dlr.NetworkCode,
			"price":             dlr.Price,
			"scts":              dlr.SCTS,
			"message_timestamp": dlr.MessageTimestamp,
		},
	}, nil
}

func parseQueryParams(r *http.Request) *gosms.Status {
	return &gosms.Status{
		MessageID: r.URL.Query().Get("messageId"),
		Status:    mapStatus(r.URL.Query().Get("status")),
		UpdatedAt: time.Now(),
		ErrorCode: r.URL.Query().Get("err-code"),
		Raw: map[string]any{
			"msisdn":  r.URL.Query().Get("msisdn"),
			"to":      r.URL.Query().Get("to"),
			"network": r.URL.Query().Get("network-code"),
			"price":   r.URL.Query().Get("price"),
		},
	}
}

type sendRequest struct {
	APIKey          string `json:"api_key"`
	APISecret       string `json:"api_secret"`
	To              string `json:"to"`
	From            string `json:"from"`
	Text            string `json:"text"`
	Type            string `json:"type,omitempty"`
	TTL             int    `json:"ttl,omitempty"`
	StatusReportReq string `json:"status-report-req,omitempty"`
	CallbackURL     string `json:"callback,omitempty"`
	ClientRef       string `json:"client-ref,omitempty"`
}

type sendResponse struct {
	MessageCount int `json:"message-count,string"`
	Messages     []struct {
		To               string `json:"to"`
		MessageID        string `json:"message-id"`
		Status           string `json:"status"`
		RemainingBalance string `json:"remaining-balance"`
		MessagePrice     string `json:"message-price"`
		Currency         string `json:"currency,omitempty"`
		Network          string `json:"network"`
		ErrorText        string `json:"error-text,omitempty"`
	} `json:"messages"`
}

type deliveryReceipt struct {
	MSISDN           string `json:"msisdn"`
	To               string `json:"to"`
	NetworkCode      string `json:"network-code"`
	MessageID        string `json:"messageId"`
	Price            string `json:"price"`
	Status           string `json:"status"`
	SCTS             string `json:"scts"`
	ErrCode          string `json:"err-code"`
	MessageTimestamp string `json:"message-timestamp"`
}

func mapStatus(status string) gosms.DeliveryStatus {
	switch status {
	case "delivered":
		return gosms.StatusDelivered
	case "accepted":
		return gosms.StatusAccepted
	case "buffered":
		return gosms.StatusQueued
	case "failed":
		return gosms.StatusFailed
	case "expired":
		return gosms.StatusExpired
	case "rejected":
		return gosms.StatusRejected
	default:
		return gosms.StatusUnknown
	}
}

func parseError(status, errorText string) error {
	switch status {
	case "1":
		return fmt.Errorf("%w: throttled - %s", gosms.ErrRateLimited, errorText)
	case "2":
		return fmt.Errorf("%w: missing params - %s", gosms.ErrInvalidConfig, errorText)
	case "3":
		return fmt.Errorf("%w: invalid params - %s", gosms.ErrInvalidConfig, errorText)
	case "4":
		return fmt.Errorf("%w: invalid credentials", gosms.ErrInvalidConfig)
	case "5":
		return fmt.Errorf("%w: internal error - %s", gosms.ErrProviderError, errorText)
	case "6":
		return fmt.Errorf("%w: invalid message - %s", gosms.ErrInvalidMessage, errorText)
	case "7":
		return fmt.Errorf("%w: number barred - %s", gosms.ErrBlacklisted, errorText)
	case "8":
		return fmt.Errorf("%w: partner account barred", gosms.ErrInvalidConfig)
	case "9":
		return fmt.Errorf("%w: partner quota exceeded", gosms.ErrInsufficientFunds)
	case "15":
		return fmt.Errorf("%w: invalid sender - %s", gosms.ErrInvalidConfig, errorText)
	default:
		return fmt.Errorf("%w: [%s] %s", gosms.ErrProviderError, status, errorText)
	}
}
