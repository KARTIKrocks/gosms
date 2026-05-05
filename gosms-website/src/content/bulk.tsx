import CodeBlock from '../components/CodeBlock';
import ModuleSection from '../components/ModuleSection';

export default function BulkDocs() {
  return (
    <ModuleSection
      id="bulk"
      title="Bulk Messaging"
      description="Send many messages at once with Batch (different bodies per recipient) or SendToMany (one body, many recipients)."
      importPath="github.com/KARTIKrocks/gosms"
      features={[
        'Batch.Send fans out individual Messages and returns per-recipient Results',
        'SendToMany sends the same body to many recipients in one call',
        'Failures on individual messages do not abort the batch — inspect Result.Error',
      ]}
    >
      {/* ── Batch ── */}
      <h3 id="bulk-batch" className="text-lg font-semibold text-text-heading mt-8 mb-2">Batch</h3>
      <p className="text-text-muted mb-3">
        Use <code className="text-accent">Batch</code> when each recipient gets a different body or different per-message options.
      </p>
      <CodeBlock code={`batch := gosms.NewBatch()
batch.AddNew("+15551111111", "Message 1")
batch.AddNew("+15552222222", "Message 2")
batch.AddNew("+15553333333", "Message 3")

results, err := batch.Send(ctx, client)
for _, result := range results {
    if result.Success() {
        log.Printf("Sent to %s: %s", result.To, result.MessageID)
    } else {
        log.Printf("Failed to %s: %s", result.To, result.Error)
    }
}`} />

      {/* ── SendToMany ── */}
      <h3 id="bulk-many" className="text-lg font-semibold text-text-heading mt-8 mb-2">SendToMany</h3>
      <p className="text-text-muted mb-3">
        Use <code className="text-accent">SendToMany</code> when the body is identical across recipients (broadcasts, marketing blasts).
      </p>
      <CodeBlock code={`results, err := gosms.SendToMany(ctx, client,
    "Flash sale! 50% off today only!",
    "+15551111111",
    "+15552222222",
    "+15553333333",
)`} />
    </ModuleSection>
  );
}
