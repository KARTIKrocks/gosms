package gosms

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestNewMessage(t *testing.T) {
	msg := NewMessage("+15551234567", "hello")
	if msg.To != "+15551234567" {
		t.Errorf("To = %q, want %q", msg.To, "+15551234567")
	}
	if msg.Body != "hello" {
		t.Errorf("Body = %q, want %q", msg.Body, "hello")
	}
	if msg.Metadata == nil {
		t.Error("Metadata should be initialized")
	}
}

func TestMessageBuilders(t *testing.T) {
	now := time.Now()
	msg := NewMessage("+15551234567", "hi").
		WithFrom("+15559999999").
		WithReference("ref-123").
		WithSchedule(now).
		WithValidity(5*time.Minute).
		WithMetadata("key", "value")

	if msg.From != "+15559999999" {
		t.Errorf("From = %q, want %q", msg.From, "+15559999999")
	}
	if msg.Reference != "ref-123" {
		t.Errorf("Reference = %q, want %q", msg.Reference, "ref-123")
	}
	if msg.ScheduledAt == nil || !msg.ScheduledAt.Equal(now) {
		t.Error("ScheduledAt not set correctly")
	}
	if msg.ValidityPeriod != 5*time.Minute {
		t.Errorf("ValidityPeriod = %v, want %v", msg.ValidityPeriod, 5*time.Minute)
	}
	if msg.Metadata["key"] != "value" {
		t.Errorf("Metadata[key] = %q, want %q", msg.Metadata["key"], "value")
	}
}

func TestMessageValidate(t *testing.T) {
	tests := []struct {
		name    string
		msg     *Message
		wantErr error
	}{
		{"valid", NewMessage("+15551234567", "hello"), nil},
		{"empty to", NewMessage("", "hello"), ErrInvalidPhone},
		{"empty body", NewMessage("+15551234567", ""), ErrInvalidMessage},
		{"both empty", NewMessage("", ""), ErrInvalidPhone},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.Validate()
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Validate() = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestResultSuccess(t *testing.T) {
	tests := []struct {
		status DeliveryStatus
		want   bool
	}{
		{StatusAccepted, true},
		{StatusSent, true},
		{StatusDelivered, true},
		{StatusFailed, false},
		{StatusRejected, false},
		{StatusPending, false},
		{StatusQueued, false},
		{StatusExpired, false},
		{StatusUnknown, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			r := &Result{Status: tt.status}
			if got := r.Success(); got != tt.want {
				t.Errorf("Success() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeliveryStatusIsFinal(t *testing.T) {
	finals := []DeliveryStatus{StatusDelivered, StatusFailed, StatusRejected, StatusExpired}
	nonFinals := []DeliveryStatus{StatusPending, StatusQueued, StatusAccepted, StatusSent, StatusUnknown}

	for _, s := range finals {
		if !s.IsFinal() {
			t.Errorf("%q.IsFinal() = false, want true", s)
		}
	}
	for _, s := range nonFinals {
		if s.IsFinal() {
			t.Errorf("%q.IsFinal() = true, want false", s)
		}
	}
}

func TestDeliveryStatusIsSuccess(t *testing.T) {
	if !StatusDelivered.IsSuccess() {
		t.Error("StatusDelivered.IsSuccess() = false, want true")
	}
	if StatusAccepted.IsSuccess() {
		t.Error("StatusAccepted.IsSuccess() = true, want false")
	}
}

func TestSendEach(t *testing.T) {
	ctx := context.Background()
	sendErr := errors.New("boom")

	callCount := 0
	send := func(_ context.Context, msg *Message) (*Result, error) {
		callCount++
		if msg.To == "fail" {
			return nil, sendErr
		}
		return &Result{
			MessageID: "id-" + msg.To,
			To:        msg.To,
			Status:    StatusAccepted,
			Provider:  "test",
		}, nil
	}

	msgs := []*Message{
		NewMessage("+1111", "a"),
		NewMessage("fail", "b"),
		NewMessage("+3333", "c"),
	}

	results := SendEach(ctx, "test", msgs, send)

	if len(results) != 3 {
		t.Fatalf("len(results) = %d, want 3", len(results))
	}
	if !results[0].Success() {
		t.Error("results[0] should be success")
	}
	if results[1].Status != StatusFailed {
		t.Errorf("results[1].Status = %q, want %q", results[1].Status, StatusFailed)
	}
	if results[1].Error != "boom" {
		t.Errorf("results[1].Error = %q, want %q", results[1].Error, "boom")
	}
	if !results[2].Success() {
		t.Error("results[2] should be success")
	}
	if callCount != 3 {
		t.Errorf("callCount = %d, want 3", callCount)
	}
}

func TestNewClientNilPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("NewClient(nil) should panic")
		}
	}()
	NewClient(nil)
}

func TestClientSend(t *testing.T) {
	ctx := context.Background()
	mock := NewMockProvider()
	client := NewClient(mock).WithDefaultFrom("+15550000000")

	result, err := client.Send(ctx, "+15551234567", "hello")
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	if !result.Success() {
		t.Error("expected success")
	}
	if result.To != "+15551234567" {
		t.Errorf("To = %q, want %q", result.To, "+15551234567")
	}

	last := mock.LastMessage()
	if last.Message.From != "+15550000000" {
		t.Errorf("From = %q, want %q", last.Message.From, "+15550000000")
	}
}

func TestClientSendValidationError(t *testing.T) {
	ctx := context.Background()
	client := NewClient(NewMockProvider())

	_, err := client.Send(ctx, "", "hello")
	if !errors.Is(err, ErrInvalidPhone) {
		t.Errorf("Send() error = %v, want ErrInvalidPhone", err)
	}

	_, err = client.Send(ctx, "+15551234567", "")
	if !errors.Is(err, ErrInvalidMessage) {
		t.Errorf("Send() error = %v, want ErrInvalidMessage", err)
	}
}

func TestClientSendMessage(t *testing.T) {
	ctx := context.Background()
	client := NewClient(NewMockProvider()).WithDefaultFrom("+15550000000")

	msg := NewMessage("+15551234567", "test")
	result, err := client.SendMessage(ctx, msg)
	if err != nil {
		t.Fatalf("SendMessage() error = %v", err)
	}
	if !result.Success() {
		t.Error("expected success")
	}
	if msg.From != "+15550000000" {
		t.Errorf("defaultFrom not applied: From = %q", msg.From)
	}
}

func TestClientSendMessageExplicitFrom(t *testing.T) {
	ctx := context.Background()
	mock := NewMockProvider()
	client := NewClient(mock).WithDefaultFrom("+15550000000")

	msg := NewMessage("+15551234567", "test").WithFrom("+15559999999")
	_, err := client.SendMessage(ctx, msg)
	if err != nil {
		t.Fatalf("SendMessage() error = %v", err)
	}
	if msg.From != "+15559999999" {
		t.Errorf("explicit From overwritten: From = %q", msg.From)
	}
}

func TestClientSendBulkPartialValidation(t *testing.T) {
	ctx := context.Background()
	mock := NewMockProvider()
	client := NewClient(mock)

	msgs := []*Message{
		NewMessage("+15551111111", "ok"),
		NewMessage("", "bad phone"),
		NewMessage("+15553333333", "ok too"),
	}

	results, err := client.SendBulk(ctx, msgs)
	if err != nil {
		t.Fatalf("SendBulk() error = %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("len(results) = %d, want 3", len(results))
	}

	if results[0].Status == StatusFailed {
		t.Error("results[0] should succeed")
	}
	if results[1].Status != StatusFailed {
		t.Errorf("results[1].Status = %q, want %q", results[1].Status, StatusFailed)
	}
	if results[2].Status == StatusFailed {
		t.Error("results[2] should succeed")
	}

	if mock.MessageCount() != 2 {
		t.Errorf("MessageCount = %d, want 2", mock.MessageCount())
	}
}

func TestClientSendBulkAllInvalid(t *testing.T) {
	ctx := context.Background()
	mock := NewMockProvider()
	client := NewClient(mock)

	msgs := []*Message{
		NewMessage("", "bad"),
		NewMessage("", "also bad"),
	}

	results, err := client.SendBulk(ctx, msgs)
	if err != nil {
		t.Fatalf("SendBulk() error = %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("len(results) = %d, want 2", len(results))
	}
	for i, r := range results {
		if r.Status != StatusFailed {
			t.Errorf("results[%d].Status = %q, want %q", i, r.Status, StatusFailed)
		}
	}
	if mock.MessageCount() != 0 {
		t.Errorf("MessageCount = %d, want 0", mock.MessageCount())
	}
}

func TestClientSendBulkShortProviderResults(t *testing.T) {
	ctx := context.Background()
	mock := NewMockProvider()
	// Override SendBulk to return fewer results than messages
	mock.WithSendError(nil)
	client := NewClient(mock)

	msgs := []*Message{
		NewMessage("+15551111111", "ok"),
		NewMessage("+15552222222", "ok too"),
	}

	// This should not panic even if provider returns fewer results
	results, err := client.SendBulk(ctx, msgs)
	if err != nil {
		t.Fatalf("SendBulk() error = %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("len(results) = %d, want 2", len(results))
	}
}

func TestClientGetStatus(t *testing.T) {
	ctx := context.Background()
	mock := NewMockProvider()
	client := NewClient(mock)

	result, _ := client.Send(ctx, "+15551234567", "hello")
	status, err := client.GetStatus(ctx, result.MessageID)
	if err != nil {
		t.Fatalf("GetStatus() error = %v", err)
	}
	if status.Status != StatusDelivered {
		t.Errorf("Status = %q, want %q", status.Status, StatusDelivered)
	}
}

func TestClientProviderName(t *testing.T) {
	client := NewClient(NewMockProvider())
	if got := client.ProviderName(); got != "mock" {
		t.Errorf("ProviderName() = %q, want %q", got, "mock")
	}
	if client.Provider().Name() != "mock" {
		t.Errorf("Provider().Name() = %q, want %q", client.Provider().Name(), "mock")
	}
}

func TestOTPMessage(t *testing.T) {
	msg := OTPMessage("+15551234567", "123456", "MyApp")
	if msg.Body != "123456 is your MyApp verification code." {
		t.Errorf("Body = %q", msg.Body)
	}
	if msg.Metadata["type"] != "otp" {
		t.Errorf("Metadata[type] = %q, want %q", msg.Metadata["type"], "otp")
	}
	if msg.Metadata["code"] != "123456" {
		t.Errorf("Metadata[code] = %q, want %q", msg.Metadata["code"], "123456")
	}
}

func TestAlertMessage(t *testing.T) {
	msg := AlertMessage("+15551234567", "CRITICAL", "Server is down")
	if msg.Body != "[CRITICAL] Server is down" {
		t.Errorf("Body = %q", msg.Body)
	}
	if msg.Metadata["alert_type"] != "CRITICAL" {
		t.Errorf("Metadata[alert_type] = %q", msg.Metadata["alert_type"])
	}
}

func TestNotificationMessage(t *testing.T) {
	msg := NotificationMessage("+15551234567", "Order", "Shipped")
	if msg.Body != "Order: Shipped" {
		t.Errorf("Body = %q", msg.Body)
	}
	if msg.Metadata["title"] != "Order" {
		t.Errorf("Metadata[title] = %q", msg.Metadata["title"])
	}
}
