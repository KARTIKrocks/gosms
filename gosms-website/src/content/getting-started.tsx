import CodeBlock from '../components/CodeBlock';
import { useVersion } from '../hooks/useVersion';

export default function GettingStarted() {
  const { selectedVersion, getInstallCmd, isLatest } = useVersion();
  const coreCmd = getInstallCmd(selectedVersion);
  const suffix = isLatest ? '' : `@${selectedVersion}`;

  return (
    <section id="getting-started" className="py-10 border-b border-border">
      <h2 className="text-2xl font-bold text-text-heading mb-4">Getting Started</h2>

      <h3 className="text-lg font-semibold text-text-heading mt-6 mb-2">Installation</h3>
      <p className="text-text-muted mb-3">
        Each provider is a <strong>separate Go module</strong> — install only what you need.
      </p>
      <CodeBlock lang="bash" code={`# Core (required)
${coreCmd}

# Providers (pick one or more)
go get github.com/KARTIKrocks/gosms/twilio${suffix}
go get github.com/KARTIKrocks/gosms/sns${suffix}
go get github.com/KARTIKrocks/gosms/vonage${suffix}
go get github.com/KARTIKrocks/gosms/msg91${suffix}`} />

      <h3 className="text-lg font-semibold text-text-heading mt-8 mb-2">Quick Start</h3>
      <p className="text-text-muted mb-3">
        Send your first SMS via Twilio. The same <code className="text-accent">Client.Send</code> works for every provider:
      </p>
      <CodeBlock code={`package main

import (
    "context"
    "log"

    "github.com/KARTIKrocks/gosms"
    "github.com/KARTIKrocks/gosms/twilio"
)

func main() {
    provider, err := twilio.NewProvider(twilio.Config{
        AccountSID: "account_sid",
        AuthToken:  "auth_token",
        From:       "+15551234567",
    })
    if err != nil {
        log.Fatal(err)
    }

    client := gosms.NewClient(provider)

    result, err := client.Send(context.Background(), "+15559876543", "Hello from gosms!")
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Message sent: %s, Status: %s", result.MessageID, result.Status)
}`} />
    </section>
  );
}
