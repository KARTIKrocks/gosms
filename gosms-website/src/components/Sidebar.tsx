import { useEffect, useRef, useState, useMemo } from 'react';

interface SidebarProps {
  open: boolean;
  onClose: () => void;
}

interface ChildItem {
  id: string;
  label: string;
}

interface SectionItem {
  id: string;
  label: string;
  children?: ChildItem[];
}

const sections: SectionItem[] = [
  { id: 'top', label: 'Overview' },
  { id: 'getting-started', label: 'Getting Started' },
  {
    id: 'core',
    label: 'Core API',
    children: [
      { id: 'core-provider', label: 'Provider Interface' },
      { id: 'core-client', label: 'Client' },
      { id: 'core-message', label: 'Message Builder' },
      { id: 'core-result', label: 'Result & Status' },
    ],
  },
  {
    id: 'providers',
    label: 'Providers',
    children: [
      { id: 'providers-twilio', label: 'Twilio' },
      { id: 'providers-sns', label: 'AWS SNS' },
      { id: 'providers-vonage', label: 'Vonage' },
      { id: 'providers-msg91', label: 'MSG91' },
    ],
  },
  {
    id: 'bulk',
    label: 'Bulk Messaging',
    children: [
      { id: 'bulk-batch', label: 'Batch' },
      { id: 'bulk-many', label: 'SendToMany' },
    ],
  },
  {
    id: 'multi-provider',
    label: 'Multi-Provider',
    children: [
      { id: 'multi-fallback', label: 'Fallback' },
      { id: 'multi-roundrobin', label: 'Round-Robin' },
    ],
  },
  {
    id: 'webhooks',
    label: 'Webhooks',
    children: [
      { id: 'webhooks-status', label: 'Get Status' },
      { id: 'webhooks-twilio', label: 'Twilio Webhook' },
      { id: 'webhooks-vonage', label: 'Vonage Webhook' },
      { id: 'webhooks-msg91', label: 'MSG91 Webhook' },
    ],
  },
  {
    id: 'otp',
    label: 'OTP',
    children: [
      { id: 'otp-send', label: 'Send OTP' },
      { id: 'otp-verify', label: 'Verify & Resend' },
    ],
  },
  {
    id: 'helpers',
    label: 'Helpers',
    children: [
      { id: 'helpers-validate', label: 'Phone Validation' },
      { id: 'helpers-segments', label: 'SMS Segments' },
      { id: 'helpers-templates', label: 'Message Templates' },
    ],
  },
  {
    id: 'testing',
    label: 'Testing',
    children: [
      { id: 'testing-mock', label: 'Mock Provider' },
      { id: 'testing-assertions', label: 'Assertions & Errors' },
    ],
  },
  {
    id: 'errors',
    label: 'Errors',
    children: [
      { id: 'errors-sentinels', label: 'Sentinel Errors' },
      { id: 'errors-statuses', label: 'Status Values' },
    ],
  },
  { id: 'examples', label: 'Examples' },
];

function updateHash(id: string) {
  const url = new URL(window.location.href);
  if (id === 'top') {
    url.hash = '';
  } else {
    url.hash = id;
  }
  if (window.location.hash !== url.hash) {
    history.replaceState(null, '', url.toString());
  }
}

export default function Sidebar({ open, onClose }: SidebarProps) {
  const allIds = useMemo(
    () =>
      sections.flatMap((s) =>
        s.children ? [s.id, ...s.children.map((c) => c.id)] : [s.id],
      ),
    [],
  );

  const parentMap = useMemo(() => {
    const map = new Map<string, string>();
    for (const s of sections) {
      if (s.children) {
        for (const c of s.children) {
          map.set(c.id, s.id);
        }
      }
    }
    return map;
  }, []);

  const [active, setActive] = useState(() => {
    const hash = window.location.hash.slice(1);
    return hash && document.getElementById(hash) ? hash : 'top';
  });
  const [expanded, setExpanded] = useState<string | null>(() => {
    const hash = window.location.hash.slice(1);
    const parent = parentMap.get(hash);
    if (parent) return parent;
    const section = sections.find((s) => s.id === hash);
    return section?.children ? hash : null;
  });

  const visibleSet = useRef(new Set<string>());
  const isScrollingTo = useRef<string | null>(null);
  const scrollTimer = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    const hash = window.location.hash.slice(1);
    if (hash) {
      setTimeout(() => {
        document.getElementById(hash)?.scrollIntoView({ behavior: 'smooth' });
      }, 100);
    }
  }, []);

  const setActiveAndHash = (id: string) => {
    setActive(id);
    updateHash(id);

    const parent = parentMap.get(id);
    if (parent) {
      setExpanded(parent);
    } else {
      const section = sections.find((s) => s.id === id);
      if (section?.children) {
        setExpanded(id);
      }
    }
  };

  useEffect(() => {
    const updateActive = () => {
      if (isScrollingTo.current) return;
      if (visibleSet.current.size === 0) return;

      const navbarOffset = 80;
      let bestId: string | null = null;
      let bestDistance = Infinity;

      for (const id of visibleSet.current) {
        const el = document.getElementById(id);
        if (!el) continue;
        const rect = el.getBoundingClientRect();
        const distance = Math.abs(rect.top - navbarOffset);
        if (distance < bestDistance) {
          bestDistance = distance;
          bestId = id;
        }
      }

      if (bestId) {
        setActiveAndHash(bestId);
      }
    };

    const observer = new IntersectionObserver(
      (entries) => {
        for (const entry of entries) {
          if (entry.isIntersecting) {
            visibleSet.current.add(entry.target.id);
          } else {
            visibleSet.current.delete(entry.target.id);
          }
        }
        updateActive();
      },
      { rootMargin: '-80px 0px -40% 0px', threshold: 0.1 }
    );

    allIds.forEach((id) => {
      const el = document.getElementById(id);
      if (el) observer.observe(el);
    });

    const onScroll = () => {
      if (!isScrollingTo.current) return;
      if (scrollTimer.current) clearTimeout(scrollTimer.current);
      scrollTimer.current = setTimeout(() => {
        isScrollingTo.current = null;
        updateActive();
      }, 150);
    };

    window.addEventListener('scroll', onScroll, { passive: true });

    return () => {
      observer.disconnect();
      window.removeEventListener('scroll', onScroll);
      if (scrollTimer.current) clearTimeout(scrollTimer.current);
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [allIds]);

  const handleClick = (id: string) => {
    isScrollingTo.current = id;
    setActiveAndHash(id);
    document.getElementById(id)?.scrollIntoView({ behavior: 'smooth' });
    onClose();
  };

  const toggleExpand = (id: string) => {
    setExpanded((prev) => (prev === id ? null : id));
  };

  const isActive = (id: string) => active === id;

  const isSectionActive = (section: SectionItem) =>
    active === section.id ||
    (section.children?.some((c) => c.id === active) ?? false);

  const sectionClass = (section: SectionItem) =>
    `flex items-center justify-between w-full px-3 py-1.5 rounded-md text-sm transition-colors cursor-pointer ${
      isSectionActive(section)
        ? 'text-primary font-medium'
        : 'text-text-muted hover:text-text hover:bg-bg-card'
    }`;

  const subItemClass = (id: string) =>
    `block w-full text-left pl-6 pr-3 py-1 rounded-md text-xs transition-colors cursor-pointer ${
      isActive(id)
        ? 'bg-primary/10 text-primary font-medium'
        : 'text-text-muted hover:text-text hover:bg-bg-card'
    }`;

  return (
    <>
      {open && (
        <div
          className="fixed inset-0 bg-black/50 z-30 md:hidden"
          onClick={onClose}
        />
      )}

      <aside
        className={`fixed top-16 left-0 bottom-0 w-64 bg-bg-sidebar border-r border-border overflow-y-auto z-40 transition-transform ${
          open ? 'translate-x-0' : '-translate-x-full'
        } md:translate-x-0`}
      >
        <nav className="p-4 space-y-0.5">
          {sections.map((section) =>
            section.children ? (
              <div key={section.id}>
                <button
                  onClick={() => {
                    if (expanded === section.id) {
                      toggleExpand(section.id);
                    } else {
                      handleClick(section.id);
                    }
                  }}
                  className={sectionClass(section)}
                >
                  <span>{section.label}</span>
                  <svg
                    className={`w-3.5 h-3.5 transition-transform ${
                      expanded === section.id ? 'rotate-90' : ''
                    }`}
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M9 5l7 7-7 7"
                    />
                  </svg>
                </button>
                {expanded === section.id && (
                  <div className="mt-0.5 space-y-0.5">
                    {section.children.map((child) => (
                      <button
                        key={child.id}
                        onClick={() => handleClick(child.id)}
                        className={subItemClass(child.id)}
                      >
                        {child.label}
                      </button>
                    ))}
                  </div>
                )}
              </div>
            ) : (
              <button
                key={section.id}
                onClick={() => handleClick(section.id)}
                className={`block w-full text-left px-3 py-1.5 rounded-md text-sm transition-colors cursor-pointer ${
                  isActive(section.id)
                    ? 'bg-primary/10 text-primary font-medium'
                    : 'text-text-muted hover:text-text hover:bg-bg-card'
                }`}
              >
                {section.label}
              </button>
            )
          )}
        </nav>
      </aside>
    </>
  );
}
