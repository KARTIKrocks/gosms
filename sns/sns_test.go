package sns

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"

	gosms "github.com/KARTIKrocks/gosms"
)

// publishResponse is the XML response for the SNS Publish API.
type publishResponse struct {
	XMLName       xml.Name `xml:"PublishResponse"`
	PublishResult struct {
		MessageId string `xml:"MessageId"`
	} `xml:"PublishResult"`
}

func writePublishResponse(w http.ResponseWriter, messageID string) {
	resp := publishResponse{}
	resp.PublishResult.MessageId = messageID
	w.Header().Set("Content-Type", "text/xml")
	w.WriteHeader(200)
	xml.NewEncoder(w).Encode(resp)
}

func writeEmptyXMLResponse(w http.ResponseWriter, action string) {
	w.Header().Set("Content-Type", "text/xml")
	fmt.Fprintf(w, `<%sResponse><%sResult></%sResult><ResponseMetadata><RequestId>test-id</RequestId></ResponseMetadata></%sResponse>`, action, action, action, action)
}

func writeXMLError(w http.ResponseWriter, code int, errType, message string) {
	w.Header().Set("Content-Type", "text/xml")
	w.WriteHeader(code)
	fmt.Fprintf(w, `<ErrorResponse><Error><Type>Sender</Type><Code>%s</Code><Message>%s</Message></Error></ErrorResponse>`, errType, message)
}

func newTestProvider(endpoint string, cfg Config) *Provider {
	snsClient := sns.New(sns.Options{
		Region:       "us-east-1",
		BaseEndpoint: aws.String(endpoint),
		Credentials: aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
			return aws.Credentials{
				AccessKeyID:     "test",
				SecretAccessKey: "test",
			}, nil
		}),
	})

	if cfg.SMSType == "" {
		cfg.SMSType = SMSTransactional
	}

	return &Provider{
		client: snsClient,
		config: cfg,
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Region != "us-east-1" {
		t.Errorf("Region = %q, want us-east-1", cfg.Region)
	}
	if cfg.SMSType != SMSTransactional {
		t.Errorf("SMSType = %q, want Transactional", cfg.SMSType)
	}
}

func TestName(t *testing.T) {
	p := &Provider{}
	if p.Name() != "aws-sns" {
		t.Errorf("Name() = %q, want aws-sns", p.Name())
	}
}

func TestGetStatusUnsupported(t *testing.T) {
	p := &Provider{}
	_, err := p.GetStatus(context.Background(), "msg-001")
	if !errors.Is(err, gosms.ErrUnsupported) {
		t.Errorf("error = %v, want ErrUnsupported", err)
	}
}

func TestSendSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		bodyStr := string(body)

		if r.Method != "POST" {
			t.Errorf("method = %q, want POST", r.Method)
		}

		// URL-encoded body: + becomes %2B
		if !strings.Contains(bodyStr, "PhoneNumber=%2B15551234567") {
			t.Errorf("request body missing phone number: %s", bodyStr)
		}
		if !strings.Contains(bodyStr, "Message=Hello+from+SNS") {
			t.Errorf("request body missing message: %s", bodyStr)
		}
		if !strings.Contains(bodyStr, "Transactional") {
			t.Errorf("request body missing SMSType: %s", bodyStr)
		}
		if !strings.Contains(bodyStr, "TestApp") {
			t.Errorf("request body missing SenderID: %s", bodyStr)
		}
		if !strings.Contains(bodyStr, "MaxPrice") {
			t.Errorf("request body missing MaxPrice: %s", bodyStr)
		}

		writePublishResponse(w, "sns-msg-001")
	}))
	defer srv.Close()

	p := newTestProvider(srv.URL, Config{
		SenderID: "TestApp",
		SMSType:  SMSTransactional,
		MaxPrice: "0.50",
	})

	result, err := p.Send(context.Background(), gosms.NewMessage("+15551234567", "Hello from SNS"))
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	if result.MessageID != "sns-msg-001" {
		t.Errorf("MessageID = %q, want sns-msg-001", result.MessageID)
	}
	if result.Status != gosms.StatusAccepted {
		t.Errorf("Status = %q, want accepted", result.Status)
	}
	if result.Provider != "aws-sns" {
		t.Errorf("Provider = %q, want aws-sns", result.Provider)
	}
	if result.To != "+15551234567" {
		t.Errorf("To = %q", result.To)
	}
	if result.Raw["message_id"] != "sns-msg-001" {
		t.Errorf("Raw[message_id] = %v", result.Raw["message_id"])
	}
}

func TestSendWithMessageFrom(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		// Message-level From ("MsgSender") should override config SenderID
		if !strings.Contains(string(body), "MsgSender") {
			t.Errorf("request should contain message-level sender ID: %s", body)
		}
		writePublishResponse(w, "sns-msg-002")
	}))
	defer srv.Close()

	p := newTestProvider(srv.URL, Config{SenderID: "ConfigSender"})
	msg := gosms.NewMessage("+15551234567", "test").WithFrom("MsgSender")

	result, err := p.Send(context.Background(), msg)
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	if result.MessageID != "sns-msg-002" {
		t.Errorf("MessageID = %q", result.MessageID)
	}
}

func TestSendNoSenderID(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if strings.Contains(string(body), "SenderID") {
			t.Errorf("request should not contain SenderID when not configured: %s", body)
		}
		writePublishResponse(w, "sns-msg-003")
	}))
	defer srv.Close()

	p := newTestProvider(srv.URL, Config{})
	_, err := p.Send(context.Background(), gosms.NewMessage("+15551234567", "test"))
	if err != nil {
		t.Fatalf("Send() error = %v", err)
	}
}

func TestSendError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeXMLError(w, 400, "InvalidParameter", "Invalid parameter: PhoneNumber")
	}))
	defer srv.Close()

	p := newTestProvider(srv.URL, Config{})
	_, err := p.Send(context.Background(), gosms.NewMessage("+invalid", "test"))
	if !errors.Is(err, gosms.ErrSendFailed) {
		t.Errorf("error = %v, want ErrSendFailed", err)
	}
}

func TestSendBulk(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		writePublishResponse(w, fmt.Sprintf("sns-bulk-%d", callCount))
	}))
	defer srv.Close()

	p := newTestProvider(srv.URL, Config{})
	msgs := []*gosms.Message{
		gosms.NewMessage("+15551111111", "msg 1"),
		gosms.NewMessage("+15552222222", "msg 2"),
		gosms.NewMessage("+15553333333", "msg 3"),
	}

	results, err := p.SendBulk(context.Background(), msgs)
	if err != nil {
		t.Fatalf("SendBulk() error = %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("len(results) = %d, want 3", len(results))
	}
	if callCount != 3 {
		t.Errorf("callCount = %d, want 3", callCount)
	}
	for i, r := range results {
		if r.Status != gosms.StatusAccepted {
			t.Errorf("results[%d].Status = %q, want accepted", i, r.Status)
		}
	}
}

func TestSendSMSTypeAttribute(t *testing.T) {
	tests := []struct {
		name    string
		smsType SMSType
		want    string
	}{
		{"transactional", SMSTransactional, "Transactional"},
		{"promotional", SMSPromotional, "Promotional"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, _ := io.ReadAll(r.Body)
				if !strings.Contains(string(body), tt.want) {
					t.Errorf("request body should contain SMSType %q: %s", tt.want, body)
				}
				writePublishResponse(w, "test")
			}))
			defer srv.Close()

			p := newTestProvider(srv.URL, Config{SMSType: tt.smsType})
			_, err := p.Send(context.Background(), gosms.NewMessage("+15551234567", "test"))
			if err != nil {
				t.Fatalf("Send() error = %v", err)
			}
		})
	}
}

func TestSetSMSAttributes(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		bodyStr := string(body)
		if !strings.Contains(bodyStr, "MonthlySpendLimit") {
			t.Error("missing MonthlySpendLimit")
		}
		if !strings.Contains(bodyStr, "100.00") {
			t.Error("missing spend limit value")
		}
		if !strings.Contains(bodyStr, "DefaultSenderID") {
			t.Error("missing DefaultSenderID")
		}
		writeEmptyXMLResponse(w, "SetSMSAttributes")
	}))
	defer srv.Close()

	p := newTestProvider(srv.URL, Config{SenderID: "TestApp"})
	err := p.SetSMSAttributes(context.Background(), "100.00", "", "")
	if err != nil {
		t.Fatalf("SetSMSAttributes() error = %v", err)
	}
}

func TestSetSMSAttributesWithDeliveryRole(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		bodyStr := string(body)
		if !strings.Contains(bodyStr, "DeliveryStatusIAMRole") {
			t.Error("missing DeliveryStatusIAMRole")
		}
		if !strings.Contains(bodyStr, "DeliveryStatusSuccessSamplingRate") {
			t.Error("missing DeliveryStatusSuccessSamplingRate")
		}
		writeEmptyXMLResponse(w, "SetSMSAttributes")
	}))
	defer srv.Close()

	p := newTestProvider(srv.URL, Config{})
	err := p.SetSMSAttributes(context.Background(), "", "arn:aws:iam::123:role/SNS", "100")
	if err != nil {
		t.Fatalf("SetSMSAttributes() error = %v", err)
	}
}

func TestGetSMSAttributes(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/xml")
		fmt.Fprint(w, `<GetSMSAttributesResponse>
			<GetSMSAttributesResult>
				<attributes>
					<entry><key>DefaultSMSType</key><value>Transactional</value></entry>
					<entry><key>MonthlySpendLimit</key><value>100.00</value></entry>
				</attributes>
			</GetSMSAttributesResult>
		</GetSMSAttributesResponse>`)
	}))
	defer srv.Close()

	p := newTestProvider(srv.URL, Config{})
	attrs, err := p.GetSMSAttributes(context.Background())
	if err != nil {
		t.Fatalf("GetSMSAttributes() error = %v", err)
	}
	if attrs["DefaultSMSType"] != "Transactional" {
		t.Errorf("DefaultSMSType = %q, want Transactional", attrs["DefaultSMSType"])
	}
	if attrs["MonthlySpendLimit"] != "100.00" {
		t.Errorf("MonthlySpendLimit = %q", attrs["MonthlySpendLimit"])
	}
}

func TestCheckIfPhoneNumberIsOptedOut(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), "15551234567") {
			t.Error("request missing phone number")
		}
		w.Header().Set("Content-Type", "text/xml")
		fmt.Fprint(w, `<CheckIfPhoneNumberIsOptedOutResponse>
			<CheckIfPhoneNumberIsOptedOutResult>
				<isOptedOut>true</isOptedOut>
			</CheckIfPhoneNumberIsOptedOutResult>
		</CheckIfPhoneNumberIsOptedOutResponse>`)
	}))
	defer srv.Close()

	p := newTestProvider(srv.URL, Config{})
	optedOut, err := p.CheckIfPhoneNumberIsOptedOut(context.Background(), "+15551234567")
	if err != nil {
		t.Fatalf("CheckIfPhoneNumberIsOptedOut() error = %v", err)
	}
	if !optedOut {
		t.Error("expected opted out = true")
	}
}

func TestCheckIfPhoneNumberNotOptedOut(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/xml")
		fmt.Fprint(w, `<CheckIfPhoneNumberIsOptedOutResponse>
			<CheckIfPhoneNumberIsOptedOutResult>
				<isOptedOut>false</isOptedOut>
			</CheckIfPhoneNumberIsOptedOutResult>
		</CheckIfPhoneNumberIsOptedOutResponse>`)
	}))
	defer srv.Close()

	p := newTestProvider(srv.URL, Config{})
	optedOut, err := p.CheckIfPhoneNumberIsOptedOut(context.Background(), "+15559999999")
	if err != nil {
		t.Fatalf("error = %v", err)
	}
	if optedOut {
		t.Error("expected opted out = false")
	}
}

func TestListPhoneNumbersOptedOut(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "text/xml")
		if callCount == 1 {
			fmt.Fprint(w, `<ListPhoneNumbersOptedOutResponse>
				<ListPhoneNumbersOptedOutResult>
					<phoneNumbers>
						<member>+15551111111</member>
						<member>+15552222222</member>
					</phoneNumbers>
					<nextToken>page2</nextToken>
				</ListPhoneNumbersOptedOutResult>
			</ListPhoneNumbersOptedOutResponse>`)
		} else {
			fmt.Fprint(w, `<ListPhoneNumbersOptedOutResponse>
				<ListPhoneNumbersOptedOutResult>
					<phoneNumbers>
						<member>+15553333333</member>
					</phoneNumbers>
				</ListPhoneNumbersOptedOutResult>
			</ListPhoneNumbersOptedOutResponse>`)
		}
	}))
	defer srv.Close()

	p := newTestProvider(srv.URL, Config{})
	numbers, err := p.ListPhoneNumbersOptedOut(context.Background())
	if err != nil {
		t.Fatalf("ListPhoneNumbersOptedOut() error = %v", err)
	}
	if len(numbers) != 3 {
		t.Fatalf("len(numbers) = %d, want 3", len(numbers))
	}
	if numbers[0] != "+15551111111" || numbers[2] != "+15553333333" {
		t.Errorf("numbers = %v", numbers)
	}
	if callCount != 2 {
		t.Errorf("callCount = %d, want 2 (pagination)", callCount)
	}
}

func TestOptInPhoneNumber(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), "15551234567") {
			t.Error("request missing phone number")
		}
		writeEmptyXMLResponse(w, "OptInPhoneNumber")
	}))
	defer srv.Close()

	p := newTestProvider(srv.URL, Config{})
	err := p.OptInPhoneNumber(context.Background(), "+15551234567")
	if err != nil {
		t.Fatalf("OptInPhoneNumber() error = %v", err)
	}
}

func TestNewProviderWithCustomClient(t *testing.T) {
	customClient := sns.New(sns.Options{Region: "eu-west-1"})
	p, err := NewProvider(context.Background(), Config{
		Client: customClient,
	})
	if err != nil {
		t.Fatalf("NewProvider() error = %v", err)
	}
	if p.client != customClient {
		t.Error("expected custom client to be used")
	}
}

func TestNewProviderWithStaticCredentials(t *testing.T) {
	p, err := NewProvider(context.Background(), Config{
		Region:          "us-west-2",
		AccessKeyID:     "AKID",
		SecretAccessKey: "SECRET",
	})
	if err != nil {
		t.Fatalf("NewProvider() error = %v", err)
	}
	if p.Name() != "aws-sns" {
		t.Errorf("Name() = %q", p.Name())
	}
}

// Compile-time check that Provider satisfies gosms.Provider.
var _ gosms.Provider = (*Provider)(nil)
