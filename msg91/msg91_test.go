package msg91

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	gosms "github.com/KARTIKrocks/gosms"
)

func newTestServer(handler http.HandlerFunc) (*httptest.Server, *Provider) {
	srv := httptest.NewServer(handler)
	p, _ := NewProvider(Config{
		AuthKey:    "test_authkey",
		SenderID:   "TESTID",
		TemplateID: "tmpl_default",
		BaseURL:    srv.URL,
	})
	return srv, p
}

func TestNewProviderValidation(t *testing.T) {
	if _, err := NewProvider(Config{}); !errors.Is(err, gosms.ErrInvalidConfig) {
		t.Errorf("empty config: error = %v, want ErrInvalidConfig", err)
	}

	p, err := NewProvider(Config{AuthKey: "k"})
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if p.Name() != "msg91" {
		t.Errorf("Name() = %q", p.Name())
	}
	if p.config.Country != "91" {
		t.Errorf("default Country = %q, want 91", p.config.Country)
	}
	if p.config.Route != RouteTransactional {
		t.Errorf("default Route = %q, want %q", p.config.Route, RouteTransactional)
	}
}

func assertFlowRequest(t *testing.T, r *http.Request) {
	t.Helper()
	if r.Method != http.MethodPost {
		t.Errorf("method = %q", r.Method)
	}
	if r.URL.Path != flowPath {
		t.Errorf("path = %q", r.URL.Path)
	}
	if r.Header.Get("authkey") != "test_authkey" {
		t.Errorf("authkey header = %q", r.Header.Get("authkey"))
	}
	if r.Header.Get("Content-Type") != "application/json" {
		t.Errorf("content-type = %q", r.Header.Get("Content-Type"))
	}

	body, _ := io.ReadAll(r.Body)
	var fr flowRequest
	if err := json.Unmarshal(body, &fr); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if fr.TemplateID != "tmpl_default" {
		t.Errorf("template_id = %q", fr.TemplateID)
	}
	if fr.SenderID != "TESTID" {
		t.Errorf("sender = %q", fr.SenderID)
	}
	if fr.Route != string(RouteTransactional) {
		t.Errorf("route = %q", fr.Route)
	}
	if len(fr.Recipients) != 1 {
		t.Fatalf("recipients len = %d", len(fr.Recipients))
	}
	if fr.Recipients[0]["mobiles"] != "919876543210" {
		t.Errorf("mobiles = %v", fr.Recipients[0]["mobiles"])
	}
}

func TestSendSuccess(t *testing.T) {
	srv, p := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		assertFlowRequest(t, r)
		_ = json.NewEncoder(w).Encode(flowResponse{Type: "success", Message: "req_abc123"})
	})
	defer srv.Close()

	msg := gosms.NewMessage("+919876543210", "hello")
	res, err := p.Send(context.Background(), msg)
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	if res.MessageID != "req_abc123" {
		t.Errorf("MessageID = %q", res.MessageID)
	}
	if res.Status != gosms.StatusAccepted {
		t.Errorf("Status = %q", res.Status)
	}
	if res.Provider != "msg91" {
		t.Errorf("Provider = %q", res.Provider)
	}
	if res.Segments != 1 {
		t.Errorf("Segments = %d", res.Segments)
	}
}

func TestSendRequiresTemplateID(t *testing.T) {
	p, _ := NewProvider(Config{AuthKey: "k", BaseURL: "http://localhost"})
	_, err := p.Send(context.Background(), gosms.NewMessage("+919876543210", "hi"))
	if !errors.Is(err, gosms.ErrInvalidConfig) {
		t.Errorf("error = %v, want ErrInvalidConfig", err)
	}
}

func TestSendWithVars(t *testing.T) {
	srv, p := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var fr flowRequest
		_ = json.Unmarshal(body, &fr)

		rec := fr.Recipients[0]
		if rec["name"] != "Kartik" {
			t.Errorf("name var = %v", rec["name"])
		}
		if rec["otp"] != "1234" {
			t.Errorf("otp var = %v", rec["otp"])
		}
		if _, ok := rec["body"]; ok {
			t.Error("body fallback should not be set when vars are present")
		}
		_ = json.NewEncoder(w).Encode(flowResponse{Type: "success", Message: "req_vars"})
	})
	defer srv.Close()

	msg := gosms.NewMessage("+919876543210", "")
	SetVar(msg, "name", "Kartik")
	SetVar(msg, "otp", "1234")

	if _, err := p.Send(context.Background(), msg); err != nil {
		t.Fatalf("Send() error = %v", err)
	}
}

func TestSendBodyFallbackWhenNoVars(t *testing.T) {
	srv, p := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var fr flowRequest
		_ = json.Unmarshal(body, &fr)
		if fr.Recipients[0]["body"] != "hello world" {
			t.Errorf("body fallback = %v", fr.Recipients[0]["body"])
		}
		_ = json.NewEncoder(w).Encode(flowResponse{Type: "success", Message: "req_x"})
	})
	defer srv.Close()

	if _, err := p.Send(context.Background(), gosms.NewMessage("+919876543210", "hello world")); err != nil {
		t.Fatalf("Send() error = %v", err)
	}
}

func TestSendPerMessageTemplateOverride(t *testing.T) {
	srv, p := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var fr flowRequest
		_ = json.Unmarshal(body, &fr)
		if fr.TemplateID != "tmpl_override" {
			t.Errorf("template_id = %q, want override", fr.TemplateID)
		}
		_ = json.NewEncoder(w).Encode(flowResponse{Type: "success", Message: "req_o"})
	})
	defer srv.Close()

	msg := gosms.NewMessage("+919876543210", "hi")
	SetTemplateID(msg, "tmpl_override")
	if _, err := p.Send(context.Background(), msg); err != nil {
		t.Fatalf("Send() error = %v", err)
	}
}

func TestSendErrorMapping(t *testing.T) {
	tests := []struct {
		name    string
		message string
		wantErr error
	}{
		{"invalid mobile", "Invalid mobile number", gosms.ErrInvalidPhone},
		{"bad authkey", "Invalid authkey", gosms.ErrInvalidConfig},
		{"low balance", "Insufficient balance", gosms.ErrInsufficientFunds},
		{"rate limit", "Rate limit exceeded", gosms.ErrRateLimited},
		{"dnd", "Number is in DND list", gosms.ErrBlacklisted},
		{"bad template", "Template not approved", gosms.ErrInvalidConfig},
		{"unknown", "Some other failure", gosms.ErrProviderError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := parseFlowError(flowResponse{Type: "error", Message: tt.message})
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestSendHTTPError(t *testing.T) {
	srv, p := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		_ = json.NewEncoder(w).Encode(flowResponse{Type: "error", Message: "Invalid authkey"})
	})
	defer srv.Close()

	_, err := p.Send(context.Background(), gosms.NewMessage("+919876543210", "hi"))
	if !errors.Is(err, gosms.ErrInvalidConfig) {
		t.Errorf("error = %v, want ErrInvalidConfig", err)
	}
}

func TestSendHTTPErrorNonJSON(t *testing.T) {
	srv, p := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(502)
		_, _ = w.Write([]byte("<html>Bad Gateway</html>"))
	})
	defer srv.Close()

	_, err := p.Send(context.Background(), gosms.NewMessage("+919876543210", "hi"))
	if !errors.Is(err, gosms.ErrProviderError) {
		t.Fatalf("error = %v, want ErrProviderError", err)
	}
	if !strings.Contains(err.Error(), "502") {
		t.Errorf("error = %q, want status 502 mentioned", err.Error())
	}
}

func TestBaseURLTrailingSlashTrimmed(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != flowPath {
			t.Errorf("path = %q, want %q (trailing slash should be trimmed)", r.URL.Path, flowPath)
		}
		_ = json.NewEncoder(w).Encode(flowResponse{Type: "success", Message: "req_ok"})
	}))
	defer srv.Close()

	p, err := NewProvider(Config{
		AuthKey:    "k",
		TemplateID: "t",
		BaseURL:    srv.URL + "/",
	})
	if err != nil {
		t.Fatalf("NewProvider error = %v", err)
	}
	if _, err := p.Send(context.Background(), gosms.NewMessage("+919876543210", "hi")); err != nil {
		t.Fatalf("Send() error = %v", err)
	}
}

func TestSendBulkBatchesByTemplate(t *testing.T) {
	callCount := 0
	recipientsSeen := []int{}
	srv, p := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		body, _ := io.ReadAll(r.Body)
		var fr flowRequest
		_ = json.Unmarshal(body, &fr)
		recipientsSeen = append(recipientsSeen, len(fr.Recipients))
		_ = json.NewEncoder(w).Encode(flowResponse{Type: "success", Message: "req_" + fr.TemplateID})
	})
	defer srv.Close()

	m1 := gosms.NewMessage("+919876500001", "a")
	m2 := gosms.NewMessage("+919876500002", "b")
	m3 := gosms.NewMessage("+919876500003", "c")
	SetTemplateID(m3, "tmpl_other")

	results, err := p.SendBulk(context.Background(), []*gosms.Message{m1, m2, m3})
	if err != nil {
		t.Fatalf("SendBulk() error = %v", err)
	}
	if callCount != 2 {
		t.Errorf("callCount = %d, want 2 (one per template group)", callCount)
	}
	if len(results) != 3 {
		t.Fatalf("len(results) = %d", len(results))
	}
	for i, r := range results {
		if r == nil {
			t.Fatalf("result[%d] is nil", i)
		}
		if r.Status != gosms.StatusAccepted {
			t.Errorf("result[%d].Status = %q", i, r.Status)
		}
	}
	if results[0].MessageID != "req_tmpl_default" || results[1].MessageID != "req_tmpl_default" {
		t.Error("first two should share request ID")
	}
	if results[2].MessageID != "req_tmpl_other" {
		t.Errorf("result[2] MessageID = %q", results[2].MessageID)
	}
}

func TestSendBulkEmpty(t *testing.T) {
	p, _ := NewProvider(Config{AuthKey: "k", TemplateID: "t"})
	res, err := p.SendBulk(context.Background(), nil)
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if res != nil {
		t.Errorf("results = %v, want nil", res)
	}
}

func TestSendBulkGroupErrorIsolated(t *testing.T) {
	srv, p := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var fr flowRequest
		_ = json.Unmarshal(body, &fr)
		if fr.TemplateID == "tmpl_bad" {
			w.WriteHeader(400)
			_ = json.NewEncoder(w).Encode(flowResponse{Type: "error", Message: "Insufficient balance"})
			return
		}
		_ = json.NewEncoder(w).Encode(flowResponse{Type: "success", Message: "req_ok"})
	})
	defer srv.Close()

	good := gosms.NewMessage("+919876500001", "good")
	bad := gosms.NewMessage("+919876500002", "bad")
	SetTemplateID(bad, "tmpl_bad")

	results, err := p.SendBulk(context.Background(), []*gosms.Message{good, bad})
	if err != nil {
		t.Fatalf("unexpected error = %v", err)
	}
	if results[0].Status != gosms.StatusAccepted {
		t.Errorf("good.Status = %q", results[0].Status)
	}
	if results[1].Status != gosms.StatusFailed {
		t.Errorf("bad.Status = %q", results[1].Status)
	}
	if !strings.Contains(results[1].Error, "balance") {
		t.Errorf("bad.Error = %q", results[1].Error)
	}
}

func TestSendBulkSplitsByMixedSenders(t *testing.T) {
	sendersSeen := []string{}
	srv, p := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var fr flowRequest
		_ = json.Unmarshal(body, &fr)
		sendersSeen = append(sendersSeen, fr.SenderID)
		_ = json.NewEncoder(w).Encode(flowResponse{Type: "success", Message: "req_" + fr.SenderID})
	})
	defer srv.Close()

	m1 := gosms.NewMessage("+919876500001", "a").WithFrom("SENDR1")
	m2 := gosms.NewMessage("+919876500002", "b").WithFrom("SENDR1")
	m3 := gosms.NewMessage("+919876500003", "c").WithFrom("SENDR2")

	results, err := p.SendBulk(context.Background(), []*gosms.Message{m1, m2, m3})
	if err != nil {
		t.Fatalf("SendBulk() error = %v", err)
	}
	if len(sendersSeen) != 2 {
		t.Fatalf("sendersSeen = %v, want 2 calls (one per sender)", sendersSeen)
	}
	if results[0].MessageID != "req_SENDR1" || results[1].MessageID != "req_SENDR1" {
		t.Error("first two should share SENDR1 request_id")
	}
	if results[2].MessageID != "req_SENDR2" {
		t.Errorf("result[2].MessageID = %q", results[2].MessageID)
	}
}

func TestSendBulkChunksLargeGroups(t *testing.T) {
	recipientsPerCall := []int{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var fr flowRequest
		_ = json.Unmarshal(body, &fr)
		recipientsPerCall = append(recipientsPerCall, len(fr.Recipients))
		_ = json.NewEncoder(w).Encode(flowResponse{Type: "success", Message: fmt.Sprintf("req_%d", len(recipientsPerCall))})
	}))
	defer srv.Close()

	p, _ := NewProvider(Config{
		AuthKey:              "k",
		SenderID:             "TESTID",
		TemplateID:           "tmpl_default",
		BaseURL:              srv.URL,
		MaxRecipientsPerCall: 2,
	})

	msgs := []*gosms.Message{
		gosms.NewMessage("+919876500001", "a"),
		gosms.NewMessage("+919876500002", "b"),
		gosms.NewMessage("+919876500003", "c"),
		gosms.NewMessage("+919876500004", "d"),
		gosms.NewMessage("+919876500005", "e"),
	}

	results, err := p.SendBulk(context.Background(), msgs)
	if err != nil {
		t.Fatalf("SendBulk() error = %v", err)
	}
	if got, want := recipientsPerCall, []int{2, 2, 1}; !equalInts(got, want) {
		t.Errorf("recipientsPerCall = %v, want %v", got, want)
	}
	if len(results) != 5 {
		t.Fatalf("len(results) = %d", len(results))
	}
	for i, r := range results {
		if r == nil || r.Status != gosms.StatusAccepted {
			t.Errorf("result[%d] = %+v", i, r)
		}
	}
	// Different chunks yield different request IDs.
	if results[0].MessageID == results[4].MessageID {
		t.Errorf("chunks should have distinct request_ids; got shared %q", results[0].MessageID)
	}
}

func TestSendBulkChunkingDisabledWithNegativeCap(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		_ = json.NewEncoder(w).Encode(flowResponse{Type: "success", Message: "req"})
	}))
	defer srv.Close()

	p, _ := NewProvider(Config{
		AuthKey:              "k",
		SenderID:             "TESTID",
		TemplateID:           "tmpl_default",
		BaseURL:              srv.URL,
		MaxRecipientsPerCall: -1,
	})

	msgs := make([]*gosms.Message, 50)
	for i := range msgs {
		msgs[i] = gosms.NewMessage(fmt.Sprintf("+91987650%04d", i), "hi")
	}
	if _, err := p.SendBulk(context.Background(), msgs); err != nil {
		t.Fatalf("SendBulk() error = %v", err)
	}
	if calls != 1 {
		t.Errorf("calls = %d, want 1 (chunking disabled)", calls)
	}
}

func equalInts(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestGetStatusUnsupported(t *testing.T) {
	p, _ := NewProvider(Config{AuthKey: "k"})
	_, err := p.GetStatus(context.Background(), "req_x")
	if !errors.Is(err, gosms.ErrUnsupported) {
		t.Errorf("error = %v, want ErrUnsupported", err)
	}
}

func TestSendOTPSuccess(t *testing.T) {
	srv, p := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != otpSendPath {
			t.Errorf("path = %q, want %q", r.URL.Path, otpSendPath)
		}
		if r.Header.Get("authkey") != "test_authkey" {
			t.Errorf("authkey header = %q", r.Header.Get("authkey"))
		}
		q := r.URL.Query()
		if q.Get("template_id") != "tmpl_default" {
			t.Errorf("template_id = %q", q.Get("template_id"))
		}
		if q.Get("mobile") != "919876543210" {
			t.Errorf("mobile = %q", q.Get("mobile"))
		}
		if q.Get("otp") != "" {
			t.Errorf("otp = %q, want empty (server-generated)", q.Get("otp"))
		}
		_ = json.NewEncoder(w).Encode(otpResponse{Type: "success", Message: "req_otp_001"})
	})
	defer srv.Close()

	res, err := p.SendOTP(context.Background(), &gosms.OTPRequest{Phone: "+919876543210"})
	if err != nil {
		t.Fatalf("SendOTP() error = %v", err)
	}
	if res.MessageID != "req_otp_001" {
		t.Errorf("MessageID = %q", res.MessageID)
	}
	if res.Phone != "+919876543210" {
		t.Errorf("Phone = %q", res.Phone)
	}
}

func TestSendOTPWithExplicitCode(t *testing.T) {
	srv, p := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("otp") != "4242" {
			t.Errorf("otp = %q, want 4242", q.Get("otp"))
		}
		if q.Get("otp_length") != "4" {
			t.Errorf("otp_length = %q", q.Get("otp_length"))
		}
		if q.Get("otp_expiry") != "5" {
			t.Errorf("otp_expiry minutes = %q", q.Get("otp_expiry"))
		}
		_ = json.NewEncoder(w).Encode(otpResponse{Type: "success", Message: "req_otp_002"})
	})
	defer srv.Close()

	_, err := p.SendOTP(context.Background(), &gosms.OTPRequest{
		Phone:  "+919876543210",
		OTP:    "4242",
		Length: 4,
		Expiry: 5 * time.Minute,
	})
	if err != nil {
		t.Fatalf("SendOTP() error = %v", err)
	}
}

func TestSendOTPPerCallTemplateOverride(t *testing.T) {
	srv, p := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("template_id"); got != "tmpl_otp_x" {
			t.Errorf("template_id = %q, want tmpl_otp_x", got)
		}
		_ = json.NewEncoder(w).Encode(otpResponse{Type: "success", Message: "req_otp_003"})
	})
	defer srv.Close()

	_, err := p.SendOTP(context.Background(), &gosms.OTPRequest{
		Phone:      "+919876543210",
		TemplateID: "tmpl_otp_x",
	})
	if err != nil {
		t.Fatalf("SendOTP() error = %v", err)
	}
}

func TestSendOTPVarsSentAsJSONBody(t *testing.T) {
	srv, p := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("content-type = %q", ct)
		}
		body, _ := io.ReadAll(r.Body)
		var got map[string]string
		if err := json.Unmarshal(body, &got); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if got["name"] != "Alice" {
			t.Errorf("name var = %q", got["name"])
		}
		_ = json.NewEncoder(w).Encode(otpResponse{Type: "success", Message: "req_otp_004"})
	})
	defer srv.Close()

	_, err := p.SendOTP(context.Background(), &gosms.OTPRequest{
		Phone: "+919876543210",
		Vars:  map[string]string{"name": "Alice"},
	})
	if err != nil {
		t.Fatalf("SendOTP() error = %v", err)
	}
}

func TestSendOTPRequiresPhone(t *testing.T) {
	p, _ := NewProvider(Config{AuthKey: "k", TemplateID: "t"})
	if _, err := p.SendOTP(context.Background(), &gosms.OTPRequest{}); !errors.Is(err, gosms.ErrInvalidConfig) {
		t.Errorf("error = %v, want ErrInvalidConfig", err)
	}
	if _, err := p.SendOTP(context.Background(), nil); !errors.Is(err, gosms.ErrInvalidConfig) {
		t.Errorf("nil request: error = %v, want ErrInvalidConfig", err)
	}
}

func TestSendOTPRequiresTemplate(t *testing.T) {
	p, _ := NewProvider(Config{AuthKey: "k"})
	_, err := p.SendOTP(context.Background(), &gosms.OTPRequest{Phone: "+919876543210"})
	if !errors.Is(err, gosms.ErrInvalidConfig) {
		t.Errorf("error = %v, want ErrInvalidConfig", err)
	}
}

func TestSendOTPProviderError(t *testing.T) {
	srv, p := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
		_ = json.NewEncoder(w).Encode(otpResponse{Type: "error", Message: "Invalid authkey"})
	})
	defer srv.Close()

	_, err := p.SendOTP(context.Background(), &gosms.OTPRequest{Phone: "+919876543210"})
	if !errors.Is(err, gosms.ErrInvalidConfig) {
		t.Errorf("error = %v, want ErrInvalidConfig", err)
	}
}

func TestVerifyOTPSuccess(t *testing.T) {
	srv, p := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != otpVerifyPath {
			t.Errorf("path = %q", r.URL.Path)
		}
		if r.URL.Query().Get("mobile") != "919876543210" {
			t.Errorf("mobile = %q", r.URL.Query().Get("mobile"))
		}
		if r.URL.Query().Get("otp") != "1234" {
			t.Errorf("otp = %q", r.URL.Query().Get("otp"))
		}
		_ = json.NewEncoder(w).Encode(otpResponse{Type: "success", Message: "OTP verified"})
	})
	defer srv.Close()

	res, err := p.VerifyOTP(context.Background(), "+919876543210", "1234")
	if err != nil {
		t.Fatalf("VerifyOTP() error = %v", err)
	}
	if !res.Verified {
		t.Error("Verified = false, want true")
	}
}

func TestVerifyOTPFailure(t *testing.T) {
	srv, p := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(otpResponse{Type: "error", Message: "OTP mismatch"})
	})
	defer srv.Close()

	res, err := p.VerifyOTP(context.Background(), "+919876543210", "9999")
	if err != nil {
		t.Fatalf("VerifyOTP() error = %v", err)
	}
	if res.Verified {
		t.Error("Verified = true, want false")
	}
	if res.Message != "OTP mismatch" {
		t.Errorf("Message = %q", res.Message)
	}
}

func TestVerifyOTPHTTPAuthErrorReturnsError(t *testing.T) {
	srv, p := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
		_ = json.NewEncoder(w).Encode(otpResponse{Type: "error", Message: "Invalid authkey"})
	})
	defer srv.Close()

	res, err := p.VerifyOTP(context.Background(), "+919876543210", "1234")
	if !errors.Is(err, gosms.ErrInvalidConfig) {
		t.Fatalf("error = %v, want ErrInvalidConfig", err)
	}
	if res != nil {
		t.Errorf("result = %v, want nil on transport failure", res)
	}
}

func TestVerifyOTPMismatchOn400ReturnsVerifiedFalse(t *testing.T) {
	srv, p := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		_ = json.NewEncoder(w).Encode(otpResponse{Type: "error", Message: "OTP not match"})
	})
	defer srv.Close()

	res, err := p.VerifyOTP(context.Background(), "+919876543210", "9999")
	if err != nil {
		t.Fatalf("VerifyOTP() error = %v, want nil for wrong-OTP", err)
	}
	if res.Verified {
		t.Error("Verified = true, want false")
	}
}

func TestVerifyOTPNonJSONError(t *testing.T) {
	srv, p := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(503)
		_, _ = w.Write([]byte("<html>Service Unavailable</html>"))
	})
	defer srv.Close()

	_, err := p.VerifyOTP(context.Background(), "+919876543210", "1234")
	if !errors.Is(err, gosms.ErrProviderError) {
		t.Fatalf("error = %v, want ErrProviderError", err)
	}
	if !strings.Contains(err.Error(), "503") {
		t.Errorf("error = %q, want status 503 mentioned", err.Error())
	}
}

func TestVerifyOTPRequiresInput(t *testing.T) {
	p, _ := NewProvider(Config{AuthKey: "k"})
	if _, err := p.VerifyOTP(context.Background(), "", "1234"); !errors.Is(err, gosms.ErrInvalidConfig) {
		t.Errorf("missing phone: error = %v", err)
	}
	if _, err := p.VerifyOTP(context.Background(), "+91987", ""); !errors.Is(err, gosms.ErrInvalidConfig) {
		t.Errorf("missing otp: error = %v", err)
	}
}

func TestResendOTPSuccess(t *testing.T) {
	srv, p := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != otpResendPath {
			t.Errorf("path = %q", r.URL.Path)
		}
		if r.URL.Query().Get("retrytype") != "voice" {
			t.Errorf("retrytype = %q", r.URL.Query().Get("retrytype"))
		}
		_ = json.NewEncoder(w).Encode(otpResponse{Type: "success", Message: "retried"})
	})
	defer srv.Close()

	if err := p.ResendOTP(context.Background(), "+919876543210", "voice"); err != nil {
		t.Fatalf("ResendOTP() error = %v", err)
	}
}

func TestResendOTPDefaultChannel(t *testing.T) {
	srv, p := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("retrytype") != "text" {
			t.Errorf("retrytype = %q, want text", r.URL.Query().Get("retrytype"))
		}
		_ = json.NewEncoder(w).Encode(otpResponse{Type: "success", Message: "retried"})
	})
	defer srv.Close()

	if err := p.ResendOTP(context.Background(), "+919876543210", ""); err != nil {
		t.Fatalf("ResendOTP() error = %v", err)
	}
}

func TestResendOTPInvalidChannel(t *testing.T) {
	p, _ := NewProvider(Config{AuthKey: "k"})
	err := p.ResendOTP(context.Background(), "+919876543210", "fax")
	if !errors.Is(err, gosms.ErrInvalidConfig) {
		t.Errorf("error = %v, want ErrInvalidConfig", err)
	}
}

func TestResendOTPProviderError(t *testing.T) {
	srv, p := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(otpResponse{Type: "error", Message: "boom"})
	})
	defer srv.Close()

	err := p.ResendOTP(context.Background(), "+919876543210", "text")
	if !errors.Is(err, gosms.ErrProviderError) {
		t.Errorf("error = %v, want ErrProviderError", err)
	}
}

func TestResendOTPAuthErrorMapsToConfig(t *testing.T) {
	srv, p := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
		_ = json.NewEncoder(w).Encode(otpResponse{Type: "error", Message: "Invalid authkey"})
	})
	defer srv.Close()

	err := p.ResendOTP(context.Background(), "+919876543210", "text")
	if !errors.Is(err, gosms.ErrInvalidConfig) {
		t.Errorf("error = %v, want ErrInvalidConfig", err)
	}
}

func TestParseWebhook(t *testing.T) {
	form := "requestId=req_xyz&mobile=919876543210&status=delivered&statusCode=1&description=Delivered+successfully"
	r := httptest.NewRequest(http.MethodPost, "/webhook?"+form, nil)
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	status, err := ParseWebhook(r)
	if err != nil {
		t.Fatalf("ParseWebhook() error = %v", err)
	}
	if status.MessageID != "req_xyz" {
		t.Errorf("MessageID = %q", status.MessageID)
	}
	if status.Status != gosms.StatusDelivered {
		t.Errorf("Status = %q", status.Status)
	}
	if status.ErrorCode != "1" {
		t.Errorf("ErrorCode = %q", status.ErrorCode)
	}
	if status.ErrorMessage != "Delivered successfully" {
		t.Errorf("ErrorMessage = %q", status.ErrorMessage)
	}
}

func TestParseWebhookFallsBackToNumericStatusCode(t *testing.T) {
	form := "requestId=req_num&mobile=919876543210&status=&statusCode=2&description=Absent+subscriber"
	r := httptest.NewRequest(http.MethodPost, "/webhook?"+form, nil)
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	status, err := ParseWebhook(r)
	if err != nil {
		t.Fatalf("ParseWebhook() error = %v", err)
	}
	if status.Status != gosms.StatusFailed {
		t.Errorf("Status = %q, want failed (via code 2)", status.Status)
	}
	if status.ErrorCode != "2" {
		t.Errorf("ErrorCode = %q", status.ErrorCode)
	}
}

func TestMapStatusCode(t *testing.T) {
	tests := []struct {
		in   string
		want gosms.DeliveryStatus
	}{
		{"1", gosms.StatusDelivered},
		{"2", gosms.StatusFailed},
		{"5", gosms.StatusQueued},
		{"8", gosms.StatusSent},
		{"9", gosms.StatusRejected},
		{"16", gosms.StatusExpired},
		{"17", gosms.StatusRejected},
		{"25", gosms.StatusRejected},
		{"26", gosms.StatusRejected},
		{"999", gosms.StatusUnknown},
		{"", gosms.StatusUnknown},
	}
	for _, tt := range tests {
		if got := mapStatusCode(tt.in); got != tt.want {
			t.Errorf("mapStatusCode(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestMapStatus(t *testing.T) {
	tests := []struct {
		in   string
		want gosms.DeliveryStatus
	}{
		{"delivered", gosms.StatusDelivered},
		{"DLVRD", gosms.StatusDelivered},
		{"sent", gosms.StatusSent},
		{"submitted", gosms.StatusSent},
		{"queued", gosms.StatusQueued},
		{"pending", gosms.StatusQueued},
		{"failed", gosms.StatusFailed},
		{"undelivered", gosms.StatusFailed},
		{"rejected", gosms.StatusRejected},
		{"ndnc", gosms.StatusRejected},
		{"dnd", gosms.StatusRejected},
		{"expired", gosms.StatusExpired},
		{"mystery", gosms.StatusUnknown},
	}

	for _, tt := range tests {
		if got := mapStatus(tt.in); got != tt.want {
			t.Errorf("mapStatus(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestNormalizeRecipient(t *testing.T) {
	p, _ := NewProvider(Config{AuthKey: "k"})
	tests := []struct {
		in   string
		want string
	}{
		{"+919876543210", "919876543210"},
		{"919876543210", "919876543210"},
		{"9876543210", "919876543210"},
		{"+1 (555) 123-4567", "15551234567"},
		{" +91 98765 43210 ", "919876543210"},
	}
	for _, tt := range tests {
		if got := p.normalizeRecipient(tt.in); got != tt.want {
			t.Errorf("normalizeRecipient(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestSetVarInitializesMetadata(t *testing.T) {
	msg := &gosms.Message{To: "+919876543210"}
	SetVar(msg, "name", "Kartik")
	if msg.Metadata[metaVarPrefix+"name"] != "Kartik" {
		t.Errorf("var not set: %v", msg.Metadata)
	}
}

func TestSetTemplateIDInitializesMetadata(t *testing.T) {
	msg := &gosms.Message{To: "+919876543210"}
	SetTemplateID(msg, "tmpl_x")
	if msg.Metadata[metaTemplateID] != "tmpl_x" {
		t.Errorf("template id not set: %v", msg.Metadata)
	}
}

func TestSenderFromMessageOverridesConfig(t *testing.T) {
	srv, p := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var fr flowRequest
		_ = json.Unmarshal(body, &fr)
		if fr.SenderID != "MYSNDR" {
			t.Errorf("sender = %q, want override", fr.SenderID)
		}
		_ = json.NewEncoder(w).Encode(flowResponse{Type: "success", Message: "req_s"})
	})
	defer srv.Close()

	msg := gosms.NewMessage("+919876543210", "hi").WithFrom("MYSNDR")
	if _, err := p.Send(context.Background(), msg); err != nil {
		t.Fatalf("Send() error = %v", err)
	}
}
