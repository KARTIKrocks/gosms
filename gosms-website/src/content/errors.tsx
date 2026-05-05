import CodeBlock from '../components/CodeBlock';
import ModuleSection from '../components/ModuleSection';

export default function ErrorsDocs() {
  return (
    <ModuleSection
      id="errors"
      title="Errors & Statuses"
      description="Sentinel errors for common failure modes and a Status enum that tracks delivery progress."
      importPath="github.com/KARTIKrocks/gosms"
      features={[
        'Sentinel errors compatible with errors.Is',
        'Provider-specific failures wrapped under ErrProviderError',
        'Status enum with IsFinal / IsSuccess helpers',
      ]}
    >
      {/* ── Sentinels ── */}
      <h3 id="errors-sentinels" className="text-lg font-semibold text-text-heading mt-8 mb-2">Sentinel Errors</h3>
      <CodeBlock code={`result, err := client.Send(ctx, to, body)
if err != nil {
    switch {
    case errors.Is(err, gosms.ErrInvalidPhone):
        log.Println("Invalid phone number")
    case errors.Is(err, gosms.ErrInvalidMessage):
        log.Println("Invalid message content")
    case errors.Is(err, gosms.ErrRateLimited):
        log.Println("Rate limited, try again later")
    case errors.Is(err, gosms.ErrInsufficientFunds):
        log.Println("Account balance too low")
    case errors.Is(err, gosms.ErrBlacklisted):
        log.Println("Number is blacklisted")
    case errors.Is(err, gosms.ErrProviderError):
        log.Println("Provider error:", err)
    default:
        log.Println("Unknown error:", err)
    }
}`} />

      {/* ── Statuses ── */}
      <h3 id="errors-statuses" className="text-lg font-semibold text-text-heading mt-8 mb-2">Status Values</h3>
      <div className="overflow-x-auto mb-4">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-border text-left">
              <th className="py-2 pr-4 text-text-heading font-semibold">Status</th>
              <th className="py-2 text-text-heading font-semibold">Description</th>
            </tr>
          </thead>
          <tbody>
            <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">StatusPending</td><td className="py-2 text-text-muted">Message is pending</td></tr>
            <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">StatusQueued</td><td className="py-2 text-text-muted">Message is queued for delivery</td></tr>
            <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">StatusAccepted</td><td className="py-2 text-text-muted">Message accepted by provider</td></tr>
            <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">StatusSent</td><td className="py-2 text-text-muted">Message sent to carrier</td></tr>
            <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">StatusDelivered</td><td className="py-2 text-text-muted">Message delivered to recipient</td></tr>
            <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">StatusFailed</td><td className="py-2 text-text-muted">Delivery failed</td></tr>
            <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">StatusRejected</td><td className="py-2 text-text-muted">Message was rejected</td></tr>
            <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">StatusExpired</td><td className="py-2 text-text-muted">Message expired before delivery</td></tr>
            <tr className="border-b border-border/50"><td className="py-2 pr-4 font-mono text-accent whitespace-nowrap">StatusUnknown</td><td className="py-2 text-text-muted">Status unknown</td></tr>
          </tbody>
        </table>
      </div>
    </ModuleSection>
  );
}
