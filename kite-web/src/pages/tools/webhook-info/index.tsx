import ToolLayout from "@/tools/common/components/ToolLayout";
import WebhookInfoTool from "@/tools/webhook-info/components/WebhookInfoTool";

export default function WebhookInfoPage() {
  return (
    <ToolLayout
      title="Webhook Info"
      description="Get information about Discord webhooks from the webhook URL. Paste
          your webhook URL below to get started!"
    >
      <WebhookInfoTool />
    </ToolLayout>
  );
}
