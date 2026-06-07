package twilio

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	gosms "github.com/KARTIKrocks/gosms"
)

func newTestServer(handler http.HandlerFunc) (*httptest.Server, *Provider) {
	srv := httptest.NewServer(handler)
	p, _ := NewProvider(Config{
		AccountSID: "AC_test",
		AuthToken:  "test_token",
		From:       "+15550000000",
		BaseURL:    srv.URL,
	})
	return srv, p
}

func TestNewProviderValidation(t *testing.T) {
	_, err := NewProvider(Config{})
	if !errors.Is(err, gosms.ErrInvalidConfig) {
		t.Errorf("error = %v, want ErrInvalidConfig", err)
	}

	_, err = NewProvider(Config{AccountSID: "AC_test"})
	if !errors.Is(err, gosms.ErrInvalidConfig) {
		t.Errorf("error = %v, want ErrInvalidConfig", err)
	}

	p, err := NewProvider(Config{AccountSID: "AC_test", AuthToken: "tok"})
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if p.Name() != "twilio" {
		t.Errorf("Name() = %q", p.Name())
	}
}

func TestSendSuccess(t *testing.T) {
	srv, p := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("method = %q, want POST", r.Method)
		}

		u, pw, ok := r.BasicAuth()
		if !ok || u != "AC_test" || pw != "test_token" {
			t.Error("basic auth incorrect")
		}

		if err := r.ParseForm(); err != nil {
			t.Fatal(err)
		}
		if r.FormValue("To") != "+15551234567" {
			t.Errorf("To = %q", r.FormValue("To"))
		}
		if r.FormValue("Body") != "hello" {
			t.Errorf("Body = %q", r.FormValue("Body"))
		}
		if r.FormValue("From") != "+15550000000" {
			t.Errorf("From = %q", r.FormValue("From"))
		}

		resp := messageResponse{
			SID:         "SM_test123",
			To:          "+15551234567",
			Status:      "queued",
			NumSegments: 1,
			Price:       "0.0075",
			PriceUnit:   "USD",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	defer srv.Close()

	result, err := p.Send(context.Background(), gosms.NewMessage("+15551234567", "hello"))
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	if result.MessageID != "SM_test123" {
		t.Errorf("MessageID = %q", result.MessageID)
	}
	if result.Status != gosms.StatusQueued {
		t.Errorf("Status = %q, want %q", result.Status, gosms.StatusQueued)
	}
	if result.Cost != "0.0075" {
		t.Errorf("Cost = %q", result.Cost)
	}
	if result.Segments != 1 {
		t.Errorf("Segments = %d", result.Segments)
	}
}

func TestSendNoFrom(t *testing.T) {
	p, _ := NewProvider(Config{
		AccountSID: "AC_test",
		AuthToken:  "test_token",
		BaseURL:    "http://localhost",
	})

	_, err := p.Send(context.Background(), gosms.NewMessage("+15551234567", "hello"))
	if !errors.Is(err, gosms.ErrInvalidConfig) {
		t.Errorf("error = %v, want ErrInvalidConfig", err)
	}
}

func TestSendMessagingServiceSID(t *testing.T) {
	srv, _ := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		if r.FormValue("MessagingServiceSid") != "MG_test" {
			t.Errorf("MessagingServiceSid = %q", r.FormValue("MessagingServiceSid"))
		}
		if r.FormValue("From") != "" {
			t.Error("From should not be set when MessagingServiceSid is used")
		}
		resp := messageResponse{SID: "SM_test", Status: "queued", NumSegments: 1}
		json.NewEncoder(w).Encode(resp)
	})
	defer srv.Close()

	p, _ := NewProvider(Config{
		AccountSID:          "AC_test",
		AuthToken:           "test_token",
		MessagingServiceSID: "MG_test",
		BaseURL:             srv.URL,
	})

	_, err := p.Send(context.Background(), gosms.NewMessage("+15551234567", "hello"))
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}
}

func TestSendHTTPError(t *testing.T) {
	srv, p := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		resp := messageResponse{Code: 21211, Message: "Invalid 'To' Phone Number"}
		json.NewEncoder(w).Encode(resp)
	})
	defer srv.Close()

	_, err := p.Send(context.Background(), gosms.NewMessage("+15551234567", "hello"))
	if !errors.Is(err, gosms.ErrInvalidPhone) {
		t.Errorf("error = %v, want ErrInvalidPhone", err)
	}
}

func TestSendErrorCodes(t *testing.T) {
	tests := []struct {
		code    int
		wantErr error
	}{
		{21211, gosms.ErrInvalidPhone},
		{21614, gosms.ErrInvalidPhone},
		{21408, gosms.ErrBlacklisted},
		{20003, gosms.ErrInsufficientFunds},
		{20429, gosms.ErrRateLimited},
		{99999, gosms.ErrProviderError},
	}

	for _, tt := range tests {
		resp := messageResponse{Code: tt.code, Message: "test"}
		err := parseError(resp)
		if !errors.Is(err, tt.wantErr) {
			t.Errorf("code %d: error = %v, want %v", tt.code, err, tt.wantErr)
		}
	}
}

func TestGetStatusInvalidID(t *testing.T) {
	p, _ := NewProvider(Config{
		AccountSID: "AC_test",
		AuthToken:  "test_token",
		BaseURL:    "http://localhost",
	})

	_, err := p.GetStatus(context.Background(), "../../../etc/passwd")
	if !errors.Is(err, gosms.ErrInvalidConfig) {
		t.Errorf("error = %v, want ErrInvalidConfig for path traversal", err)
	}

	_, err = p.GetStatus(context.Background(), "SM test 123")
	if !errors.Is(err, gosms.ErrInvalidConfig) {
		t.Errorf("error = %v, want ErrInvalidConfig for spaces", err)
	}
}

func TestGetStatusSuccess(t *testing.T) {
	srv, p := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("method = %q, want GET", r.Method)
		}
		resp := messageResponse{
			SID:    "SM_test123",
			Status: "delivered",
		}
		json.NewEncoder(w).Encode(resp)
	})
	defer srv.Close()

	status, err := p.GetStatus(context.Background(), "SM_test123")
	if err != nil {
		t.Fatalf("GetStatus() error = %v", err)
	}
	if status.Status != gosms.StatusDelivered {
		t.Errorf("Status = %q, want %q", status.Status, gosms.StatusDelivered)
	}
	if status.ErrorCode != "" {
		t.Errorf("ErrorCode = %q, want empty", status.ErrorCode)
	}
}

func TestGetStatusWithError(t *testing.T) {
	srv, p := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		resp := messageResponse{
			SID:          "SM_test123",
			Status:       "failed",
			ErrorCode:    30001,
			ErrorMessage: "Queue overflow",
		}
		json.NewEncoder(w).Encode(resp)
	})
	defer srv.Close()

	status, err := p.GetStatus(context.Background(), "SM_test123")
	if err != nil {
		t.Fatalf("GetStatus() error = %v", err)
	}
	if status.ErrorCode != "30001" {
		t.Errorf("ErrorCode = %q, want %q", status.ErrorCode, "30001")
	}
	if status.ErrorMessage != "Queue overflow" {
		t.Errorf("ErrorMessage = %q", status.ErrorMessage)
	}
}

func TestSendBulk(t *testing.T) {
	callCount := 0
	srv, p := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		resp := messageResponse{SID: "SM_bulk", Status: "queued", NumSegments: 1}
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
		{"queued", gosms.StatusQueued},
		{"accepted", gosms.StatusAccepted},
		{"sending", gosms.StatusSent},
		{"sent", gosms.StatusSent},
		{"delivered", gosms.StatusDelivered},
		{"undelivered", gosms.StatusFailed},
		{"failed", gosms.StatusFailed},
		{"canceled", gosms.StatusRejected},
		{"unknown_status", gosms.StatusUnknown},
	}

	for _, tt := range tests {
		if got := mapStatus(tt.input); got != tt.want {
			t.Errorf("mapStatus(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestFormatErrorCode(t *testing.T) {
	if got := formatErrorCode(0); got != "" {
		t.Errorf("formatErrorCode(0) = %q, want empty", got)
	}
	if got := formatErrorCode(30001); got != "30001" {
		t.Errorf("formatErrorCode(30001) = %q, want %q", got, "30001")
	}
}

func TestParseWebhook(t *testing.T) {
	form := "MessageSid=SM_test&MessageStatus=delivered&ErrorCode=&ErrorMessage=&AccountSid=AC_test&From=%2B15550000000&To=%2B15551234567"
	r := httptest.NewRequest("POST", "/webhook", nil)
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.Body = http.NoBody
	r.Form = nil
	// Use a proper form request
	r = httptest.NewRequest("POST", "/webhook?"+form, nil)
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	status, err := ParseWebhook(r)
	if err != nil {
		t.Fatalf("ParseWebhook() error = %v", err)
	}
	if status.MessageID != "SM_test" {
		t.Errorf("MessageID = %q", status.MessageID)
	}
	if status.Status != gosms.StatusDelivered {
		t.Errorf("Status = %q, want %q", status.Status, gosms.StatusDelivered)
	}
}

func TestSendWithStatusCallback(t *testing.T) {
	srv, _ := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		if r.FormValue("StatusCallback") != "https://example.com/callback" {
			t.Errorf("StatusCallback = %q", r.FormValue("StatusCallback"))
		}
		resp := messageResponse{SID: "SM_test", Status: "queued", NumSegments: 1}
		json.NewEncoder(w).Encode(resp)
	})
	defer srv.Close()

	p, _ := NewProvider(Config{
		AccountSID:     "AC_test",
		AuthToken:      "test_token",
		From:           "+15550000000",
		StatusCallback: "https://example.com/callback",
		BaseURL:        srv.URL,
	})

	_, err := p.Send(context.Background(), gosms.NewMessage("+15551234567", "hello"))
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}
}
