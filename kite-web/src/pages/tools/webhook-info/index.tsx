import ToolLayout from "@/tools/common/components/ToolLayout";
import WebhookInfoTool from "@/tools/webhook-info/components/WebhookInfoTool";

export default function WebhookInfoPage() {
  return (
    <ToolLayout
      toolId="webhook-info"
      title="Webhook Info"
      description="Get information about a Discord webhook from its URL, like its name, avatar, channel ID and server ID. Paste your webhook URL below to get started!"
    >
      <WebhookInfoTool />
    </ToolLayout>
  );
}
