import { useState } from 'react';
import ThemeProvider from './components/ThemeProvider';
import VersionProvider from './components/VersionProvider';
import Navbar from './components/Navbar';
import Sidebar from './components/Sidebar';
import VersionBanner from './components/VersionBanner';
import Hero from './components/Hero';
import GettingStarted from './content/getting-started';
import CoreDocs from './content/core';
import ProvidersDocs from './content/providers';
import BulkDocs from './content/bulk';
import MultiProviderDocs from './content/multi-provider';
import WebhooksDocs from './content/webhooks';
import OTPDocs from './content/otp';
import HelpersDocs from './content/helpers';
import TestingDocs from './content/testing';
import ErrorsDocs from './content/errors';
import ExamplesDocs from './content/examples';

export default function App() {
  const [menuOpen, setMenuOpen] = useState(false);

  return (
    <ThemeProvider>
    <VersionProvider>
      <div className="min-h-screen">
        <Navbar onMenuToggle={() => setMenuOpen((o) => !o)} menuOpen={menuOpen} />
        <Sidebar open={menuOpen} onClose={() => setMenuOpen(false)} />

        <main className="pt-16 md:pl-64">
          <div className="max-w-4xl mx-auto px-4 md:px-8 pb-20">
            <div className="pt-4">
              <VersionBanner />
            </div>
            <Hero />
            <GettingStarted />
            <CoreDocs />
            <ProvidersDocs />
            <BulkDocs />
            <MultiProviderDocs />
            <WebhooksDocs />
            <OTPDocs />
            <HelpersDocs />
            <TestingDocs />
            <ErrorsDocs />
            <ExamplesDocs />

            <footer className="py-10 text-center text-sm text-text-muted border-t border-border mt-10">
              <p>
                gosms is open source under the{' '}
                <a
                  href="https://github.com/KARTIKrocks/gosms/blob/main/LICENSE"
                  className="text-primary hover:underline"
                  target="_blank"
                  rel="noopener noreferrer"
                >
                  MIT License
                </a>
              </p>
            </footer>
          </div>
        </main>
      </div>
    </VersionProvider>
    </ThemeProvider>
  );
}
