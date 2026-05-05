import CodeBlock from '../components/CodeBlock';
import ModuleSection from '../components/ModuleSection';

export default function TestingDocs() {
  return (
    <ModuleSection
      id="testing"
      title="Testing with MockProvider"
      description="MockProvider implements Provider in-memory — assert on what was sent, simulate failures, and inject errors."
      importPath="github.com/KARTIKrocks/gosms"
      features={[
        'Records every Send call for later inspection',
        'WithFailAll forces every send to return StatusFailed',
        'WithSendError makes every send return a typed error (e.g. ErrRateLimited)',
        'Reset clears recorded state between test cases',
      ]}
    >
      {/* ── Mock Provider ── */}
      <h3 id="testing-mock" className="text-lg font-semibold text-text-heading mt-8 mb-2">Mock Provider</h3>
      <CodeBlock code={`mock := gosms.NewMockProvider()
client := gosms.NewClient(mock)

// Send message
result, err := client.Send(ctx, "+15551234567", "Test message")

// Verify
if mock.MessageCount() != 1 {
    t.Error("Expected 1 message")
}

lastMsg := mock.LastMessage()
if lastMsg.Message.Body != "Test message" {
    t.Error("Message body mismatch")
}`} />

      {/* ── Assertions & Errors ── */}
      <h3 id="testing-assertions" className="text-lg font-semibold text-text-heading mt-8 mb-2">Simulating Failures</h3>
      <p className="text-text-muted mb-3">
        Mock can simulate two kinds of failures: a successful API call that reports
        <code className="text-accent"> StatusFailed</code> on the result, or an outright Go error from <code className="text-accent">Send</code>.
      </p>
      <CodeBlock code={`// Simulate provider-side failure (no Go error, but Result.Status == Failed)
mock.WithFailAll(true)
result, err := client.Send(ctx, "+15551234567", "This will fail")
// result.Status == gosms.StatusFailed, err == nil

// Simulate transport / API errors (returns from Send)
mock.WithSendError(gosms.ErrRateLimited)
_, err = client.Send(ctx, "+15551234567", "This will error")
// errors.Is(err, gosms.ErrRateLimited) == true

// Reset mock between test cases
mock.Reset()`} />
    </ModuleSection>
  );
}
