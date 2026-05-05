import CodeBlock from '../components/CodeBlock';
import ModuleSection from '../components/ModuleSection';

export default function ProvidersDocs() {
  return (
    <ModuleSection
      id="providers"
      title="Providers"
      description="Each provider lives in its own Go module so you only pull in the SDK you actually use."
      importPath="github.com/KARTIKrocks/gosms/{provider}"
      features={[
        'Twilio — global SMS via REST API',
        'AWS SNS — Amazon SNS publish for SMS',
        'Vonage — Vonage Messages API',
        'MSG91 — India SMS gateway with DLT Flow templates and OTP',
      ]}
    >
      {/* ── Twilio ── */}
      <h3 id="providers-twilio" className="text-lg font-semibold text-text-heading mt-8 mb-2">Twilio</h3>
      <p className="text-text-muted mb-3">
        Twilio uses Account SID + Auth Token. Set the <code className="text-accent">From</code> number once
        in <code className="text-accent">Config</code>, override per message with <code className="text-accent">msg.WithFrom</code>.
      </p>
      <CodeBlock code={`import (
    "github.com/KARTIKrocks/gosms"
    "github.com/KARTIKrocks/gosms/twilio"
)

provider, err := twilio.NewProvider(twilio.Config{
    AccountSID: "account_sid",
    AuthToken:  "auth_token",
    From:       "+15551234567",
})
if err != nil {
    log.Fatal(err)
}

client := gosms.NewClient(provider)

result, err := client.Send(ctx, "+15559876543", "Hello from gosms!")
if err != nil {
    log.Fatal(err)
}
log.Printf("Message sent: %s, Status: %s", result.MessageID, result.Status)`} />

      {/* ── AWS SNS ── */}
      <h3 id="providers-sns" className="text-lg font-semibold text-text-heading mt-8 mb-2">AWS SNS</h3>
      <p className="text-text-muted mb-3">
        AWS SNS supports transactional and promotional SMS, sender IDs, monthly spend limits, and opt-out tracking.
      </p>
      <CodeBlock code={`import (
    "github.com/KARTIKrocks/gosms"
    "github.com/KARTIKrocks/gosms/sns"
)

config := sns.DefaultConfig()
config.Region = "us-east-1"
config.AccessKeyID = "access_key"
config.SecretAccessKey = "secret_key"
config.SenderID = "MyApp"
config.SMSType = sns.SMSTransactional

provider, err := sns.NewProvider(ctx, config)
if err != nil {
    log.Fatal(err)
}

client := gosms.NewClient(provider)
result, err := client.Send(ctx, "+15559876543", "Your code is 123456")`} />

      <h4 className="text-base font-semibold text-text-heading mt-6 mb-2">SNS Account-Level Operations</h4>
      <CodeBlock code={`import "github.com/KARTIKrocks/gosms/sns"

provider, _ := sns.NewProvider(ctx, config)

// Set account-level SMS attributes
err := provider.SetSMSAttributes(ctx,
    "100.00",                          // Monthly spend limit (USD)
    "arn:aws:iam::123:role/SNSRole",   // IAM role for delivery logs
    "100",                             // Success sampling rate %
)

// Check opt-out status
optedOut, err := provider.CheckIfPhoneNumberIsOptedOut(ctx, "+15551234567")

// List opted-out numbers
numbers, err := provider.ListPhoneNumbersOptedOut(ctx)

// Opt a number back in
err = provider.OptInPhoneNumber(ctx, "+15551234567")`} />

      {/* ── Vonage ── */}
      <h3 id="providers-vonage" className="text-lg font-semibold text-text-heading mt-8 mb-2">Vonage</h3>
      <p className="text-text-muted mb-3">
        Vonage uses an API key + secret. <code className="text-accent">From</code> can be a phone number or alphanumeric sender ID.
      </p>
      <CodeBlock code={`import (
    "github.com/KARTIKrocks/gosms"
    "github.com/KARTIKrocks/gosms/vonage"
)

provider, err := vonage.NewProvider(vonage.Config{
    APIKey:    "api_key",
    APISecret: "api_secret",
    From:      "MyApp",
})
if err != nil {
    log.Fatal(err)
}

client := gosms.NewClient(provider)
result, err := client.Send(ctx, "+15559876543", "Hello from Vonage!")`} />

      {/* ── MSG91 ── */}
      <h3 id="providers-msg91" className="text-lg font-semibold text-text-heading mt-8 mb-2">MSG91</h3>
      <p className="text-text-muted mb-3">
        MSG91 is the standard SMS gateway for India. It uses <strong>DLT-approved Flow templates</strong> —
        variables are passed via <code className="text-accent">msg91.SetVar</code>, not the
        <code className="text-accent"> Body</code> field.
      </p>
      <CodeBlock code={`import (
    "github.com/KARTIKrocks/gosms"
    "github.com/KARTIKrocks/gosms/msg91"
)

provider, err := msg91.NewProvider(msg91.Config{
    AuthKey:    "your_authkey",
    SenderID:   "SENDER",         // 6-char DLT sender ID
    TemplateID: "tmpl_xxx",       // DLT-approved Flow template
    Route:      msg91.RouteTransactional,
})
if err != nil {
    log.Fatal(err)
}

client := gosms.NewClient(provider)

msg := gosms.NewMessage("+919876543210", "")
msg91.SetVar(msg, "name", "Kartik")
msg91.SetVar(msg, "otp", "1234")

result, err := client.SendMessage(ctx, msg)`} />

      <h4 className="text-base font-semibold text-text-heading mt-6 mb-2">Template variables and Body fallback</h4>
      <p className="text-text-muted mb-3">
        MSG91 Flow templates reference placeholders like <code className="text-accent">##name##</code> or
        <code className="text-accent"> ##otp##</code>. Set each one with <code className="text-accent">msg91.SetVar</code>.
        For templates with a single <code className="text-accent">##body##</code> placeholder, any non-empty
        <code className="text-accent"> Message.Body</code> is automatically passed as <code className="text-accent">body</code> when
        no vars are set — so the unified <code className="text-accent">client.Send(ctx, to, text)</code> path works without
        extra wiring.
      </p>

      <h4 className="text-base font-semibold text-text-heading mt-6 mb-2">Per-message overrides</h4>
      <p className="text-text-muted mb-3">
        Use <code className="text-accent">msg91.SetTemplateID(msg, "tmpl_other")</code> to override
        <code className="text-accent"> Config.TemplateID</code> for a single message, or
        <code className="text-accent"> msg.WithFrom("OTHER")</code> to override the sender ID.
      </p>

      <h4 className="text-base font-semibold text-text-heading mt-6 mb-2">Phone normalization</h4>
      <p className="text-text-muted mb-3">
        Non-E.164 numbers up to 10 digits are prefixed with <code className="text-accent">Config.Country</code> (default <code className="text-accent">91</code>).
        <code className="text-accent"> +919876543210</code>, <code className="text-accent">919876543210</code>, and
        <code className="text-accent"> 9876543210</code> all normalize identically.
      </p>

      <h4 className="text-base font-semibold text-text-heading mt-6 mb-2">Bulk grouping</h4>
      <p className="text-text-muted mb-3">
        <code className="text-accent">SendBulk</code> groups recipients by effective
        <code className="text-accent"> (template_id, sender)</code> and sends each group as one Flow API call.
        Groups larger than <code className="text-accent">Config.MaxRecipientsPerCall</code> (default 1000) are
        automatically chunked across multiple calls; set a negative value to disable chunking.
      </p>
    </ModuleSection>
  );
}
