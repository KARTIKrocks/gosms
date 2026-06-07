// Example: using MockProvider in tests.
//
// Run: go test -v .
package main

import (
	"context"
	"errors"
	"testing"

	"github.com/KARTIKrocks/gosms"
)

func TestSendSMS(t *testing.T) {
	mock := gosms.NewMockProvider()
	client := gosms.NewClient(mock).WithDefaultFrom("+15551234567")
	ctx := context.Background()

	result, err := client.Send(ctx, "+15559876543", "Hello!")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Success() {
		t.Fatalf("expected success, got status %s", result.Status)
	}

	// Verify the mock recorded the message.
	if mock.MessageCount() != 1 {
		t.Fatalf("expected 1 message, got %d", mock.MessageCount())
	}

	last := mock.LastMessage()
	if last.Message.To != "+15559876543" {
		t.Errorf("expected to +15559876543, got %s", last.Message.To)
	}
	if last.Message.Body != "Hello!" {
		t.Errorf("expected body 'Hello!', got %s", last.Message.Body)
	}
}

func TestSendFailure(t *testing.T) {
	mock := gosms.NewMockProvider()
	mock.WithSendError(gosms.ErrRateLimited)
	client := gosms.NewClient(mock)
	ctx := context.Background()

	_, err := client.Send(ctx, "+15559876543", "This will fail")
	if !errors.Is(err, gosms.ErrRateLimited) {
		t.Fatalf("expected ErrRateLimited, got %v", err)
	}
}

func TestDeliveryStatus(t *testing.T) {
	mock := gosms.NewMockProvider()
	client := gosms.NewClient(mock).WithDefaultFrom("+15551234567")
	ctx := context.Background()

	result, _ := client.Send(ctx, "+15559876543", "Check status")

	status, err := client.GetStatus(ctx, result.MessageID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// MockProvider with deliverAll=true (default) marks as delivered.
	if status.Status != gosms.StatusDelivered {
		t.Errorf("expected delivered, got %s", status.Status)
	}
}

func TestBulkSend(t *testing.T) {
	mock := gosms.NewMockProvider()
	client := gosms.NewClient(mock).WithDefaultFrom("+15551234567")
	ctx := context.Background()

	results, err := gosms.SendToMany(ctx, client,
		"Bulk test",
		"+15551111111", "+15552222222", "+15553333333",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	for _, r := range results {
		if !r.Success() {
			t.Errorf("expected success for %s, got %s", r.To, r.Status)
		}
	}

	if mock.MessageCount() != 3 {
		t.Errorf("expected 3 messages, got %d", mock.MessageCount())
	}
}
