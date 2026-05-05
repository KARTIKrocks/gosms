import CodeBlock from '../components/CodeBlock';
import ModuleSection from '../components/ModuleSection';

export default function OTPDocs() {
  return (
    <ModuleSection
      id="otp"
      title="OTP (MSG91)"
      description="MSG91 implements the optional OTPProvider interface for the full one-time-password flow."
      importPath="github.com/KARTIKrocks/gosms/msg91"
      features={[
        'Server-side OTP generation — pass an empty OTP and MSG91 generates one',
        'Verify OTPs with a single VerifyOTP call',
        'Resend over text or voice channel',
        'Detect capability via type assertion on the Provider',
      ]}
    >
      {/* ── Send ── */}
      <h3 id="otp-send" className="text-lg font-semibold text-text-heading mt-8 mb-2">Send OTP</h3>
      <p className="text-text-muted mb-3">
        Callers holding a <code className="text-accent">gosms.Provider</code> can detect OTP support with a type assertion.
        When <code className="text-accent">OTPRequest.OTP</code> is empty, MSG91 generates the code server-side.
      </p>
      <CodeBlock code={`if otp, ok := provider.(gosms.OTPProvider); ok {
    _, err := otp.SendOTP(ctx, &gosms.OTPRequest{
        Phone:  "+919876543210",
        Length: 6,
        Expiry: 5 * time.Minute,
    })
    if err != nil {
        log.Fatal(err)
    }
}`} />

      {/* ── Verify & Resend ── */}
      <h3 id="otp-verify" className="text-lg font-semibold text-text-heading mt-8 mb-2">Verify & Resend</h3>
      <CodeBlock code={`if otp, ok := provider.(gosms.OTPProvider); ok {
    vr, err := otp.VerifyOTP(ctx, "+919876543210", "123456")
    if err == nil && vr.Verified {
        // OTP matched
    }

    // Resend via "text" or "voice"
    _ = otp.ResendOTP(ctx, "+919876543210", "voice")
}`} />
    </ModuleSection>
  );
}
