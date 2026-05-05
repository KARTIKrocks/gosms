import CodeBlock from '../components/CodeBlock';
import ModuleSection from '../components/ModuleSection';

export default function CoreDocs() {
  return (
    <ModuleSection
      id="core"
      title="Core API"
      description="The unified Provider interface, Client wrapper, Message builder, and Result type that all providers share."
      importPath="github.com/KARTIKrocks/gosms"
      features={[
        'Single Provider interface backed by every supported gateway',
        'Fluent Message builder with metadata, scheduling, and validity',
        'Result type with ID, status, raw response, and error',
        'Context-aware Send / SendMessage / SendBulk methods',
      ]}
    >
      {/* ── Provider Interface ── */}
      <h3 id="core-provider" className="text-lg font-semibold text-text-heading mt-8 mb-2">Provider Interface</h3>
      <p className="text-text-muted mb-3">
        Every backend (Twilio, SNS, Vonage, MSG91, Mock) implements the same minimal interface.
        Anything that satisfies it is a drop-in provider:
      </p>
      <CodeBlock code={`type Provider interface {
    Send(ctx context.Context, msg *Message) (*Result, error)
    SendBulk(ctx context.Context, msgs []*Message) ([]*Result, error)
    GetStatus(ctx context.Context, messageID string) (*StatusInfo, error)
    Name() string
}

// Optional capability — provider may also implement OTPProvider
type OTPProvider interface {
    SendOTP(ctx context.Context, req *OTPRequest) (*Result, error)
    VerifyOTP(ctx context.Context, phone, code string) (*VerifyResult, error)
    ResendOTP(ctx context.Context, phone, retryType string) error
}`} />

      {/* ── Client ── */}
      <h3 id="core-client" className="text-lg font-semibold text-text-heading mt-8 mb-2">Client</h3>
      <p className="text-text-muted mb-3">
        <code className="text-accent">Client</code> wraps any <code className="text-accent">Provider</code> and
        adds convenience methods. It is safe for concurrent use after construction.
      </p>
      <div className="overflow-x-auto mb-4">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-border text-left">
              <th className="py-2 pr-4 text-text-heading font-semibold">Method</th>
              <th className="py-2 text-text-heading font-semibold">Description</th>
            </tr>
          </thead>
          <tbody>
            <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">Send(ctx, to, body)</td><td className="py-2 text-text-muted">Send a plain text message — shortcut over SendMessage</td></tr>
            <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">SendMessage(ctx, msg)</td><td className="py-2 text-text-muted">Send a fully-built Message (with overrides, schedule, metadata)</td></tr>
            <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">SendBulk(ctx, msgs)</td><td className="py-2 text-text-muted">Send a slice of Messages — provider may chunk internally</td></tr>
            <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">GetStatus(ctx, id)</td><td className="py-2 text-text-muted">Look up delivery status by provider message ID</td></tr>
            <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">Provider()</td><td className="py-2 text-text-muted">Access the underlying Provider (for type assertions like OTPProvider)</td></tr>
          </tbody>
        </table>
      </div>

      {/* ── Message Builder ── */}
      <h3 id="core-message" className="text-lg font-semibold text-text-heading mt-8 mb-2">Message Builder</h3>
      <p className="text-text-muted mb-3">
        Build a message with the fluent API. All <code className="text-accent">With*</code> methods return
        <code className="text-accent"> *Message</code> for chaining.
      </p>
      <CodeBlock code={`msg := gosms.NewMessage("+15559876543", "Hello!").
    WithFrom("+15551234567").
    WithReference("order-123").
    WithValidity(1 * time.Hour).
    WithMetadata("user_id", "12345")

result, err := client.SendMessage(ctx, msg)`} />

      <h4 className="text-base font-semibold text-text-heading mt-6 mb-2">Scheduled Messages (Twilio)</h4>
      <CodeBlock code={`msg := gosms.NewMessage("+15559876543", "Reminder: Your appointment is tomorrow").
    WithSchedule(time.Now().Add(24 * time.Hour))

result, err := client.SendMessage(ctx, msg)`} />

      {/* ── Result & Status ── */}
      <h3 id="core-result" className="text-lg font-semibold text-text-heading mt-8 mb-2">Result & Status</h3>
      <p className="text-text-muted mb-3">
        Every send returns a <code className="text-accent">*Result</code>. Inspect it for delivery state,
        error details, or raw provider data.
      </p>
      <CodeBlock code={`type Result struct {
    MessageID string         // Provider-assigned ID
    To        string         // Recipient (E.164)
    Status    Status         // StatusPending / Sent / Delivered / Failed / ...
    Error     error          // Set on per-message failure (bulk)
    Raw       map[string]any // Raw provider response
    SentAt    time.Time
}

if result.Success() {
    log.Printf("Sent to %s: %s", result.To, result.MessageID)
}
if result.Status.IsFinal() {
    // No further status updates expected
}`} />
    </ModuleSection>
  );
}
