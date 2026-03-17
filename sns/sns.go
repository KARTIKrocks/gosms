// Package sns provides an AWS SNS SMS provider for gosms.
package sns

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sns/types"

	gosms "github.com/KARTIKrocks/gosms"
)

// SMSType represents the type of SMS message.
type SMSType string

const (
	// SMSPromotional is for marketing/promotional messages.
	SMSPromotional SMSType = "Promotional"
	// SMSTransactional is for critical/transactional messages.
	SMSTransactional SMSType = "Transactional"
)

// Config holds AWS SNS-specific configuration.
type Config struct {
	// Region is the AWS region.
	Region string

	// AccessKeyID is the AWS access key ID.
	AccessKeyID string

	// SecretAccessKey is the AWS secret access key.
	SecretAccessKey string

	// SenderID is the default sender ID (11 alphanumeric characters max).
	SenderID string

	// SMSType is the type of SMS (Promotional or Transactional).
	SMSType SMSType

	// MaxPrice is the maximum price per SMS (USD).
	MaxPrice string

	// Client is a custom SNS client (optional).
	Client *sns.Client
}

// DefaultConfig returns a default SNS configuration.
func DefaultConfig() Config {
	return Config{
		Region:  "us-east-1",
		SMSType: SMSTransactional,
	}
}

// Provider implements the gosms.Provider interface for AWS SNS.
type Provider struct {
	client *sns.Client
	config Config
}

// NewProvider creates a new AWS SNS provider.
func NewProvider(ctx context.Context, snsConfig Config) (*Provider, error) {
	var client *sns.Client

	if snsConfig.Client != nil {
		client = snsConfig.Client
	} else {
		var opts []func(*config.LoadOptions) error

		opts = append(opts, config.WithRegion(snsConfig.Region))

		if snsConfig.AccessKeyID != "" && snsConfig.SecretAccessKey != "" {
			opts = append(opts, config.WithCredentialsProvider(
				credentials.NewStaticCredentialsProvider(
					snsConfig.AccessKeyID,
					snsConfig.SecretAccessKey,
					"",
				),
			))
		}

		cfg, err := config.LoadDefaultConfig(ctx, opts...)
		if err != nil {
			return nil, fmt.Errorf("%w: %w", gosms.ErrInvalidConfig, err)
		}

		client = sns.NewFromConfig(cfg)
	}

	return &Provider{
		client: client,
		config: snsConfig,
	}, nil
}

// Name returns the provider name.
func (p *Provider) Name() string {
	return "aws-sns"
}

// Send sends an SMS message via AWS SNS.
func (p *Provider) Send(ctx context.Context, msg *gosms.Message) (*gosms.Result, error) {
	attrs := make(map[string]types.MessageAttributeValue)

	attrs["AWS.SNS.SMS.SMSType"] = types.MessageAttributeValue{
		DataType:    aws.String("String"),
		StringValue: aws.String(string(p.config.SMSType)),
	}

	senderID := msg.From
	if senderID == "" {
		senderID = p.config.SenderID
	}
	if senderID != "" {
		attrs["AWS.SNS.SMS.SenderID"] = types.MessageAttributeValue{
			DataType:    aws.String("String"),
			StringValue: aws.String(senderID),
		}
	}

	if p.config.MaxPrice != "" {
		attrs["AWS.SNS.SMS.MaxPrice"] = types.MessageAttributeValue{
			DataType:    aws.String("Number"),
			StringValue: aws.String(p.config.MaxPrice),
		}
	}

	input := &sns.PublishInput{
		PhoneNumber:       aws.String(msg.To),
		Message:           aws.String(msg.Body),
		MessageAttributes: attrs,
	}

	output, err := p.client.Publish(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", gosms.ErrSendFailed, err)
	}

	return &gosms.Result{
		MessageID: aws.ToString(output.MessageId),
		To:        msg.To,
		Status:    gosms.StatusAccepted,
		Provider:  p.Name(),
		SentAt:    time.Now(),
		Raw: map[string]any{
			"message_id": aws.ToString(output.MessageId),
		},
	}, nil
}

// SendBulk sends multiple SMS messages.
func (p *Provider) SendBulk(ctx context.Context, msgs []*gosms.Message) ([]*gosms.Result, error) {
	return gosms.SendEach(ctx, p.Name(), msgs, p.Send), nil
}

// GetStatus retrieves the delivery status of a message.
// Note: AWS SNS doesn't provide direct status lookup via API.
// Status must be tracked via CloudWatch or delivery status logging.
func (p *Provider) GetStatus(_ context.Context, _ string) (*gosms.Status, error) {
	return nil, fmt.Errorf("%w: SNS requires CloudWatch for delivery status", gosms.ErrUnsupported)
}

// SetSMSAttributes sets the default SMS attributes for the account.
func (p *Provider) SetSMSAttributes(ctx context.Context, monthlySpendLimit, deliveryStatusIAMRole, deliveryStatusSuccessSamplingRate string) error {
	attrs := make(map[string]string)

	if monthlySpendLimit != "" {
		attrs["MonthlySpendLimit"] = monthlySpendLimit
	}

	if deliveryStatusIAMRole != "" {
		attrs["DeliveryStatusIAMRole"] = deliveryStatusIAMRole
		attrs["DeliveryStatusSuccessSamplingRate"] = deliveryStatusSuccessSamplingRate
	}

	if p.config.SenderID != "" {
		attrs["DefaultSenderID"] = p.config.SenderID
	}

	attrs["DefaultSMSType"] = string(p.config.SMSType)

	_, err := p.client.SetSMSAttributes(ctx, &sns.SetSMSAttributesInput{
		Attributes: attrs,
	})

	return err
}

// GetSMSAttributes retrieves the current SMS attributes for the account.
func (p *Provider) GetSMSAttributes(ctx context.Context) (map[string]string, error) {
	output, err := p.client.GetSMSAttributes(ctx, &sns.GetSMSAttributesInput{})
	if err != nil {
		return nil, err
	}

	return output.Attributes, nil
}

// CheckIfPhoneNumberIsOptedOut checks if a phone number has opted out.
func (p *Provider) CheckIfPhoneNumberIsOptedOut(ctx context.Context, phoneNumber string) (bool, error) {
	output, err := p.client.CheckIfPhoneNumberIsOptedOut(ctx, &sns.CheckIfPhoneNumberIsOptedOutInput{
		PhoneNumber: aws.String(phoneNumber),
	})
	if err != nil {
		return false, err
	}

	return output.IsOptedOut, nil
}

// ListPhoneNumbersOptedOut lists phone numbers that have opted out.
func (p *Provider) ListPhoneNumbersOptedOut(ctx context.Context) ([]string, error) {
	var numbers []string
	var nextToken *string

	for {
		output, err := p.client.ListPhoneNumbersOptedOut(ctx, &sns.ListPhoneNumbersOptedOutInput{
			NextToken: nextToken,
		})
		if err != nil {
			return nil, err
		}

		numbers = append(numbers, output.PhoneNumbers...)

		if output.NextToken == nil {
			break
		}
		nextToken = output.NextToken
	}

	return numbers, nil
}

// OptInPhoneNumber opts a phone number back in.
func (p *Provider) OptInPhoneNumber(ctx context.Context, phoneNumber string) error {
	_, err := p.client.OptInPhoneNumber(ctx, &sns.OptInPhoneNumberInput{
		PhoneNumber: aws.String(phoneNumber),
	})
	return err
}
