// Package gosms provides a unified interface for sending SMS messages
// via multiple providers including Twilio, AWS SNS, and Vonage.
package gosms

import (
	"context"
	"errors"
	"time"
)

// Sentinel errors for SMS operations.
var (
	ErrInvalidConfig     = errors.New("sms: invalid configuration")
	ErrInvalidPhone      = errors.New("sms: invalid phone number")
	ErrInvalidMessage    = errors.New("sms: invalid message")
	ErrSendFailed        = errors.New("sms: send failed")
	ErrProviderError     = errors.New("sms: provider error")
	ErrRateLimited       = errors.New("sms: rate limited")
	ErrInsufficientFunds = errors.New("sms: insufficient funds")
	ErrBlacklisted       = errors.New("sms: number blacklisted")
	ErrUnsupported       = errors.New("sms: operation not supported")
)

// Provider represents an SMS provider.
type Provider interface {
	// Send sends an SMS message.
	Send(ctx context.Context, msg *Message) (*Result, error)

	// SendBulk sends multiple SMS messages.
	SendBulk(ctx context.Context, msgs []*Message) ([]*Result, error)

	// GetStatus retrieves the delivery status of a message.
	GetStatus(ctx context.Context, messageID string) (*Status, error)

	// Name returns the provider name.
	Name() string
}

// Message represents an SMS message.
type Message struct {
	// To is the recipient phone number (E.164 format recommended).
	To string

	// From is the sender ID or phone number.
	From string

	// Body is the message content.
	Body string

	// Reference is an optional client reference.
	Reference string

	// ScheduledAt schedules the message for future delivery.
	ScheduledAt *time.Time

	// ValidityPeriod is how long the message is valid for delivery.
	ValidityPeriod time.Duration

	// Metadata holds additional provider-specific data.
	Metadata map[string]string
}

// NewMessage creates a new SMS message.
func NewMessage(to, body string) *Message {
	return &Message{
		To:       to,
		Body:     body,
		Metadata: make(map[string]string),
	}
}

// WithFrom sets the sender ID.
func (m *Message) WithFrom(from string) *Message {
	m.From = from
	return m
}

// WithReference sets the client reference.
func (m *Message) WithReference(ref string) *Message {
	m.Reference = ref
	return m
}

// WithSchedule schedules the message for later delivery.
func (m *Message) WithSchedule(t time.Time) *Message {
	m.ScheduledAt = &t
	return m
}

// WithValidity sets the validity period.
func (m *Message) WithValidity(d time.Duration) *Message {
	m.ValidityPeriod = d
	return m
}

// WithMetadata adds metadata.
func (m *Message) WithMetadata(key, value string) *Message {
	m.Metadata[key] = value
	return m
}

// Validate validates the message.
func (m *Message) Validate() error {
	if m.To == "" {
		return ErrInvalidPhone
	}
	if m.Body == "" {
		return ErrInvalidMessage
	}
	return nil
}

// Result represents the result of sending an SMS.
type Result struct {
	// MessageID is the provider's message identifier.
	MessageID string

	// To is the recipient phone number.
	To string

	// Status is the initial status.
	Status DeliveryStatus

	// Provider is the provider name.
	Provider string

	// Cost is the message cost (if available).
	Cost string

	// Currency is the cost currency.
	Currency string

	// Segments is the number of message segments.
	Segments int

	// SentAt is when the message was sent.
	SentAt time.Time

	// Error contains any error message.
	Error string

	// Raw contains the raw provider response.
	Raw map[string]any
}

// Success returns true if the message was accepted for delivery.
// This checks that the provider accepted the message (accepted, sent, or delivered),
// not that it was confirmed delivered. Use DeliveryStatus.IsSuccess for confirmed delivery.
func (r *Result) Success() bool {
	return r.Status == StatusAccepted || r.Status == StatusSent || r.Status == StatusDelivered
}

// Status represents the delivery status of a message.
type Status struct {
	// MessageID is the provider's message identifier.
	MessageID string

	// Status is the current delivery status.
	Status DeliveryStatus

	// UpdatedAt is when the status was last updated.
	UpdatedAt time.Time

	// ErrorCode is the error code (if failed).
	ErrorCode string

	// ErrorMessage is the error message (if failed).
	ErrorMessage string

	// Raw contains the raw provider response.
	Raw map[string]any
}

// DeliveryStatus represents the delivery status of a message.
type DeliveryStatus string

const (
	StatusPending   DeliveryStatus = "pending"
	StatusQueued    DeliveryStatus = "queued"
	StatusAccepted  DeliveryStatus = "accepted"
	StatusSent      DeliveryStatus = "sent"
	StatusDelivered DeliveryStatus = "delivered"
	StatusFailed    DeliveryStatus = "failed"
	StatusRejected  DeliveryStatus = "rejected"
	StatusExpired   DeliveryStatus = "expired"
	StatusUnknown   DeliveryStatus = "unknown"
)

// IsFinal returns true if the status is a final state.
func (s DeliveryStatus) IsFinal() bool {
	switch s {
	case StatusDelivered, StatusFailed, StatusRejected, StatusExpired:
		return true
	default:
		return false
	}
}

// IsSuccess returns true if the status indicates success.
func (s DeliveryStatus) IsSuccess() bool {
	return s == StatusDelivered
}

// SendEach sends each message individually using the provided send function,
// collecting results. Failed sends are recorded as failed results rather than
// returning an error.
func SendEach(ctx context.Context, providerName string, msgs []*Message, send func(context.Context, *Message) (*Result, error)) []*Result {
	results := make([]*Result, len(msgs))
	for i, msg := range msgs {
		result, err := send(ctx, msg)
		if err != nil {
			results[i] = &Result{
				To:       msg.To,
				Status:   StatusFailed,
				Provider: providerName,
				Error:    err.Error(),
				SentAt:   time.Now(),
			}
		} else {
			results[i] = result
		}
	}
	return results
}

// Client is the main SMS client.
type Client struct {
	provider    Provider
	defaultFrom string
}

// NewClient creates a new SMS client. Panics if provider is nil.
func NewClient(provider Provider) *Client {
	if provider == nil {
		panic("gosms: provider must not be nil")
	}
	return &Client{
		provider: provider,
	}
}

// WithDefaultFrom sets the default sender ID.
func (c *Client) WithDefaultFrom(from string) *Client {
	c.defaultFrom = from
	return c
}

// Send sends an SMS message.
func (c *Client) Send(ctx context.Context, to, body string) (*Result, error) {
	msg := NewMessage(to, body)
	if c.defaultFrom != "" {
		msg.From = c.defaultFrom
	}
	return c.SendMessage(ctx, msg)
}

// SendMessage sends an SMS message.
func (c *Client) SendMessage(ctx context.Context, msg *Message) (*Result, error) {
	if err := msg.Validate(); err != nil {
		return nil, err
	}

	if msg.From == "" && c.defaultFrom != "" {
		msg.From = c.defaultFrom
	}

	return c.provider.Send(ctx, msg)
}

// SendBulk sends multiple SMS messages. Messages that fail validation are
// recorded as failed results rather than aborting the entire batch.
func (c *Client) SendBulk(ctx context.Context, msgs []*Message) ([]*Result, error) {
	valid := make([]*Message, 0, len(msgs))
	results := make([]*Result, len(msgs))
	validIdx := make([]int, 0, len(msgs))

	for i, msg := range msgs {
		if err := msg.Validate(); err != nil {
			results[i] = &Result{
				To:       msg.To,
				Status:   StatusFailed,
				Provider: c.provider.Name(),
				Error:    err.Error(),
				SentAt:   time.Now(),
			}
			continue
		}
		if msg.From == "" && c.defaultFrom != "" {
			msg.From = c.defaultFrom
		}
		valid = append(valid, msg)
		validIdx = append(validIdx, i)
	}

	if len(valid) == 0 {
		return results, nil
	}

	sent, err := c.provider.SendBulk(ctx, valid)
	if err != nil {
		return nil, err
	}

	for i, idx := range validIdx {
		if i < len(sent) {
			results[idx] = sent[i]
		} else {
			results[idx] = &Result{
				To:       msgs[idx].To,
				Status:   StatusFailed,
				Provider: c.provider.Name(),
				Error:    "provider returned fewer results than messages sent",
				SentAt:   time.Now(),
			}
		}
	}

	return results, nil
}

// GetStatus retrieves the delivery status of a message.
func (c *Client) GetStatus(ctx context.Context, messageID string) (*Status, error) {
	return c.provider.GetStatus(ctx, messageID)
}

// Provider returns the underlying provider.
func (c *Client) Provider() Provider {
	return c.provider
}

// ProviderName returns the provider name.
func (c *Client) ProviderName() string {
	return c.provider.Name()
}
