import CodeBlock from '../components/CodeBlock';
import ModuleSection from '../components/ModuleSection';

export default function MultiProviderDocs() {
  return (
    <ModuleSection
      id="multi-provider"
      title="Multi-Provider"
      description="MultiProvider wraps several providers behind one Provider interface — pick a strategy for how they're combined."
      importPath="github.com/KARTIKrocks/gosms"
      features={[
        'Fallback strategy — try the first provider, fall back to the next on error',
        'Round-robin strategy — rotate evenly across providers',
        'Atomic round-robin counter — safe under high concurrency',
        'Drop-in replacement: hand it to NewClient like any other Provider',
      ]}
    >
      {/* ── Fallback ── */}
      <h3 id="multi-fallback" className="text-lg font-semibold text-text-heading mt-8 mb-2">Fallback</h3>
      <p className="text-text-muted mb-3">
        Try Twilio first; if it errors, automatically try Vonage:
      </p>
      <CodeBlock code={`multi := gosms.NewMultiProvider(twilioProvider, vonageProvider)

client := gosms.NewClient(multi)
result, err := client.Send(ctx, to, body)`} />

      {/* ── Round-Robin ── */}
      <h3 id="multi-roundrobin" className="text-lg font-semibold text-text-heading mt-8 mb-2">Round-Robin</h3>
      <p className="text-text-muted mb-3">
        Rotate sends evenly across providers — useful for spreading load or staying under per-provider rate limits:
      </p>
      <CodeBlock code={`multi := gosms.NewMultiProvider(twilioProvider, vonageProvider).
    WithStrategy(gosms.StrategyRoundRobin)

client := gosms.NewClient(multi)`} />
    </ModuleSection>
  );
}
