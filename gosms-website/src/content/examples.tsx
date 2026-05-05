import { useVersion } from '../hooks/useVersion';

interface Example {
  name: string;
  path: string;
  description: string;
}

const examples: Example[] = [
  { name: 'basic', path: 'examples/basic', description: 'Core API usage with mock provider' },
  { name: 'twilio-provider', path: 'examples/twilio-provider', description: 'Sending via Twilio' },
  { name: 'sns-provider', path: 'examples/sns-provider', description: 'Sending via AWS SNS' },
  { name: 'vonage-provider', path: 'examples/vonage-provider', description: 'Sending via Vonage' },
  { name: 'msg91-provider', path: 'examples/msg91-provider', description: 'Sending via MSG91 (Flow templates + OTP)' },
  { name: 'multi-provider', path: 'examples/multi-provider', description: 'Fallback and round-robin strategies' },
  { name: 'webhooks', path: 'examples/webhooks', description: 'Delivery status webhook server' },
  { name: 'mock-testing', path: 'examples/mock-testing', description: 'Using MockProvider in tests' },
  { name: 'helpers', path: 'examples/helpers', description: 'Phone validation, normalization, segment calculation' },
];

export default function ExamplesDocs() {
  const { selectedVersion, isLatest } = useVersion();
  const ref = isLatest ? 'main' : selectedVersion;
  const repoRoot = `https://github.com/KARTIKrocks/gosms/tree/${ref}`;

  return (
    <section id="examples" className="py-10 border-b border-border last:border-b-0">
      <h2 className="text-2xl font-bold text-text-heading mb-2">Examples</h2>
      <p className="text-text-muted mb-6">
        Runnable examples in the <code className="text-accent">examples/</code> directory.
        Most run with no credentials thanks to the mock provider.
      </p>

      <div className="overflow-x-auto mb-6">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-border text-left">
              <th className="py-2 pr-4 text-text-heading font-semibold">Example</th>
              <th className="py-2 text-text-heading font-semibold">Description</th>
            </tr>
          </thead>
          <tbody>
            {examples.map((ex) => (
              <tr key={ex.name} className="border-b border-border/50">
                <td className="py-2 pr-4 font-mono whitespace-nowrap">
                  <a
                    href={`${repoRoot}/${ex.path}`}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="text-accent hover:underline"
                  >
                    {ex.name}
                  </a>
                </td>
                <td className="py-2 text-text-muted">{ex.description}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      <pre className="bg-bg-card rounded-lg p-4 text-sm overflow-x-auto border border-border">
        <code className="text-text-muted"># Run an example (no credentials needed){'\n'}cd examples/basic && go run .</code>
      </pre>
    </section>
  );
}
