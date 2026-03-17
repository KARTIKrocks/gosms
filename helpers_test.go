package gosms

import (
	"context"
	"errors"
	"testing"
)

func TestValidateE164(t *testing.T) {
	tests := []struct {
		phone string
		want  bool
	}{
		{"+15551234567", true},
		{"+12", true},
		{"+44207946", true},
		{"+919876543210", true},
		{"", false},
		{"15551234567", false},
		{"+0551234567", false},
		{"+", false},
		{"+1234567890123456", false}, // 16 digits, too long
		{"+1 555 123 4567", false},   // spaces
		{"hello", false},
	}

	for _, tt := range tests {
		t.Run(tt.phone, func(t *testing.T) {
			if got := ValidateE164(tt.phone); got != tt.want {
				t.Errorf("ValidateE164(%q) = %v, want %v", tt.phone, got, tt.want)
			}
		})
	}
}

func TestNormalizePhone(t *testing.T) {
	tests := []struct {
		name        string
		phone       string
		countryCode string
		want        string
	}{
		{"already e164", "+15551234567", "1", "+15551234567"},
		{"bare number with code", "5551234567", "1", "+15551234567"},
		{"bare number with +code", "5551234567", "+1", "+15551234567"},
		{"with dashes", "+1-555-123-4567", "", "+15551234567"},
		{"with spaces", "+1 555 123 4567", "", "+15551234567"},
		{"with parens", "(555) 123-4567", "1", "+15551234567"},
		{"empty string", "", "", ""},
		{"empty with code", "", "1", ""},
		{"no country code", "5551234567", "", "5551234567"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NormalizePhone(tt.phone, tt.countryCode); got != tt.want {
				t.Errorf("NormalizePhone(%q, %q) = %q, want %q", tt.phone, tt.countryCode, got, tt.want)
			}
		})
	}
}

func TestIsGSMEncoding(t *testing.T) {
	tests := []struct {
		name    string
		message string
		want    bool
	}{
		{"ascii text", "Hello World!", true},
		{"gsm special chars", "@$", true},
		{"gsm extended", "{}[]", true},
		{"newline", "line1\nline2", true},
		{"emoji", "Hello 😀", false},
		{"chinese", "你好", false},
		{"empty", "", true},
		{"euro sign", "€100", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsGSMEncoding(tt.message); got != tt.want {
				t.Errorf("IsGSMEncoding(%q) = %v, want %v", tt.message, got, tt.want)
			}
		})
	}
}

func TestGSMLen(t *testing.T) {
	tests := []struct {
		name    string
		message string
		want    int
	}{
		{"basic chars", "Hello", 5},
		{"extended char", "{", 2},
		{"mixed", "Hi{}", 6}, // 2 basic + 2*2 extended
		{"empty", "", 0},
		{"euro", "€50", 4}, // 2 (euro) + 2 (digits)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GSMLen(tt.message); got != tt.want {
				t.Errorf("GSMLen(%q) = %d, want %d", tt.message, got, tt.want)
			}
		})
	}
}

func TestCalculateSegments(t *testing.T) {
	tests := []struct {
		name    string
		message string
		want    int
	}{
		{"empty", "", 0},
		{"short gsm", "Hello", 1},
		{"exactly 160 gsm", string(make([]byte, 160)), 1},
		{"161 gsm chars", string(append(make([]byte, 160), 'a')), 2},
		{"short unicode", "你好", 1},
		{"single segment max unicode", string(make([]rune, 70)), 1},
	}

	// Fill the byte slices with valid GSM characters
	tests[2].message = repeatChar('A', 160)
	tests[3].message = repeatChar('A', 161)
	tests[4].message = "你好"
	tests[5].message = repeatRune('你', 70)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CalculateSegments(tt.message); got != tt.want {
				t.Errorf("CalculateSegments() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestCalculateSegmentsMultipart(t *testing.T) {
	// 306 GSM chars = ceil(306/153) = 2 segments
	if got := CalculateSegments(repeatChar('A', 306)); got != 2 {
		t.Errorf("306 GSM chars = %d segments, want 2", got)
	}
	// 307 GSM chars = ceil(307/153) = 3 segments
	if got := CalculateSegments(repeatChar('A', 307)); got != 3 {
		t.Errorf("307 GSM chars = %d segments, want 3", got)
	}
	// 134 unicode chars = ceil(134/67) = 2 segments
	if got := CalculateSegments(repeatRune('你', 134)); got != 2 {
		t.Errorf("134 unicode chars = %d segments, want 2", got)
	}
	// 135 unicode chars = ceil(135/67) = 3 segments
	if got := CalculateSegments(repeatRune('你', 135)); got != 3 {
		t.Errorf("135 unicode chars = %d segments, want 3", got)
	}
}

func TestCalculateSegmentsExtendedGSM(t *testing.T) {
	// 80 euro signs = 160 septets = 1 segment
	if got := CalculateSegments(repeatRune('€', 80)); got != 1 {
		t.Errorf("80 euro signs = %d segments, want 1", got)
	}
	// 81 euro signs = 162 septets = 2 segments
	if got := CalculateSegments(repeatRune('€', 81)); got != 2 {
		t.Errorf("81 euro signs = %d segments, want 2", got)
	}
}

func repeatChar(c byte, n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = c
	}
	return string(b)
}

func repeatRune(r rune, n int) string {
	rs := make([]rune, n)
	for i := range rs {
		rs[i] = r
	}
	return string(rs)
}

func TestBatch(t *testing.T) {
	batch := NewBatch()
	if batch.Size() != 0 {
		t.Errorf("Size() = %d, want 0", batch.Size())
	}

	batch.AddNew("+15551111111", "msg1").
		AddNew("+15552222222", "msg2").
		AddNewWithFrom("+15553333333", "msg3", "+15550000000")

	if batch.Size() != 3 {
		t.Errorf("Size() = %d, want 3", batch.Size())
	}

	msgs := batch.Messages()
	if len(msgs) != 3 {
		t.Fatalf("len(Messages()) = %d, want 3", len(msgs))
	}
	if msgs[2].From != "+15550000000" {
		t.Errorf("msgs[2].From = %q, want %q", msgs[2].From, "+15550000000")
	}

	batch.Clear()
	if batch.Size() != 0 {
		t.Errorf("Size() after Clear = %d, want 0", batch.Size())
	}
}

func TestBatchSend(t *testing.T) {
	ctx := context.Background()
	mock := NewMockProvider()
	client := NewClient(mock)

	batch := NewBatch()
	batch.AddNew("+15551111111", "a").AddNew("+15552222222", "b")

	results, err := batch.Send(ctx, client)
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("len(results) = %d, want 2", len(results))
	}
	if mock.MessageCount() != 2 {
		t.Errorf("MessageCount = %d, want 2", mock.MessageCount())
	}
}

func TestSendToMany(t *testing.T) {
	ctx := context.Background()
	mock := NewMockProvider()
	client := NewClient(mock)

	results, err := SendToMany(ctx, client, "broadcast", "+15551111111", "+15552222222", "+15553333333")
	if err != nil {
		t.Fatalf("SendToMany() error = %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("len(results) = %d, want 3", len(results))
	}
	if mock.MessageCount() != 3 {
		t.Errorf("MessageCount = %d, want 3", mock.MessageCount())
	}
}

func TestQuickSend(t *testing.T) {
	ctx := context.Background()
	mock := NewMockProvider()

	result, err := QuickSend(ctx, mock, "+15551234567", "+15550000000", "quick test")
	if err != nil {
		t.Fatalf("QuickSend() error = %v", err)
	}
	if !result.Success() {
		t.Error("expected success")
	}
	if mock.LastMessage().Message.Body != "quick test" {
		t.Errorf("Body = %q", mock.LastMessage().Message.Body)
	}
}

func TestMultiProviderFallback(t *testing.T) {
	ctx := context.Background()

	failing := NewMockProvider().WithSendError(errors.New("down"))
	working := NewMockProvider()

	multi := NewMultiProvider(failing, working)
	client := NewClient(multi)

	result, err := client.Send(ctx, "+15551234567", "test")
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	if !result.Success() {
		t.Error("expected success via fallback")
	}
	if failing.MessageCount() != 0 {
		t.Error("failing provider should have 0 messages")
	}
	if working.MessageCount() != 1 {
		t.Errorf("working provider MessageCount = %d, want 1", working.MessageCount())
	}
}

func TestMultiProviderFallbackAllFail(t *testing.T) {
	ctx := context.Background()

	p1 := NewMockProvider().WithSendError(errors.New("err1"))
	p2 := NewMockProvider().WithSendError(errors.New("err2"))

	multi := NewMultiProvider(p1, p2)
	_, err := multi.Send(ctx, NewMessage("+15551234567", "test"))
	if err == nil {
		t.Error("expected error when all providers fail")
	}
	if err.Error() != "err2" {
		t.Errorf("error = %q, want last provider error %q", err.Error(), "err2")
	}
}

func TestMultiProviderRoundRobin(t *testing.T) {
	ctx := context.Background()

	p1 := NewMockProvider()
	p2 := NewMockProvider()

	multi := NewMultiProvider(p1, p2).WithStrategy(StrategyRoundRobin)
	client := NewClient(multi)

	for i := range 4 {
		_, err := client.Send(ctx, "+15551234567", "test")
		if err != nil {
			t.Fatalf("Send() %d error = %v", i, err)
		}
	}

	if p1.MessageCount() != 2 {
		t.Errorf("p1 MessageCount = %d, want 2", p1.MessageCount())
	}
	if p2.MessageCount() != 2 {
		t.Errorf("p2 MessageCount = %d, want 2", p2.MessageCount())
	}
}

func TestMultiProviderNoProviders(t *testing.T) {
	multi := NewMultiProvider()
	_, err := multi.Send(context.Background(), NewMessage("+15551234567", "test"))
	if !errors.Is(err, ErrInvalidConfig) {
		t.Errorf("error = %v, want ErrInvalidConfig", err)
	}
}

func TestMultiProviderGetStatus(t *testing.T) {
	ctx := context.Background()
	mock := NewMockProvider()
	multi := NewMultiProvider(mock)

	result, _ := mock.Send(ctx, NewMessage("+15551234567", "test"))
	status, err := multi.GetStatus(ctx, result.MessageID)
	if err != nil {
		t.Fatalf("GetStatus() error = %v", err)
	}
	if status.Status != StatusDelivered {
		t.Errorf("Status = %q, want %q", status.Status, StatusDelivered)
	}
}

func TestMultiProviderName(t *testing.T) {
	multi := NewMultiProvider()
	if got := multi.Name(); got != "multi" {
		t.Errorf("Name() = %q, want %q", got, "multi")
	}
}

func TestMultiProviderSendBulk(t *testing.T) {
	ctx := context.Background()
	mock := NewMockProvider()
	multi := NewMultiProvider(mock)

	msgs := []*Message{
		NewMessage("+15551111111", "a"),
		NewMessage("+15552222222", "b"),
	}

	results, err := multi.SendBulk(ctx, msgs)
	if err != nil {
		t.Fatalf("SendBulk() error = %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("len(results) = %d, want 2", len(results))
	}
}
