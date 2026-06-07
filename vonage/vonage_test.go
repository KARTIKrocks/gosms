package vonage

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	gosms "github.com/KARTIKrocks/gosms"
)

func newTestServer(handler http.HandlerFunc) (*httptest.Server, *Provider) {
	srv := httptest.NewServer(handler)
	p, _ := NewProvider(Config{
		APIKey:    "test_key",
		APISecret: "test_secret",
		From:      "+15550000000",
		BaseURL:   srv.URL,
	})
	return srv, p
}

func TestNewProviderValidation(t *testing.T) {
	_, err := NewProvider(Config{})
	if !errors.Is(err, gosms.ErrInvalidConfig) {
		t.Errorf("error = %v, want ErrInvalidConfig", err)
	}

	_, err = NewProvider(Config{APIKey: "key"})
	if !errors.Is(err, gosms.ErrInvalidConfig) {
		t.Errorf("error = %v, want ErrInvalidConfig", err)
	}

	p, err := NewProvider(Config{APIKey: "key", APISecret: "secret"})
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if p.Name() != "vonage" {
		t.Errorf("Name() = %q", p.Name())
	}
}

func TestNewProviderDefaults(t *testing.T) {
	p, _ := NewProvider(Config{APIKey: "key", APISecret: "secret"})
	if p.config.Type != TypeText {
		t.Errorf("Type = %q, want %q", p.config.Type, TypeText)
	}
	if p.config.HTTPClient == nil {
		t.Error("HTTPClient should be set")
	}
	if p.config.BaseURL != defaultBaseURL {
		t.Errorf("BaseURL = %q, want %q", p.config.BaseURL, defaultBaseURL)
	}
}

func TestSendSuccess(t *testing.T) {
	srv, p := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("method = %q, want POST", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("Content-Type = %q", ct)
		}

		var req sendRequest
		json.NewDecoder(r.Body).Decode(&req)
		if req.APIKey != "test_key" {
			t.Errorf("api_key = %q", req.APIKey)
		}
		if req.To != "+15551234567" {
			t.Errorf("To = %q", req.To)
		}
		if req.Text != "hello" {
			t.Errorf("Text = %q", req.Text)
		}

		resp := sendResponse{
			MessageCount: 1,
			Messages: []struct {
				To               string `json:"to"`
				MessageID        string `json:"message-id"`
				Status           string `json:"status"`
				RemainingBalance string `json:"remaining-balance"`
				MessagePrice     string `json:"message-price"`
				Currency         string `json:"currency,omitempty"`
				Network          string `json:"network"`
				ErrorText        string `json:"error-text,omitempty"`
			}{
				{
					To:               "+15551234567",
					MessageID:        "msg-001",
					Status:           "0",
					RemainingBalance: "10.00",
					MessagePrice:     "0.05",
					Currency:         "EUR",
					Network:          "12345",
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	})
	defer srv.Close()

	result, err := p.Send(context.Background(), gosms.NewMessage("+15551234567", "hello"))
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	if result.MessageID != "msg-001" {
		t.Errorf("MessageID = %q", result.MessageID)
	}
	if result.Status != gosms.StatusAccepted {
		t.Errorf("Status = %q", result.Status)
	}
	if result.Cost != "0.05" {
		t.Errorf("Cost = %q", result.Cost)
	}
	if result.Currency != "EUR" {
		t.Errorf("Currency = %q", result.Currency)
	}
}

func TestSendNoFrom(t *testing.T) {
	p, _ := NewProvider(Config{
		APIKey:    "key",
		APISecret: "secret",
		BaseURL:   "http://localhost",
	})

	_, err := p.Send(context.Background(), gosms.NewMessage("+15551234567", "hello"))
	if !errors.Is(err, gosms.ErrInvalidConfig) {
		t.Errorf("error = %v, want ErrInvalidConfig", err)
	}
}

func TestSendAPIError(t *testing.T) {
	srv, p := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		resp := sendResponse{
			MessageCount: 1,
			Messages: []struct {
				To               string `json:"to"`
				MessageID        string `json:"message-id"`
				Status           string `json:"status"`
				RemainingBalance string `json:"remaining-balance"`
				MessagePrice     string `json:"message-price"`
				Currency         string `json:"currency,omitempty"`
				Network          string `json:"network"`
				ErrorText        string `json:"error-text,omitempty"`
			}{
				{Status: "1", ErrorText: "Throttled"},
			},
		}
		json.NewEncoder(w).Encode(resp)
	})
	defer srv.Close()

	_, err := p.Send(context.Background(), gosms.NewMessage("+15551234567", "hello"))
	if !errors.Is(err, gosms.ErrRateLimited) {
		t.Errorf("error = %v, want ErrRateLimited", err)
	}
}

func TestSendEmptyResponse(t *testing.T) {
	srv, p := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		resp := sendResponse{MessageCount: 0}
		json.NewEncoder(w).Encode(resp)
	})
	defer srv.Close()

	_, err := p.Send(context.Background(), gosms.NewMessage("+15551234567", "hello"))
	if !errors.Is(err, gosms.ErrSendFailed) {
		t.Errorf("error = %v, want ErrSendFailed", err)
	}
}

func TestSendWithReference(t *testing.T) {
	srv, p := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		var req sendRequest
		json.NewDecoder(r.Body).Decode(&req)
		if req.ClientRef != "ref-123" {
			t.Errorf("client-ref = %q, want %q", req.ClientRef, "ref-123")
		}
		resp := sendResponse{
			MessageCount: 1,
			Messages: []struct {
				To               string `json:"to"`
				MessageID        string `json:"message-id"`
				Status           string `json:"status"`
				RemainingBalance string `json:"remaining-balance"`
				MessagePrice     string `json:"message-price"`
				Currency         string `json:"currency,omitempty"`
				Network          string `json:"network"`
				ErrorText        string `json:"error-text,omitempty"`
			}{
				{Status: "0", MessageID: "msg-002"},
			},
		}
		json.NewEncoder(w).Encode(resp)
	})
	defer srv.Close()

	msg := gosms.NewMessage("+15551234567", "hello").WithReference("ref-123")
	_, err := p.Send(context.Background(), msg)
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}
}

func TestGetStatusUnsupported(t *testing.T) {
	p, _ := NewProvider(Config{APIKey: "key", APISecret: "secret"})
	_, err := p.GetStatus(context.Background(), "msg-001")
	if !errors.Is(err, gosms.ErrUnsupported) {
		t.Errorf("error = %v, want ErrUnsupported", err)
	}
}

func TestSendBulk(t *testing.T) {
	callCount := 0
	srv, p := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		resp := sendResponse{
			MessageCount: 1,
			Messages: []struct {
				To               string `json:"to"`
				MessageID        string `json:"message-id"`
				Status           string `json:"status"`
				RemainingBalance string `json:"remaining-balance"`
				MessagePrice     string `json:"message-price"`
				Currency         string `json:"currency,omitempty"`
				Network          string `json:"network"`
				ErrorText        string `json:"error-text,omitempty"`
			}{
				{Status: "0", MessageID: "msg-bulk"},
			},
		}
		json.NewEncoder(w).Encode(resp)
	})
	defer srv.Close()

	msgs := []*gosms.Message{
		gosms.NewMessage("+15551111111", "a"),
		gosms.NewMessage("+15552222222", "b"),
	}

	results, err := p.SendBulk(context.Background(), msgs)
	if err != nil {
		t.Fatalf("SendBulk() error = %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("len(results) = %d, want 2", len(results))
	}
	if callCount != 2 {
		t.Errorf("callCount = %d, want 2", callCount)
	}
}

func TestMapStatus(t *testing.T) {
	tests := []struct {
		input string
		want  gosms.DeliveryStatus
	}{
		{"delivered", gosms.StatusDelivered},
		{"accepted", gosms.StatusAccepted},
		{"buffered", gosms.StatusQueued},
		{"failed", gosms.StatusFailed},
		{"expired", gosms.StatusExpired},
		{"rejected", gosms.StatusRejected},
		{"something_else", gosms.StatusUnknown},
	}

	for _, tt := range tests {
		if got := mapStatus(tt.input); got != tt.want {
			t.Errorf("mapStatus(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestParseErrorCodes(t *testing.T) {
	tests := []struct {
		status  string
		wantErr error
	}{
		{"1", gosms.ErrRateLimited},
		{"2", gosms.ErrInvalidConfig},
		{"3", gosms.ErrInvalidConfig},
		{"4", gosms.ErrInvalidConfig},
		{"5", gosms.ErrProviderError},
		{"6", gosms.ErrInvalidMessage},
		{"7", gosms.ErrBlacklisted},
		{"8", gosms.ErrInvalidConfig},
		{"9", gosms.ErrInsufficientFunds},
		{"15", gosms.ErrInvalidConfig},
		{"99", gosms.ErrProviderError},
	}

	for _, tt := range tests {
		err := parseError(tt.status, "test error")
		if !errors.Is(err, tt.wantErr) {
			t.Errorf("status %q: error = %v, want %v", tt.status, err, tt.wantErr)
		}
	}
}

func TestParseWebhookGET(t *testing.T) {
	r := httptest.NewRequest("GET", "/webhook?messageId=msg-001&status=delivered&err-code=0&msisdn=15551234567&to=15550000000&network-code=12345&price=0.05", nil)

	status, err := ParseWebhook(r)
	if err != nil {
		t.Fatalf("ParseWebhook() error = %v", err)
	}
	if status.MessageID != "msg-001" {
		t.Errorf("MessageID = %q", status.MessageID)
	}
	if status.Status != gosms.StatusDelivered {
		t.Errorf("Status = %q", status.Status)
	}
}

func TestParseWebhookPOST(t *testing.T) {
	dlr := `{"messageId":"msg-002","status":"delivered","msisdn":"15551234567","to":"15550000000","network-code":"12345","price":"0.05","scts":"2024010112","err-code":"0","message-timestamp":"2024-01-01 12:00:00"}`

	r := httptest.NewRequest("POST", "/webhook", strings.NewReader(dlr))
	r.Header.Set("Content-Type", "application/json")

	status, err := ParseWebhook(r)
	if err != nil {
		t.Fatalf("ParseWebhook() error = %v", err)
	}
	if status.MessageID != "msg-002" {
		t.Errorf("MessageID = %q", status.MessageID)
	}
	if status.Status != gosms.StatusDelivered {
		t.Errorf("Status = %q", status.Status)
	}
	if status.ErrorCode != "0" {
		t.Errorf("ErrorCode = %q", status.ErrorCode)
	}
}

func TestParseWebhookPOSTInvalidJSON(t *testing.T) {
	r := httptest.NewRequest("POST", "/webhook?messageId=msg-003&status=failed", strings.NewReader("not json"))
	r.Header.Set("Content-Type", "application/json")

	_, err := ParseWebhook(r)
	if err == nil {
		t.Fatal("ParseWebhook() should return error for invalid JSON body")
	}
}
