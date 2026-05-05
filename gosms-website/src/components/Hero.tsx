import { useState } from 'react';
import { useVersion } from '../hooks/useVersion';

interface Feature {
  title: string;
  desc: string;
}

const features: Feature[] = [
  { title: 'Unified Provider API', desc: 'One Provider interface across Twilio, AWS SNS, Vonage, and MSG91' },
  { title: 'Multi-Module', desc: 'Each provider is a separate Go module — only download what you use' },
  { title: 'Bulk Messaging', desc: 'Batch and SendToMany helpers for high-throughput sending' },
  { title: 'Multi-Provider', desc: 'Fallback and round-robin strategies across providers' },
  { title: 'OTP Support', desc: 'MSG91 OTPProvider for full send / verify / resend flow' },
  { title: 'Webhook Parsing', desc: 'Built-in delivery status webhook parsers for Twilio and Vonage' },
  { title: 'Phone Validation', desc: 'E.164 validation, normalization, and GSM 03.38 segment calculation' },
  { title: 'Mock Provider', desc: 'In-memory mock for tests, with configurable failures and assertions' },
  { title: 'Message Templates', desc: 'Pre-built OTP, alert, and notification message helpers' },
  { title: 'Thread Safe', desc: 'All providers and the client are safe for concurrent use' },
];

export default function Hero() {
  const [copied, setCopied] = useState(false);
  const { selectedVersion, getInstallCmd } = useVersion();
  const installCmd = getInstallCmd(selectedVersion);

  const handleCopy = () => {
    navigator.clipboard.writeText(installCmd);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <section id="top" className="py-16 border-b border-border">
      <h1 className="text-4xl md:text-5xl font-bold text-text-heading mb-4">
        Unified Go SMS library
      </h1>
      <p className="text-lg text-text-muted max-w-2xl mb-8">
        Send SMS in Go through one unified API. Twilio, AWS SNS, Vonage, MSG91 —
        bulk sending, fallback strategies, webhooks, OTP, and mock testing built in.
      </p>

      <div className="flex items-center gap-2 bg-bg-card border border-border rounded-lg px-4 py-3 max-w-lg mb-10">
        <span className="text-text-muted select-none">$</span>
        <code className="flex-1 text-sm font-mono text-accent">{installCmd}</code>
        <button
          onClick={handleCopy}
          className="text-xs text-text-muted hover:text-text px-2 py-1 rounded bg-overlay hover:bg-overlay-hover transition-colors"
        >
          {copied ? 'Copied!' : 'Copy'}
        </button>
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
        {features.map((f) => (
          <div key={f.title} className="bg-bg-card border border-border rounded-lg p-4">
            <h3 className="text-sm font-semibold text-text-heading mb-1">{f.title}</h3>
            <p className="text-xs text-text-muted">{f.desc}</p>
          </div>
        ))}
      </div>
    </section>
  );
}
