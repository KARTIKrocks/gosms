import CodeBlock from '../components/CodeBlock';
import ModuleSection from '../components/ModuleSection';

export default function HelpersDocs() {
  return (
    <ModuleSection
      id="helpers"
      title="Helpers"
      description="Phone-number utilities, GSM 03.38 segment calculation, and pre-built message templates."
      importPath="github.com/KARTIKrocks/gosms"
      features={[
        'E.164 validation and normalization',
        'GSM 7-bit encoding detection and SMS segment counting',
        'Pre-built OTP, alert, and notification templates',
      ]}
    >
      {/* ── Phone Validation ── */}
      <h3 id="helpers-validate" className="text-lg font-semibold text-text-heading mt-8 mb-2">Phone Validation</h3>
      <CodeBlock code={`// Validate E.164 format
if gosms.ValidateE164("+15551234567") {
    log.Println("Valid E.164 number")
}

// Normalize a phone number (adds default country code, strips formatting)
normalized := gosms.NormalizePhone("555-123-4567", "+1")
// Returns: +15551234567`} />

      {/* ── SMS Segments ── */}
      <h3 id="helpers-segments" className="text-lg font-semibold text-text-heading mt-8 mb-2">SMS Segments</h3>
      <p className="text-text-muted mb-3">
        SMS messages are billed per segment. Single GSM 7-bit messages fit 160 chars; UCS-2 (Unicode) fits 70.
        Concatenated messages reserve 7 bytes for the user-data header, dropping the limits to 153 / 67 per segment.
      </p>
      <CodeBlock code={`// Check if message uses GSM 7-bit encoding
if gosms.IsGSMEncoding("Hello world") {
    log.Println("GSM encoding (160 char limit)")
}

// Calculate SMS segments
segments := gosms.CalculateSegments("Hello, this is a test message!")
log.Printf("Message will use %d segment(s)", segments)`} />

      {/* ── Templates ── */}
      <h3 id="helpers-templates" className="text-lg font-semibold text-text-heading mt-8 mb-2">Message Templates</h3>
      <p className="text-text-muted mb-3">
        Convenience builders for common SMS shapes. Each returns a <code className="text-accent">*Message</code> ready for <code className="text-accent">SendMessage</code>.
      </p>
      <CodeBlock code={`// OTP: "123456 is your MyApp verification code."
msg := gosms.OTPMessage("+15551234567", "123456", "MyApp")

// Alert: "[URGENT] Server is down!"
msg := gosms.AlertMessage("+15551234567", "URGENT", "Server is down!")

// Notification: "Order Update: Your order has shipped"
msg := gosms.NotificationMessage("+15551234567", "Order Update", "Your order has shipped")`} />
    </ModuleSection>
  );
}
