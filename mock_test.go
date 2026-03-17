package gosms

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestMockProviderSend(t *testing.T) {
	ctx := context.Background()
	mock := NewMockProvider()

	result, err := mock.Send(ctx, NewMessage("+15551234567", "hello"))
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}

	if result.MessageID == "" {
		t.Error("MessageID should not be empty")
	}
	if result.To != "+15551234567" {
		t.Errorf("To = %q", result.To)
	}
	if result.Status != StatusAccepted {
		t.Errorf("Status = %q, want %q", result.Status, StatusAccepted)
	}
	if result.Provider != "mock" {
		t.Errorf("Provider = %q", result.Provider)
	}
	if result.Cost != "0.01" {
		t.Errorf("Cost = %q", result.Cost)
	}
	if result.Segments != 1 {
		t.Errorf("Segments = %d", result.Segments)
	}
}

func TestMockProviderName(t *testing.T) {
	if got := NewMockProvider().Name(); got != "mock" {
		t.Errorf("Name() = %q, want %q", got, "mock")
	}
}

func TestMockProviderSendError(t *testing.T) {
	ctx := context.Background()
	mock := NewMockProvider().WithSendError(errors.New("network error"))

	_, err := mock.Send(ctx, NewMessage("+15551234567", "hello"))
	if err == nil || err.Error() != "network error" {
		t.Errorf("error = %v, want 'network error'", err)
	}
	if mock.MessageCount() != 0 {
		t.Errorf("MessageCount = %d, want 0", mock.MessageCount())
	}
}

func TestMockProviderFailAll(t *testing.T) {
	ctx := context.Background()
	mock := NewMockProvider().WithFailAll(true)

	result, err := mock.Send(ctx, NewMessage("+15551234567", "hello"))
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	if result.Status != StatusFailed {
		t.Errorf("Status = %q, want %q", result.Status, StatusFailed)
	}
}

func TestMockProviderDeliverAll(t *testing.T) {
	ctx := context.Background()
	mock := NewMockProvider().WithDeliverAll(true)

	result, _ := mock.Send(ctx, NewMessage("+15551234567", "hello"))
	status, err := mock.GetStatus(ctx, result.MessageID)
	if err != nil {
		t.Fatalf("GetStatus() error = %v", err)
	}
	if status.Status != StatusDelivered {
		t.Errorf("Status = %q, want %q", status.Status, StatusDelivered)
	}
}

func TestMockProviderNoDeliverAll(t *testing.T) {
	ctx := context.Background()
	mock := NewMockProvider().WithDeliverAll(false)

	result, _ := mock.Send(ctx, NewMessage("+15551234567", "hello"))
	status, _ := mock.GetStatus(ctx, result.MessageID)
	if status.Status != StatusAccepted {
		t.Errorf("Status = %q, want %q", status.Status, StatusAccepted)
	}
}

func TestMockProviderGetStatusUnknown(t *testing.T) {
	ctx := context.Background()
	mock := NewMockProvider()

	status, err := mock.GetStatus(ctx, "nonexistent")
	if err != nil {
		t.Fatalf("GetStatus() error = %v", err)
	}
	if status.Status != StatusUnknown {
		t.Errorf("Status = %q, want %q", status.Status, StatusUnknown)
	}
}

func TestMockProviderStatusError(t *testing.T) {
	ctx := context.Background()
	mock := NewMockProvider().WithStatusError(errors.New("status error"))

	_, err := mock.GetStatus(ctx, "any")
	if err == nil || err.Error() != "status error" {
		t.Errorf("error = %v, want 'status error'", err)
	}
}

func TestMockProviderSendBulk(t *testing.T) {
	ctx := context.Background()
	mock := NewMockProvider()

	msgs := []*Message{
		NewMessage("+15551111111", "a"),
		NewMessage("+15552222222", "b"),
	}

	results, err := mock.SendBulk(ctx, msgs)
	if err != nil {
		t.Fatalf("SendBulk() error = %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("len(results) = %d, want 2", len(results))
	}
	if mock.MessageCount() != 2 {
		t.Errorf("MessageCount = %d, want 2", mock.MessageCount())
	}
}

func TestMockProviderMessages(t *testing.T) {
	ctx := context.Background()
	mock := NewMockProvider()

	mock.Send(ctx, NewMessage("+15551111111", "a"))
	mock.Send(ctx, NewMessage("+15552222222", "b"))

	msgs := mock.Messages()
	if len(msgs) != 2 {
		t.Fatalf("len(Messages()) = %d, want 2", len(msgs))
	}

	// Verify it's a copy
	msgs[0] = nil
	if mock.Messages()[0] == nil {
		t.Error("Messages() should return a copy")
	}
}

func TestMockProviderLastMessage(t *testing.T) {
	ctx := context.Background()
	mock := NewMockProvider()

	if mock.LastMessage() != nil {
		t.Error("LastMessage() should be nil when empty")
	}

	mock.Send(ctx, NewMessage("+15551111111", "first"))
	mock.Send(ctx, NewMessage("+15552222222", "second"))

	last := mock.LastMessage()
	if last.Message.Body != "second" {
		t.Errorf("LastMessage().Body = %q, want %q", last.Message.Body, "second")
	}
}

func TestMockProviderClear(t *testing.T) {
	ctx := context.Background()
	mock := NewMockProvider()

	mock.Send(ctx, NewMessage("+15551234567", "hello"))
	mock.Clear()

	if mock.MessageCount() != 0 {
		t.Errorf("MessageCount after Clear = %d", mock.MessageCount())
	}
}

func TestMockProviderReset(t *testing.T) {
	ctx := context.Background()
	mock := NewMockProvider()
	mock.WithSendError(errors.New("err"))
	mock.WithStatusError(errors.New("err"))
	mock.WithFailAll(true)
	mock.WithDeliverAll(false)
	mock.WithLatency(time.Second)

	mock.Reset()

	result, err := mock.Send(ctx, NewMessage("+15551234567", "hello"))
	if err != nil {
		t.Fatalf("Send() after Reset error = %v", err)
	}
	if result.Status != StatusAccepted {
		t.Errorf("Status = %q, want %q", result.Status, StatusAccepted)
	}

	status, err := mock.GetStatus(ctx, result.MessageID)
	if err != nil {
		t.Fatalf("GetStatus() after Reset error = %v", err)
	}
	if status.Status != StatusDelivered {
		t.Errorf("Status = %q, want %q", status.Status, StatusDelivered)
	}
}

func TestMockProviderSetStatus(t *testing.T) {
	ctx := context.Background()
	mock := NewMockProvider()

	mock.SetStatus("custom-id", &Status{
		MessageID: "custom-id",
		Status:    StatusFailed,
	})

	status, err := mock.GetStatus(ctx, "custom-id")
	if err != nil {
		t.Fatalf("GetStatus() error = %v", err)
	}
	if status.Status != StatusFailed {
		t.Errorf("Status = %q, want %q", status.Status, StatusFailed)
	}
}

func TestMockProviderFindMessagesByTo(t *testing.T) {
	ctx := context.Background()
	mock := NewMockProvider()

	mock.Send(ctx, NewMessage("+15551111111", "a"))
	mock.Send(ctx, NewMessage("+15552222222", "b"))
	mock.Send(ctx, NewMessage("+15551111111", "c"))

	found := mock.FindMessagesByTo("+15551111111")
	if len(found) != 2 {
		t.Errorf("found %d messages, want 2", len(found))
	}

	found = mock.FindMessagesByTo("+15559999999")
	if len(found) != 0 {
		t.Errorf("found %d messages, want 0", len(found))
	}
}

func TestMockProviderFindMessageByID(t *testing.T) {
	ctx := context.Background()
	mock := NewMockProvider()

	result, _ := mock.Send(ctx, NewMessage("+15551234567", "hello"))

	found := mock.FindMessageByID(result.MessageID)
	if found == nil {
		t.Fatal("FindMessageByID returned nil")
	}
	if found.Message.Body != "hello" {
		t.Errorf("Body = %q, want %q", found.Message.Body, "hello")
	}

	if mock.FindMessageByID("nonexistent") != nil {
		t.Error("FindMessageByID should return nil for unknown ID")
	}
}

func TestMockProviderLatency(t *testing.T) {
	ctx := context.Background()
	mock := NewMockProvider().WithLatency(10 * time.Millisecond)

	start := time.Now()
	_, err := mock.Send(ctx, NewMessage("+15551234567", "hello"))
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	if elapsed < 10*time.Millisecond {
		t.Errorf("elapsed = %v, expected >= 10ms", elapsed)
	}
}

func TestMockProviderLatencyContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	mock := NewMockProvider().WithLatency(5 * time.Second)

	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	_, err := mock.Send(ctx, NewMessage("+15551234567", "hello"))
	elapsed := time.Since(start)

	if err != context.Canceled {
		t.Errorf("error = %v, want context.Canceled", err)
	}
	if elapsed >= 1*time.Second {
		t.Errorf("elapsed = %v, should have cancelled quickly", elapsed)
	}
}

func TestGenerateID(t *testing.T) {
	ids := make(map[string]bool)
	for range 100 {
		id := generateID()
		if ids[id] {
			t.Fatalf("duplicate ID generated: %s", id)
		}
		ids[id] = true
	}
}
