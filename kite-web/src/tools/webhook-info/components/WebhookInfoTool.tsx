import BaseInput from "@/tools/common/components/BaseInput";
import { parseWebhookUrl } from "@/tools/common/utils/webhook";
import { useQuery } from "@tanstack/react-query";
import { useState } from "react";
import WebhookInfo, {
  WebhookData,
} from "@/tools/webhook-info/components/WebhookInfo";

function useWebhookInfoQuery(url: string) {
  return useQuery({
    queryKey: ["webhook", url],
    queryFn: () => {
      if (!parseWebhookUrl(url)) {
        return {
          success: false,
          error: "Invalid Discord webhook URL",
        } as WebhookResponse;
      }

      return fetch(url).then(async (res) => {
        if (res.status != 200) {
          return {
            success: false,
            error: "Failed to fetch webhook",
          } as WebhookResponse;
        }

        return {
          success: true,
          ...(await res.json()),
        } as WebhookResponse;
      });
    },
    enabled: !!url,
  });
}

type WebhookResponse =
  | { success: false; error: string }
  | ({
      success: true;
    } & WebhookData);

export default function WebhookInfoTool() {
  const [webhookUrl, setWebhookUrl] = useState("");

  const { data } = useWebhookInfoQuery(webhookUrl);

  return (
    <div>
      <div className="mb-5">
        <BaseInput
          label=""
          type="text"
          placeholder="https://discord.com/api/webhooks/..."
          value={webhookUrl}
          onChange={setWebhookUrl}
          error={!data?.success ? data?.error : undefined}
        />
      </div>
      {data?.success && <WebhookInfo data={data} />}
    </div>
  );
}
