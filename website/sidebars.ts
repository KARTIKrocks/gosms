import type { SidebarsConfig } from '@docusaurus/plugin-content-docs';

// This runs in Node.js - Don't use client-side code here (browser APIs, JSX...)

/**
 * Mirrors the shape of the package itself: the core contract first
 * (client -> message -> result/status), then one page per provider module,
 * then the composition layer (multi-provider, bulk), then operational
 * concerns (webhooks, helpers, testing, errors).
 */
const sidebars: SidebarsConfig = {
  docsSidebar: [
    'intro',
    'getting-started',
    {
      type: 'category',
      label: 'Core',
      collapsed: false,
      items: ['core'],
    },
    {
      type: 'category',
      label: 'Providers',
      collapsed: false,
      items: ['providers', 'twilio', 'sns', 'vonage', 'msg91'],
    },
    {
      type: 'category',
      label: 'Advanced',
      collapsed: false,
      items: ['multi-provider', 'bulk', 'webhooks', 'otp'],
    },
    {
      type: 'category',
      label: 'Reference',
      collapsed: false,
      items: ['helpers', 'testing', 'errors'],
    },
  ],
};

export default sidebars;
