package gosms

import (
	"context"
	"regexp"
	"sync/atomic"
	"unicode/utf8"
)

// e164Regex is the regular expression for E.164 phone number format.
var e164Regex = regexp.MustCompile(`^\+[1-9]\d{1,14}$`)

// nonDigitRegex is precompiled to avoid recompilation on every NormalizePhone call.
var nonDigitRegex = regexp.MustCompile(`[^\d]`)

// ValidateE164 checks if a phone number is in E.164 format.
func ValidateE164(phone string) bool {
	return e164Regex.MatchString(phone)
}

// NormalizePhone attempts to normalize a phone number to E.164 format.
// This is a basic normalization and may not work for all countries.
func NormalizePhone(phone, defaultCountryCode string) string {
	hasPlus := len(phone) > 0 && phone[0] == '+'
	phone = nonDigitRegex.ReplaceAllString(phone, "")

	if hasPlus {
		phone = "+" + phone
	}

	if phone == "" {
		return phone
	}

	if phone[0] != '+' {
		if defaultCountryCode != "" {
			if defaultCountryCode[0] != '+' {
				defaultCountryCode = "+" + defaultCountryCode
			}
			phone = defaultCountryCode + phone
		}
	}

	return phone
}

// GSM 03.38 Basic Character Set.
var gsmBasicChars = func() map[rune]bool {
	m := make(map[rune]bool)
	for _, r := range "@£$¥èéùìòÇ\nØø\rÅåΔ_ΦΓΛΩΠΨΣΘΞ\x1BÆæßÉ !\"#¤%&'()*+,-./0123456789:;<=>?¡ABCDEFGHIJKLMNOPQRSTUVWXYZÄÖÑÜ§¿abcdefghijklmnopqrstuvwxyzäöñüà" {
		m[r] = true
	}
	return m
}()

// GSM 03.38 Extended Character Set (each counts as 2 septets).
var gsmExtendedChars = func() map[rune]bool {
	m := make(map[rune]bool)
	for _, r := range "{}[]~\\^|€\f" {
		m[r] = true
	}
	return m
}()

// IsGSMEncoding reports whether message can be encoded using GSM 7-bit encoding.
func IsGSMEncoding(message string) bool {
	for _, r := range message {
		if !gsmBasicChars[r] && !gsmExtendedChars[r] {
			return false
		}
	}
	return true
}

// GSMLen returns the length of a message in GSM 7-bit septets.
// Extended characters count as 2 septets each.
func GSMLen(message string) int {
	n := 0
	for _, r := range message {
		if gsmExtendedChars[r] {
			n += 2
		} else {
			n++
		}
	}
	return n
}

// utf16Len returns the number of UTF-16 code units needed to encode the string.
// Characters outside the Basic Multilingual Plane (e.g. emoji) require 2 code units.
func utf16Len(s string) int {
	n := 0
	for _, r := range s {
		if r > 0xFFFF {
			n += 2
		} else {
			n++
		}
	}
	return n
}

// CalculateSegments calculates the number of SMS segments for a message.
//
// GSM 7-bit encoding: 160 septets single / 153 septets per concatenated segment.
// Unicode (UCS-2): 70 characters single / 67 characters per concatenated segment.
// Extended GSM characters ({, }, [, ], ~, \, ^, |, €) count as 2 septets.
func CalculateSegments(message string) int {
	if utf8.RuneCountInString(message) == 0 {
		return 0
	}

	if IsGSMEncoding(message) {
		length := GSMLen(message)
		if length <= 160 {
			return 1
		}
		return (length + 152) / 153
	}

	length := utf16Len(message)
	if length <= 70 {
		return 1
	}
	return (length + 66) / 67
}

// Batch represents a batch of messages to send.
type Batch struct {
	messages []*Message
}

// NewBatch creates a new message batch.
func NewBatch() *Batch {
	return &Batch{
		messages: make([]*Message, 0),
	}
}

// Add adds a message to the batch.
func (b *Batch) Add(msg *Message) *Batch {
	b.messages = append(b.messages, msg)
	return b
}

// AddNew adds a new message to the batch.
func (b *Batch) AddNew(to, body string) *Batch {
	return b.Add(NewMessage(to, body))
}

// AddNewWithFrom adds a new message with sender to the batch.
func (b *Batch) AddNewWithFrom(to, body, from string) *Batch {
	return b.Add(NewMessage(to, body).WithFrom(from))
}

// Messages returns all messages in the batch.
func (b *Batch) Messages() []*Message {
	return b.messages
}

// Size returns the number of messages in the batch.
func (b *Batch) Size() int {
	return len(b.messages)
}

// Clear clears all messages from the batch.
func (b *Batch) Clear() {
	b.messages = nil
}

// Send sends all messages in the batch using the provided client.
func (b *Batch) Send(ctx context.Context, client *Client) ([]*Result, error) {
	return client.SendBulk(ctx, b.messages)
}

// SendToMany sends the same message to multiple recipients.
func SendToMany(ctx context.Context, client *Client, body string, recipients ...string) ([]*Result, error) {
	msgs := make([]*Message, len(recipients))
	for i, to := range recipients {
		msgs[i] = NewMessage(to, body)
	}
	return client.SendBulk(ctx, msgs)
}

// QuickSend is a convenience function to quickly send an SMS.
func QuickSend(ctx context.Context, provider Provider, to, from, body string) (*Result, error) {
	client := NewClient(provider).WithDefaultFrom(from)
	return client.Send(ctx, to, body)
}

// MultiProvider allows sending via multiple providers with fallback or round-robin.
type MultiProvider struct {
	providers []Provider
	strategy  MultiProviderStrategy
	counter   atomic.Uint64
}

// MultiProviderStrategy determines how providers are selected.
type MultiProviderStrategy int

const (
	// StrategyFallback tries providers in order until one succeeds.
	StrategyFallback MultiProviderStrategy = iota
	// StrategyRoundRobin rotates through providers.
	StrategyRoundRobin
)

// NewMultiProvider creates a new multi-provider.
func NewMultiProvider(providers ...Provider) *MultiProvider {
	return &MultiProvider{
		providers: providers,
		strategy:  StrategyFallback,
	}
}

// WithStrategy sets the provider selection strategy.
func (p *MultiProvider) WithStrategy(strategy MultiProviderStrategy) *MultiProvider {
	p.strategy = strategy
	return p
}

// Name returns the provider name.
func (p *MultiProvider) Name() string {
	return "multi"
}

// Send sends a message using the configured strategy.
func (p *MultiProvider) Send(ctx context.Context, msg *Message) (*Result, error) {
	if len(p.providers) == 0 {
		return nil, ErrInvalidConfig
	}

	switch p.strategy {
	case StrategyRoundRobin:
		idx := p.counter.Add(1) - 1
		provider := p.providers[idx%uint64(len(p.providers))]
		return provider.Send(ctx, msg)
	default:
		var lastErr error
		for _, provider := range p.providers {
			result, err := provider.Send(ctx, msg)
			if err == nil {
				return result, nil
			}
			lastErr = err
		}
		return nil, lastErr
	}
}

// SendBulk sends multiple messages.
func (p *MultiProvider) SendBulk(ctx context.Context, msgs []*Message) ([]*Result, error) {
	return SendEach(ctx, p.Name(), msgs, p.Send), nil
}

// GetStatus retrieves status from the first provider that doesn't return an error.
func (p *MultiProvider) GetStatus(ctx context.Context, messageID string) (*Status, error) {
	for _, provider := range p.providers {
		status, err := provider.GetStatus(ctx, messageID)
		if err == nil {
			return status, nil
		}
	}
	return nil, ErrUnsupported
}

// OTPMessage creates a message formatted for OTP delivery.
func OTPMessage(to, code, appName string) *Message {
	body := code + " is your " + appName + " verification code."
	return NewMessage(to, body).
		WithMetadata("type", "otp").
		WithMetadata("code", code)
}

// AlertMessage creates a message formatted for alerts.
func AlertMessage(to, alertType, message string) *Message {
	body := "[" + alertType + "] " + message
	return NewMessage(to, body).
		WithMetadata("type", "alert").
		WithMetadata("alert_type", alertType)
}

// NotificationMessage creates a message formatted for notifications.
func NotificationMessage(to, title, message string) *Message {
	body := title + ": " + message
	return NewMessage(to, body).
		WithMetadata("type", "notification").
		WithMetadata("title", title)
}
