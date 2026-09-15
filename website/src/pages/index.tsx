import Link from '@docusaurus/Link';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import Layout from '@theme/Layout';
import type { ReactNode } from 'react';

import styles from './index.module.css';

type Feature = {
  readonly title: string;
  readonly description: string;
};

type Capability = {
  readonly capability: string;
};

type Provider = {
  readonly name: string;
  readonly note: string;
};

const FEATURES = [
  {
    title: 'Unified Provider Interface',
    description:
      'One Send / SendBulk / GetStatus contract across every backend',
  },
  {
    title: 'Multi-Module',
    description:
      'Each provider is its own Go module — pull in only the dependencies you use',
  },
  {
    title: 'Zero Deps in Core',
    description: 'The root module has no third-party dependencies at all',
  },
  {
    title: 'Message Builder',
    description:
      'Fluent API for scheduling, validity, references, and metadata',
  },
  {
    title: 'Multi-Provider',
    description: 'Fallback and round-robin strategies across multiple backends',
  },
  {
    title: 'Delivery Tracking',
    description:
      'Status polling plus webhook parsing for Twilio, Vonage, and MSG91',
  },
  {
    title: 'Phone Utilities',
    description:
      'E.164 validation, normalization, and GSM/UCS-2 segment counting',
  },
  {
    title: 'OTP Flow',
    description:
      'Send / verify / resend via the optional OTPProvider capability',
  },
  {
    title: 'Mock Provider',
    description: 'In-memory provider with failure injection for tests',
  },
] as const satisfies readonly Feature[];

// Everything gosms provides that hand-rolling each provider's SDK leaves to
// you. Kept in sync with the "Why gosms?" table in the repository README.
const CAPABILITIES = [
  { capability: 'One interface across Twilio, SNS, Vonage, and MSG91' },
  { capability: 'Only pull the dependencies for providers you use' },
  { capability: 'Delivery status polling + webhook parsing' },
  { capability: 'Fallback and round-robin across providers' },
  { capability: 'E.164 validation, normalization, segment counting' },
  { capability: 'DLT template variables and OTP flow (MSG91)' },
  { capability: 'Mock provider with failure/latency injection' },
] as const satisfies readonly Capability[];

const PROVIDERS = [
  { name: 'Twilio', note: 'Global SMS, scheduled sends, webhook status' },
  {
    name: 'AWS SNS',
    note: 'Transactional/promotional SMS, opt-out management',
  },
  { name: 'Vonage', note: 'Global SMS with webhook delivery receipts' },
  { name: 'MSG91', note: 'India DLT Flow templates, server-side OTP' },
] as const satisfies readonly Provider[];

const INSTALL_COMMAND = 'go get github.com/KARTIKrocks/gosms';

function Hero(): ReactNode {
  return (
    <header className={styles.hero}>
      <div className="container">
        <h1 className={styles.title}>Unified SMS sending for Go</h1>
        <p className={styles.subtitle}>
          One <code>Provider</code> interface for Twilio, AWS SNS, Vonage, and
          MSG91 — each behind its own Go module, so you only pull in the
          dependencies for the providers you actually use.
        </p>

        <div className={styles.buttons}>
          <Link
            className="button button--primary button--lg"
            to="/docs/getting-started">
            Get Started
          </Link>
          <Link
            className="button button--secondary button--lg"
            to="https://pkg.go.dev/github.com/KARTIKrocks/gosms">
            API Reference
          </Link>
        </div>

        <div className={styles.install}>
          <span className={styles.prompt} aria-hidden="true">
            $
          </span>
          <code>{INSTALL_COMMAND}</code>
        </div>
      </div>
    </header>
  );
}

function Features(): ReactNode {
  return (
    <section className="container" aria-label="Features">
      <div className={styles.features}>
        {FEATURES.map((feature) => (
          <article key={feature.title} className={styles.card}>
            <h2>{feature.title}</h2>
            <p>{feature.description}</p>
          </article>
        ))}
      </div>
    </section>
  );
}

function WhyGosms(): ReactNode {
  return (
    <section className={styles.section}>
      <div className="container">
        <h2 className={styles.sectionTitle}>Why gosms?</h2>
        <p className={styles.sectionLead}>
          Every provider ships its own SDK with its own request shape, its own
          error strings, and its own webhook format. gosms gives you one
          interface, one error set, and one webhook-parsing convention — and
          keeps each provider in its own module so switching (or supporting more
          than one) doesn't drag in dependencies you don't need.
        </p>

        <div className={styles.tableScroll}>
          <table className={styles.compare}>
            <thead>
              <tr>
                <th scope="col">Capability</th>
                <th scope="col">gosms</th>
                <th scope="col">Provider SDK directly</th>
              </tr>
            </thead>
            <tbody>
              {CAPABILITIES.map(({ capability }) => (
                <tr key={capability}>
                  <th scope="row">{capability}</th>
                  <td>
                    <span className={styles.check} aria-hidden="true">
                      ✓
                    </span>
                    <span className={styles.srOnly}>Included</span>
                  </td>
                  <td className={styles.diy}>You build it</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </section>
  );
}

function Providers(): ReactNode {
  return (
    <section className={styles.section}>
      <div className="container">
        <h2 className={styles.sectionTitle}>Supported providers</h2>
        <p className={styles.sectionLead}>
          Install only the ones you need —{' '}
          <code>go get github.com/KARTIKrocks/gosms/&#123;provider&#125;</code>.
        </p>

        <div className={styles.providers}>
          {PROVIDERS.map((provider) => (
            <article key={provider.name} className={styles.provider}>
              <h3 className={styles.providerLabel}>{provider.name}</h3>
              <p className={styles.providerNote}>{provider.note}</p>
            </article>
          ))}
        </div>
      </div>
    </section>
  );
}

export default function Home(): ReactNode {
  const { siteConfig } = useDocusaurusContext();

  return (
    <Layout
      title={siteConfig.tagline}
      description="A unified SMS-sending library for Go supporting Twilio, AWS SNS, Vonage, and MSG91, each as its own Go module.">
      <Hero />
      <main>
        <Features />
        <WhyGosms />
        <Providers />
      </main>
    </Layout>
  );
}
