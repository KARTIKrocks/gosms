// Package msg91 provides an MSG91 SMS provider for gosms.
//
// MSG91 is an SMS gateway popular in India and South-East Asia. This provider
// targets the MSG91 Flow API (v5), which is the DLT-compliant path required
// for sending SMS to Indian recipients.
//
// Template variables are passed via reserved keys on [gosms.Message.Metadata].
// Use [SetVar] and [SetTemplateID] rather than writing the keys directly.
//
// The OTP endpoints (verify/retry) are exposed as provider-specific methods
// on [*Provider] and are not part of the [gosms.Provider] interface, since
// OTP verification is not a concept other providers share.
package msg91

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	gosms "github.com/KARTIKrocks/gosms"
)

const (
	defaultBaseURL = "https://control.msg91.com"
	defaultCountry = "91"

	flowPath      = "/api/v5/flow"
	otpVerifyPath = "/api/v5/otp/verify"
	otpRetryPath  = "/api/v5/otp/retry"

	// metaVarPrefix is the Metadata key prefix for Flow template variables.
	metaVarPrefix = "msg91.var."
	// metaTemplateID overrides Config.TemplateID for a single message.
	metaTemplateID = "msg91.template_id"
)

// Route identifies the MSG91 message route.
type Route string

const (
	// RouteTransactional is MSG91 route 4 (OTP, alerts, DLT transactional).
	RouteTransactional Route = "4"
	// RoutePromotional is MSG91 route 1 (marketing).
	RoutePromotional Route = "1"
)

// Config holds MSG91-specific configuration.
type Config struct {
	// AuthKey is the MSG91 auth key sent as the `authkey` header.
	AuthKey string

	// SenderID is the 6-character DLT-registered sender ID. Used as the
	// default `From` when a Message does not specify one.
	SenderID string

	// TemplateID is the default DLT-approved Flow template ID. Can be
	// overridden per-message via [SetTemplateID].
	TemplateID string

	// Route selects the MSG91 route (transactional or promotional).
	// Defaults to RouteTransactional.
	Route Route

	// Country is the default country code (digits only, no `+`) prepended
	// to recipient numbers that don't already include one. Defaults to "91".
	Country string

	// ShortURL enables MSG91's URL shortening. Defaults to false.
	ShortURL bool

	// HTTPClient is a custom HTTP client (optional).
	HTTPClient *http.Client

	// BaseURL overrides the API base URL (for testing).
	BaseURL string
}

// Provider implements the [gosms.Provider] interface for MSG91.
type Provider struct {
	config Config
}

// NewProvider creates a new MSG91 provider.
func NewProvider(config Config) (*Provider, error) {
	if config.AuthKey == "" {
		return nil, fmt.Errorf("%w: auth key required", gosms.ErrInvalidConfig)
	}

	if config.HTTPClient == nil {
		config.HTTPClient = &http.Client{Timeout: 30 * time.Second}
	}

	if config.BaseURL == "" {
		config.BaseURL = defaultBaseURL
	}
	config.BaseURL = strings.TrimRight(config.BaseURL, "/")

	if config.Country == "" {
		config.Country = defaultCountry
	}

	if config.Route == "" {
		config.Route = RouteTransactional
	}

	return &Provider{config: config}, nil
}

// Name returns the provider name.
func (p *Provider) Name() string {
	return "msg91"
}

// Send sends an SMS message via the MSG91 Flow API.
//
// The template ID is resolved from [SetTemplateID] on the message (if set),
// falling back to [Config.TemplateID]. Template variables set via [SetVar]
// are included in the recipient payload.
//
// The following [gosms.Message] fields are not supported by MSG91 Flow and
// are ignored: Reference, ValidityPeriod, ScheduledAt.
func (p *Provider) Send(ctx context.Context, msg *gosms.Message) (*gosms.Result, error) {
	templateID := templateIDFor(msg, p.config.TemplateID)
	if templateID == "" {
		return nil, fmt.Errorf("%w: template ID required", gosms.ErrInvalidConfig)
	}

	sender := msg.From
	if sender == "" {
		sender = p.config.SenderID
	}

	results, err := p.sendFlow(ctx, templateID, sender, []*gosms.Message{msg})
	if err != nil {
		return nil, err
	}
	return results[0], nil
}

// SendBulk sends multiple messages using MSG91's native multi-recipient Flow
// call. Messages are grouped by their effective template ID (per-message
// override or Config default) and each group is sent as a single API call.
// All recipients in a group share the same MSG91 request_id.
func (p *Provider) SendBulk(ctx context.Context, msgs []*gosms.Message) ([]*gosms.Result, error) {
	if len(msgs) == 0 {
		return nil, nil
	}

	type bucket struct {
		templateID string
		sender     string
		msgs       []*gosms.Message
		indices    []int
	}

	buckets := make(map[string]*bucket)
	keyOrder := make([]string, 0)

	for i, msg := range msgs {
		tid := templateIDFor(msg, p.config.TemplateID)
		sender := msg.From
		if sender == "" {
			sender = p.config.SenderID
		}
		key := tid + "\x00" + sender

		b, ok := buckets[key]
		if !ok {
			b = &bucket{templateID: tid, sender: sender}
			buckets[key] = b
			keyOrder = append(keyOrder, key)
		}
		b.msgs = append(b.msgs, msg)
		b.indices = append(b.indices, i)
	}

	results := make([]*gosms.Result, len(msgs))
	for _, key := range keyOrder {
		b := buckets[key]

		if b.templateID == "" {
			err := fmt.Errorf("%w: template ID required", gosms.ErrInvalidConfig)
			for _, idx := range b.indices {
				results[idx] = failedResult(msgs[idx], p.Name(), err)
			}
			continue
		}

		sent, err := p.sendFlow(ctx, b.templateID, b.sender, b.msgs)
		if err != nil {
			for _, idx := range b.indices {
				results[idx] = failedResult(msgs[idx], p.Name(), err)
			}
			continue
		}
		for i, idx := range b.indices {
			if i < len(sent) {
				results[idx] = sent[i]
			} else {
				results[idx] = failedResult(msgs[idx], p.Name(),
					fmt.Errorf("%w: provider returned fewer results than recipients", gosms.ErrProviderError))
			}
		}
	}

	return results, nil
}

// GetStatus is not supported via the Flow API; MSG91 delivers status via
// webhooks. Use [ParseWebhook] on your callback handler instead.
func (p *Provider) GetStatus(_ context.Context, _ string) (*gosms.Status, error) {
	return nil, fmt.Errorf("%w: MSG91 uses webhooks for delivery status (see ParseWebhook)", gosms.ErrUnsupported)
}

// sendFlow performs a single Flow API call with one or more recipients.
func (p *Provider) sendFlow(ctx context.Context, templateID, sender string, msgs []*gosms.Message) ([]*gosms.Result, error) {
	recipients := make([]map[string]any, 0, len(msgs))
	for _, msg := range msgs {
		r := map[string]any{"mobiles": p.normalizeRecipient(msg.To)}

		vars := extractVars(msg)
		for k, v := range vars {
			r[k] = v
		}
		// Fallback: expose Body as `body` when no vars are set. Templates
		// using a single ##body## placeholder can consume this directly.
		if len(vars) == 0 && msg.Body != "" {
			r["body"] = msg.Body
		}
		recipients = append(recipients, r)
	}

	reqBody := flowRequest{
		TemplateID: templateID,
		SenderID:   sender,
		Route:      string(p.config.Route),
		Recipients: recipients,
		RealTime:   "1",
	}
	if p.config.ShortURL {
		reqBody.ShortURL = "1"
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.config.BaseURL+flowPath, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("authkey", p.config.AuthKey)

	resp, err := p.config.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", gosms.ErrSendFailed, err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}

	var flow flowResponse
	if err := json.Unmarshal(body, &flow); err != nil {
		if resp.StatusCode >= 400 {
			return nil, fmt.Errorf("%w: http %d: %s", gosms.ErrProviderError, resp.StatusCode, bodySnippet(body))
		}
		return nil, fmt.Errorf("%w: decode response: %w", gosms.ErrProviderError, err)
	}

	if resp.StatusCode >= 400 || flow.Type != "success" {
		return nil, parseFlowError(flow)
	}

	now := time.Now()
	out := make([]*gosms.Result, len(msgs))
	for i, msg := range msgs {
		out[i] = &gosms.Result{
			MessageID: flow.Message,
			To:        msg.To,
			Status:    gosms.StatusAccepted,
			Provider:  p.Name(),
			Segments:  gosms.CalculateSegments(msg.Body),
			SentAt:    now,
			Raw: map[string]any{
				"request_id": flow.Message,
				"type":       flow.Type,
			},
		}
	}
	return out, nil
}

// VerifyOTP verifies a one-time password for the given phone number using
// the MSG91 OTP API. This is a provider-specific method and is not part of
// the [gosms.Provider] interface.
func (p *Provider) VerifyOTP(ctx context.Context, phone, otp string) (*VerifyResult, error) {
	if phone == "" || otp == "" {
		return nil, fmt.Errorf("%w: phone and otp required", gosms.ErrInvalidConfig)
	}

	q := url.Values{}
	q.Set("mobile", p.normalizeRecipient(phone))
	q.Set("otp", otp)

	endpoint := p.config.BaseURL + otpVerifyPath + "?" + q.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("authkey", p.config.AuthKey)

	resp, err := p.config.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", gosms.ErrSendFailed, err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}

	var otpResp otpResponse
	if err := json.Unmarshal(body, &otpResp); err != nil {
		if resp.StatusCode >= 400 {
			return nil, fmt.Errorf("%w: http %d: %s", gosms.ErrProviderError, resp.StatusCode, bodySnippet(body))
		}
		return nil, fmt.Errorf("%w: decode response: %w", gosms.ErrProviderError, err)
	}

	// Transport-level failures (auth, rate limit, server error) are
	// distinct from a mismatched OTP — surface them as errors so callers
	// can tell "wrong code" from "bad authkey".
	if resp.StatusCode >= 500 {
		return nil, parseFlowError(flowResponse(otpResp))
	}
	if resp.StatusCode >= 400 && !isOTPMismatch(otpResp.Message) {
		return nil, parseFlowError(flowResponse(otpResp))
	}

	return &VerifyResult{
		Verified: otpResp.Type == "success",
		Message:  otpResp.Message,
		Raw: map[string]any{
			"type":    otpResp.Type,
			"message": otpResp.Message,
		},
	}, nil
}

// isOTPMismatch returns true when MSG91's message indicates a wrong or
// expired OTP, rather than a transport/auth/config failure. These are the
// only 4xx cases that should return Verified:false instead of an error.
func isOTPMismatch(msg string) bool {
	lower := strings.ToLower(msg)
	return strings.Contains(lower, "otp not match") ||
		strings.Contains(lower, "otp mismatch") ||
		strings.Contains(lower, "wrong otp") ||
		strings.Contains(lower, "incorrect otp") ||
		strings.Contains(lower, "otp expired") ||
		strings.Contains(lower, "otp has expired")
}

// RetryOTP requests MSG91 to resend an OTP to the phone number via the
// given channel ("text" or "voice"). Provider-specific; not on the
// [gosms.Provider] interface.
func (p *Provider) RetryOTP(ctx context.Context, phone, channel string) error {
	if phone == "" {
		return fmt.Errorf("%w: phone required", gosms.ErrInvalidConfig)
	}
	if channel == "" {
		channel = "text"
	}
	if channel != "text" && channel != "voice" {
		return fmt.Errorf("%w: channel must be text or voice", gosms.ErrInvalidConfig)
	}

	q := url.Values{}
	q.Set("mobile", p.normalizeRecipient(phone))
	q.Set("retrytype", channel)

	endpoint := p.config.BaseURL + otpRetryPath + "?" + q.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, nil)
	if err != nil {
		return err
	}
	req.Header.Set("authkey", p.config.AuthKey)

	resp, err := p.config.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %w", gosms.ErrSendFailed, err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return err
	}

	var otpResp otpResponse
	if err := json.Unmarshal(body, &otpResp); err != nil {
		if resp.StatusCode >= 400 {
			return fmt.Errorf("%w: http %d: %s", gosms.ErrProviderError, resp.StatusCode, bodySnippet(body))
		}
		return fmt.Errorf("%w: decode response: %w", gosms.ErrProviderError, err)
	}

	if resp.StatusCode >= 400 || otpResp.Type != "success" {
		return parseFlowError(flowResponse(otpResp))
	}
	return nil
}

// ParseWebhook parses an MSG91 delivery report webhook. MSG91 delivery
// callbacks are form-encoded POST requests with fields including
// `requestId`, `mobile`, `status`, `description`, and `statusCode`.
func ParseWebhook(r *http.Request) (*gosms.Status, error) {
	if err := r.ParseForm(); err != nil {
		return nil, err
	}

	requestID := r.FormValue("requestId")
	if requestID == "" {
		requestID = r.FormValue("request_id")
	}

	return &gosms.Status{
		MessageID:    requestID,
		Status:       mapStatus(r.FormValue("status")),
		UpdatedAt:    time.Now(),
		ErrorCode:    r.FormValue("statusCode"),
		ErrorMessage: r.FormValue("description"),
		Raw: map[string]any{
			"request_id":  requestID,
			"mobile":      r.FormValue("mobile"),
			"status":      r.FormValue("status"),
			"description": r.FormValue("description"),
			"status_code": r.FormValue("statusCode"),
		},
	}, nil
}

// SetVar sets a Flow template variable on the message. Use this for
// templates with placeholders like ##name## or ##otp##.
//
//	msg := gosms.NewMessage("+919876543210", "")
//	msg91.SetVar(msg, "name", "Kartik")
//	msg91.SetVar(msg, "otp", "1234")
func SetVar(msg *gosms.Message, key, value string) *gosms.Message {
	if msg.Metadata == nil {
		msg.Metadata = make(map[string]string)
	}
	msg.Metadata[metaVarPrefix+key] = value
	return msg
}

// SetTemplateID overrides the default Flow template ID for a single message.
// If unset, [Config.TemplateID] is used.
func SetTemplateID(msg *gosms.Message, templateID string) *gosms.Message {
	if msg.Metadata == nil {
		msg.Metadata = make(map[string]string)
	}
	msg.Metadata[metaTemplateID] = templateID
	return msg
}

// VerifyResult is returned from [Provider.VerifyOTP].
type VerifyResult struct {
	// Verified is true when MSG91 confirms the OTP matches.
	Verified bool

	// Message is the raw MSG91 message string (useful for failure reasons).
	Message string

	// Raw contains the raw provider response.
	Raw map[string]any
}

func (p *Provider) normalizeRecipient(phone string) string {
	phone = strings.TrimSpace(phone)
	phone = strings.TrimPrefix(phone, "+")

	digits := make([]byte, 0, len(phone))
	for i := 0; i < len(phone); i++ {
		c := phone[i]
		if c >= '0' && c <= '9' {
			digits = append(digits, c)
		}
	}
	phone = string(digits)

	if phone == "" {
		return phone
	}

	// If the number looks like a bare national number (<= 10 digits for
	// most markets), prepend the configured country code.
	if len(phone) <= 10 && p.config.Country != "" {
		phone = p.config.Country + phone
	}
	return phone
}

func extractVars(msg *gosms.Message) map[string]string {
	out := make(map[string]string)
	for k, v := range msg.Metadata {
		if name, ok := strings.CutPrefix(k, metaVarPrefix); ok {
			out[name] = v
		}
	}
	return out
}

func templateIDFor(msg *gosms.Message, fallback string) string {
	if tid, ok := msg.Metadata[metaTemplateID]; ok && tid != "" {
		return tid
	}
	return fallback
}

func failedResult(msg *gosms.Message, provider string, err error) *gosms.Result {
	return &gosms.Result{
		To:       msg.To,
		Status:   gosms.StatusFailed,
		Provider: provider,
		Error:    err.Error(),
		SentAt:   time.Now(),
	}
}

type flowRequest struct {
	TemplateID string           `json:"template_id"`
	SenderID   string           `json:"sender,omitempty"`
	Route      string           `json:"route,omitempty"`
	ShortURL   string           `json:"short_url,omitempty"`
	RealTime   string           `json:"realTimeResponse,omitempty"`
	Recipients []map[string]any `json:"recipients"`
}

type flowResponse struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

type otpResponse struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

func bodySnippet(b []byte) string {
	const max = 200
	s := strings.TrimSpace(string(b))
	if len(s) > max {
		s = s[:max] + "..."
	}
	return s
}

func mapStatus(status string) gosms.DeliveryStatus {
	switch strings.ToLower(status) {
	case "delivered", "dlvrd":
		return gosms.StatusDelivered
	case "sent", "submitted":
		return gosms.StatusSent
	case "queued", "pending":
		return gosms.StatusQueued
	case "failed", "undelivered", "undeliv":
		return gosms.StatusFailed
	case "rejected", "ndnc", "dnd":
		return gosms.StatusRejected
	case "expired":
		return gosms.StatusExpired
	default:
		return gosms.StatusUnknown
	}
}

func parseFlowError(resp flowResponse) error {
	msg := resp.Message
	lower := strings.ToLower(msg)

	switch {
	case strings.Contains(lower, "invalid") && strings.Contains(lower, "mobile"):
		return fmt.Errorf("%w: %s", gosms.ErrInvalidPhone, msg)
	case strings.Contains(lower, "authkey") || strings.Contains(lower, "auth key") || strings.Contains(lower, "unauthorized"):
		return fmt.Errorf("%w: %s", gosms.ErrInvalidConfig, msg)
	case strings.Contains(lower, "insufficient") || strings.Contains(lower, "balance"):
		return fmt.Errorf("%w: %s", gosms.ErrInsufficientFunds, msg)
	case strings.Contains(lower, "rate") && strings.Contains(lower, "limit"):
		return fmt.Errorf("%w: %s", gosms.ErrRateLimited, msg)
	case strings.Contains(lower, "dnd") || strings.Contains(lower, "ndnc") || strings.Contains(lower, "blacklist"):
		return fmt.Errorf("%w: %s", gosms.ErrBlacklisted, msg)
	case strings.Contains(lower, "template"):
		return fmt.Errorf("%w: %s", gosms.ErrInvalidConfig, msg)
	default:
		return fmt.Errorf("%w: %s", gosms.ErrProviderError, msg)
	}
}
