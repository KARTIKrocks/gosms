package gosms

import (
	"context"
	"crypto/rand"
	"fmt"
	"sync"
	"time"
)

func generateID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

// MockProvider is a mock SMS provider for testing.
type MockProvider struct {
	mu          sync.RWMutex
	messages    []*MockMessage
	statuses    map[string]*Status
	sendError   error
	statusError error
	deliverAll  bool
	failAll     bool
	latency     time.Duration
}

// MockMessage represents a mock sent message.
type MockMessage struct {
	MessageID string
	Message   *Message
	SentAt    time.Time
	Status    DeliveryStatus
}

// NewMockProvider creates a new mock provider.
func NewMockProvider() *MockProvider {
	return &MockProvider{
		messages:   make([]*MockMessage, 0),
		statuses:   make(map[string]*Status),
		deliverAll: true,
	}
}

// Name returns the provider name.
func (p *MockProvider) Name() string {
	return "mock"
}

// Send sends a mock SMS message.
func (p *MockProvider) Send(ctx context.Context, msg *Message) (*Result, error) {
	p.mu.RLock()
	latency := p.latency
	sendErr := p.sendError
	p.mu.RUnlock()

	if latency > 0 {
		select {
		case <-time.After(latency):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	if sendErr != nil {
		return nil, sendErr
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	messageID := generateID()

	status := StatusAccepted
	if p.failAll {
		status = StatusFailed
	}

	mockMsg := &MockMessage{
		MessageID: messageID,
		Message:   msg,
		SentAt:    time.Now(),
		Status:    status,
	}
	p.messages = append(p.messages, mockMsg)

	finalStatus := status
	if p.deliverAll && !p.failAll {
		finalStatus = StatusDelivered
	}

	p.statuses[messageID] = &Status{
		MessageID: messageID,
		Status:    finalStatus,
		UpdatedAt: time.Now(),
	}

	return &Result{
		MessageID: messageID,
		To:        msg.To,
		Status:    status,
		Provider:  p.Name(),
		Cost:      "0.01",
		Currency:  "USD",
		Segments:  1,
		SentAt:    time.Now(),
		Raw:       map[string]any{"mock": true},
	}, nil
}

// SendBulk sends multiple mock SMS messages.
func (p *MockProvider) SendBulk(ctx context.Context, msgs []*Message) ([]*Result, error) {
	return SendEach(ctx, p.Name(), msgs, p.Send), nil
}

// GetStatus retrieves the status of a mock message.
func (p *MockProvider) GetStatus(ctx context.Context, messageID string) (*Status, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.statusError != nil {
		return nil, p.statusError
	}

	status, ok := p.statuses[messageID]
	if !ok {
		return &Status{
			MessageID: messageID,
			Status:    StatusUnknown,
			UpdatedAt: time.Now(),
		}, nil
	}

	return status, nil
}

// WithSendError sets the error to return on Send.
func (p *MockProvider) WithSendError(err error) *MockProvider {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.sendError = err
	return p
}

// WithStatusError sets the error to return on GetStatus.
func (p *MockProvider) WithStatusError(err error) *MockProvider {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.statusError = err
	return p
}

// WithDeliverAll sets whether all messages should be marked as delivered.
func (p *MockProvider) WithDeliverAll(deliver bool) *MockProvider {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.deliverAll = deliver
	return p
}

// WithFailAll sets whether all messages should fail.
func (p *MockProvider) WithFailAll(fail bool) *MockProvider {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.failAll = fail
	return p
}

// WithLatency sets simulated latency.
func (p *MockProvider) WithLatency(d time.Duration) *MockProvider {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.latency = d
	return p
}

// SetStatus manually sets the status for a message ID.
func (p *MockProvider) SetStatus(messageID string, status *Status) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.statuses[messageID] = status
}

// Messages returns all sent messages.
func (p *MockProvider) Messages() []*MockMessage {
	p.mu.RLock()
	defer p.mu.RUnlock()

	msgs := make([]*MockMessage, len(p.messages))
	copy(msgs, p.messages)
	return msgs
}

// LastMessage returns the last sent message.
func (p *MockProvider) LastMessage() *MockMessage {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if len(p.messages) == 0 {
		return nil
	}
	return p.messages[len(p.messages)-1]
}

// MessageCount returns the number of messages sent.
func (p *MockProvider) MessageCount() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.messages)
}

// Clear clears all sent messages.
func (p *MockProvider) Clear() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.messages = p.messages[:0]
	p.statuses = make(map[string]*Status)
}

// Reset resets the provider to its initial state.
func (p *MockProvider) Reset() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.messages = make([]*MockMessage, 0)
	p.statuses = make(map[string]*Status)
	p.sendError = nil
	p.statusError = nil
	p.deliverAll = true
	p.failAll = false
	p.latency = 0
}

// FindMessagesByTo finds all messages sent to a phone number.
func (p *MockProvider) FindMessagesByTo(to string) []*MockMessage {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var found []*MockMessage
	for _, msg := range p.messages {
		if msg.Message.To == to {
			found = append(found, msg)
		}
	}
	return found
}

// FindMessageByID finds a message by its ID.
func (p *MockProvider) FindMessageByID(id string) *MockMessage {
	p.mu.RLock()
	defer p.mu.RUnlock()

	for _, msg := range p.messages {
		if msg.MessageID == id {
			return msg
		}
	}
	return nil
}
