// Package twilio provides a Twilio SMS provider for gosms.
package twilio

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	gosms "github.com/KARTIKrocks/gosms"
)

// messageIDRegex validates Twilio message SIDs (alphanumeric, no path separators).
var messageIDRegex = regexp.MustCompile(`^[A-Za-z0-9_]+$`)

const defaultBaseURL = "https://api.twilio.com/2010-04-01"

// Config holds Twilio-specific configuration.
type Config struct {
	// AccountSID is the Twilio account SID.
	AccountSID string

	// AuthToken is the Twilio auth token.
	AuthToken string

	// From is the default sender phone number or messaging service SID.
	From string

	// MessagingServiceSID is the messaging service SID (optional).
	MessagingServiceSID string

	// StatusCallback is the URL for delivery status webhooks.
	StatusCallback string

	// HTTPClient is a custom HTTP client (optional).
	HTTPClient *http.Client

	// BaseURL overrides the API base URL (for testing).
	BaseURL string
}

// Provider implements the gosms.Provider interface for Twilio.
type Provider struct {
	config Config
}

// NewProvider creates a new Twilio provider.
func NewProvider(config Config) (*Provider, error) {
	if config.AccountSID == "" || config.AuthToken == "" {
		return nil, fmt.Errorf("%w: account SID and auth token required", gosms.ErrInvalidConfig)
	}

	if config.HTTPClient == nil {
		config.HTTPClient = &http.Client{Timeout: 30 * time.Second}
	}

	if config.BaseURL == "" {
		config.BaseURL = defaultBaseURL
	}

	return &Provider{config: config}, nil
}

// Name returns the provider name.
func (p *Provider) Name() string {
	return "twilio"
}

// Send sends an SMS message via Twilio.
func (p *Provider) Send(ctx context.Context, msg *gosms.Message) (*gosms.Result, error) {
	endpoint := fmt.Sprintf("%s/Accounts/%s/Messages.json", p.config.BaseURL, p.config.AccountSID)

	data := url.Values{}
	data.Set("To", msg.To)
	data.Set("Body", msg.Body)

	from := msg.From
	if from == "" {
		from = p.config.From
	}

	switch {
	case p.config.MessagingServiceSID != "":
		data.Set("MessagingServiceSid", p.config.MessagingServiceSID)
	case from != "":
		data.Set("From", from)
	default:
		return nil, fmt.Errorf("%w: sender (From) is required", gosms.ErrInvalidConfig)
	}

	if p.config.StatusCallback != "" {
		data.Set("StatusCallback", p.config.StatusCallback)
	}

	if msg.ScheduledAt != nil {
		data.Set("SendAt", msg.ScheduledAt.UTC().Format(time.RFC3339))
		data.Set("ScheduleType", "fixed")
	}

	if msg.ValidityPeriod > 0 {
		seconds := min(int(msg.ValidityPeriod.Seconds()), 14400)
		data.Set("ValidityPeriod", fmt.Sprintf("%d", seconds))
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(p.config.AccountSID, p.config.AuthToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := p.config.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", gosms.ErrSendFailed, err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}

	var twilioResp messageResponse
	if err := json.Unmarshal(body, &twilioResp); err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, parseError(twilioResp)
	}

	return &gosms.Result{
		MessageID: twilioResp.SID,
		To:        twilioResp.To,
		Status:    mapStatus(twilioResp.Status),
		Provider:  p.Name(),
		Cost:      twilioResp.Price,
		Currency:  twilioResp.PriceUnit,
		Segments:  twilioResp.NumSegments,
		SentAt:    time.Now(),
		Raw: map[string]any{
			"sid":          twilioResp.SID,
			"status":       twilioResp.Status,
			"num_segments": twilioResp.NumSegments,
		},
	}, nil
}

// SendBulk sends multiple SMS messages.
func (p *Provider) SendBulk(ctx context.Context, msgs []*gosms.Message) ([]*gosms.Result, error) {
	return gosms.SendEach(ctx, p.Name(), msgs, p.Send), nil
}

// GetStatus retrieves the delivery status of a message.
func (p *Provider) GetStatus(ctx context.Context, messageID string) (*gosms.Status, error) {
	if !messageIDRegex.MatchString(messageID) {
		return nil, fmt.Errorf("%w: invalid message ID format", gosms.ErrInvalidConfig)
	}
	endpoint := fmt.Sprintf("%s/Accounts/%s/Messages/%s.json", p.config.BaseURL, p.config.AccountSID, messageID)

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(p.config.AccountSID, p.config.AuthToken)

	resp, err := p.config.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}

	var twilioResp messageResponse
	if err := json.Unmarshal(body, &twilioResp); err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, parseError(twilioResp)
	}

	return &gosms.Status{
		MessageID:    twilioResp.SID,
		Status:       mapStatus(twilioResp.Status),
		UpdatedAt:    time.Now(),
		ErrorCode:    formatErrorCode(twilioResp.ErrorCode),
		ErrorMessage: twilioResp.ErrorMessage,
		Raw: map[string]any{
			"status":        twilioResp.Status,
			"error_code":    twilioResp.ErrorCode,
			"error_message": twilioResp.ErrorMessage,
		},
	}, nil
}

// ParseWebhook parses a Twilio status callback webhook.
func ParseWebhook(r *http.Request) (*gosms.Status, error) {
	if err := r.ParseForm(); err != nil {
		return nil, err
	}

	return &gosms.Status{
		MessageID:    r.FormValue("MessageSid"),
		Status:       mapStatus(r.FormValue("MessageStatus")),
		UpdatedAt:    time.Now(),
		ErrorCode:    r.FormValue("ErrorCode"),
		ErrorMessage: r.FormValue("ErrorMessage"),
		Raw: map[string]any{
			"account_sid":    r.FormValue("AccountSid"),
			"from":           r.FormValue("From"),
			"to":             r.FormValue("To"),
			"message_status": r.FormValue("MessageStatus"),
		},
	}, nil
}

type messageResponse struct {
	SID          string `json:"sid"`
	To           string `json:"to"`
	From         string `json:"from"`
	Body         string `json:"body"`
	Status       string `json:"status"`
	NumSegments  int    `json:"num_segments,string"`
	Price        string `json:"price"`
	PriceUnit    string `json:"price_unit"`
	ErrorCode    int    `json:"error_code"`
	ErrorMessage string `json:"error_message"`
	Code         int    `json:"code"`
	Message      string `json:"message"`
}

func mapStatus(status string) gosms.DeliveryStatus {
	switch status {
	case "queued":
		return gosms.StatusQueued
	case "accepted":
		return gosms.StatusAccepted
	case "sending", "sent":
		return gosms.StatusSent
	case "delivered":
		return gosms.StatusDelivered
	case "undelivered", "failed":
		return gosms.StatusFailed
	case "canceled":
		return gosms.StatusRejected
	default:
		return gosms.StatusUnknown
	}
}

func formatErrorCode(code int) string {
	if code == 0 {
		return ""
	}
	return fmt.Sprintf("%d", code)
}

func parseError(resp messageResponse) error {
	switch resp.Code {
	case 21211, 21614:
		return fmt.Errorf("%w: %s", gosms.ErrInvalidPhone, resp.Message)
	case 21408:
		return fmt.Errorf("%w: %s", gosms.ErrBlacklisted, resp.Message)
	case 20003:
		return fmt.Errorf("%w: %s", gosms.ErrInsufficientFunds, resp.Message)
	case 20429:
		return fmt.Errorf("%w: %s", gosms.ErrRateLimited, resp.Message)
	default:
		return fmt.Errorf("%w: [%d] %s", gosms.ErrProviderError, resp.Code, resp.Message)
	}
}
