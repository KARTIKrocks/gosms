import CodeBlock from '../components/CodeBlock';
import ModuleSection from '../components/ModuleSection';

export default function WebhooksDocs() {
  return (
    <ModuleSection
      id="webhooks"
      title="Delivery Status & Webhooks"
      description="Track delivery state by polling GetStatus or by parsing inbound webhook callbacks from your provider."
      importPath="github.com/KARTIKrocks/gosms/{provider}"
      features={[
        'Client.GetStatus — poll the provider for the latest delivery state',
        'twilio.ParseWebhook — parse Twilio status callback POSTs',
        'vonage.ParseWebhook — parse Vonage delivery receipts',
        'msg91.ParseWebhook — parse MSG91 delivery callbacks',
      ]}
    >
      {/* ── Get Status ── */}
      <h3 id="webhooks-status" className="text-lg font-semibold text-text-heading mt-8 mb-2">Get Status</h3>
      <p className="text-text-muted mb-3">
        Poll for the latest delivery state. Status values progress
        <code className="text-accent"> Pending → Queued → Sent → Delivered</code> (or to a final failure state).
      </p>
      <CodeBlock code={`status, err := client.GetStatus(ctx, "message_id")
if err != nil {
    log.Fatal(err)
}

if status.Status.IsFinal() {
    if status.Status.IsSuccess() {
        log.Printf("Message delivered at %v", status.UpdatedAt)
    } else {
        log.Printf("Delivery failed: %s", status.ErrorMessage)
    }
}`} />

      {/* ── Twilio Webhook ── */}
      <h3 id="webhooks-twilio" className="text-lg font-semibold text-text-heading mt-8 mb-2">Twilio Webhook</h3>
      <p className="text-text-muted mb-3">
        Configure Twilio to POST status callbacks to your endpoint. <code className="text-accent">twilio.ParseWebhook</code> turns
        the form-encoded payload into a typed <code className="text-accent">StatusInfo</code>:
      </p>
      <CodeBlock code={`import "github.com/KARTIKrocks/gosms/twilio"

http.HandleFunc("/webhook/twilio", func(w http.ResponseWriter, r *http.Request) {
    status, err := twilio.ParseWebhook(r)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    log.Printf("Message %s status: %s", status.MessageID, status.Status)
    w.WriteHeader(http.StatusOK)
})`} />

      {/* ── Vonage Webhook ── */}
      <h3 id="webhooks-vonage" className="text-lg font-semibold text-text-heading mt-8 mb-2">Vonage Webhook</h3>
      <CodeBlock code={`import "github.com/KARTIKrocks/gosms/vonage"

http.HandleFunc("/webhook/vonage", func(w http.ResponseWriter, r *http.Request) {
    status, err := vonage.ParseWebhook(r)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    log.Printf("Message %s status: %s", status.MessageID, status.Status)
    w.WriteHeader(http.StatusOK)
})`} />

      {/* ── MSG91 Webhook ── */}
      <h3 id="webhooks-msg91" className="text-lg font-semibold text-text-heading mt-8 mb-2">MSG91 Webhook</h3>
      <p className="text-text-muted mb-3">
        MSG91 delivery is webhook-driven. Parse incoming callbacks with <code className="text-accent">msg91.ParseWebhook(r)</code>.
      </p>
      <CodeBlock code={`import "github.com/KARTIKrocks/gosms/msg91"

http.HandleFunc("/webhook/msg91", func(w http.ResponseWriter, r *http.Request) {
    status, err := msg91.ParseWebhook(r)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    log.Printf("Message %s status: %s", status.MessageID, status.Status)
    w.WriteHeader(http.StatusOK)
})`} />
    </ModuleSection>
  );
}
